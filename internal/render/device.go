// Package render draws with OpenGL 4.5 core (DSA, GLSL 450). One renderer
// per pass; states mirror the original D3D9 ones documented in
// docs/apariencia.md.
package render

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/clientlog"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/go-gl/glfw/v3.4/glfw"
)

var skinIndirectSupported bool
var skinMultidrawSupported bool
var mapIndirectSupported bool
var drawParametersSupported bool
var sceneStorageSupported bool
var skinCrowdSupported bool

func configureDrawPaths(bindless, drawParameters bool, vertexBlocks, fragmentBlocks int32) {
	drawParametersSupported = drawParameters
	mapIndirectSupported = vertexBlocks >= 1 && fragmentBlocks >= 1
	sceneStorageSupported = vertexBlocks >= 4                    // scene effects and cloth
	skinCrowdSupported = vertexBlocks >= 2                       // transforms and bone palettes
	skinMultidrawSupported = drawParameters && vertexBlocks >= 3 // per-draw materials
	skinIndirectSupported = bindless && drawParametersSupported && sceneStorageSupported && fragmentBlocks >= 1
}

// Init loads GL entry points; debug installs a KHR_debug logger.
func Init(debug bool) error {
	return InitWithOptions(InitOptions{Debug: debug})
}

// InitOptions controls diagnostics and optional OpenGL extensions.
type InitOptions struct {
	Debug                 bool
	DisableBindless       bool
	DisableDrawParameters bool
}

// InitWithOptions loads entry points and selects paths for the current context.
func InitWithOptions(options InitOptions) error {
	return initDevice(options, glfw.GetProcAddress)
}

func initDevice(options InitOptions, resolve func(string) unsafe.Pointer) error {
	var missingBindless []string
	err := gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer {
		bindless := false
		switch name {
		case "glGetTextureHandleARB", "glIsTextureHandleResidentARB",
			"glMakeTextureHandleResidentARB", "glMakeTextureHandleNonResidentARB":
			bindless = true
		}
		if bindless && options.DisableBindless {
			return nil
		}
		address := resolve(name)
		// A driver lookup can return a failure sentinel instead of NULL.
		switch uintptr(address) {
		case 0, 1, 2, 3, ^uintptr(0):
			if bindless {
				missingBindless = append(missingBindless, name)
			}
			return nil
		}
		return address
	})
	if err != nil {
		return fmt.Errorf("load OpenGL entry points: %w", err)
	}
	var major, minor int32
	gl.GetIntegerv(gl.MAJOR_VERSION, &major)
	gl.GetIntegerv(gl.MINOR_VERSION, &minor)
	log.Printf("OpenGL %s, %s (vendor=%q GLSL=%q)",
		gl.GoStr(gl.GetString(gl.VERSION)), gl.GoStr(gl.GetString(gl.RENDERER)),
		gl.GoStr(gl.GetString(gl.VENDOR)), gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	if major < 4 || (major == 4 && minor < 5) {
		return fmt.Errorf("OpenGL 4.5 is required, the driver offers %d.%d", major, minor)
	}
	var profile, flags int32
	gl.GetIntegerv(gl.CONTEXT_PROFILE_MASK, &profile)
	gl.GetIntegerv(gl.CONTEXT_FLAGS, &flags)
	clientlog.Debugf("OpenGL context profile_mask=%#x flags=%#x", profile, flags)
	gpuBytes, drawCalls = 0, 0
	gpuCapacity = 0
	bindless, drawParameters, s3tc, dxt1 := false, false, false, false
	var extensionCount int32
	gl.GetIntegerv(gl.NUM_EXTENSIONS, &extensionCount)
	for i := int32(0); i < extensionCount; i++ {
		switch gl.GoStr(gl.GetStringi(gl.EXTENSIONS, uint32(i))) {
		case "GL_ARB_bindless_texture":
			bindless = true
		case "GL_ARB_shader_draw_parameters":
			drawParameters = true
		case "GL_EXT_texture_compression_s3tc":
			s3tc = true
		case "GL_EXT_texture_compression_dxt1":
			dxt1 = true
		case "GL_NVX_gpu_memory_info":
			var kib int32
			gl.GetIntegerv(0x9047, &kib) // GPU_MEMORY_INFO_DEDICATED_VIDMEM_NVX
			if kib > 0 {
				gpuCapacity = uint64(kib) * 1024
			}
		}
	}
	var vertexBlocks, fragmentBlocks int32
	gl.GetIntegerv(gl.MAX_VERTEX_SHADER_STORAGE_BLOCKS, &vertexBlocks)
	gl.GetIntegerv(gl.MAX_FRAGMENT_SHADER_STORAGE_BLOCKS, &fragmentBlocks)
	if bindless && len(missingBindless) > 0 {
		log.Printf("OpenGL bindless unavailable: missing commands %s", strings.Join(missingBindless, ", "))
	}
	configureDrawPaths(bindless && !options.DisableBindless && len(missingBindless) == 0, drawParameters && !options.DisableDrawParameters, vertexBlocks, fragmentBlocks)
	configureTextureFormats(s3tc, dxt1)
	log.Printf("OpenGL paths: bindless_indirect=%t bindless_advertised=%t bindless_disabled=%t draw_parameters=%t draw_parameters_advertised=%t draw_parameters_disabled=%t terrain_multidraw=%t scene_storage=%t skin_instancing=%t vertex_ssbo_limit=%d fragment_ssbo_limit=%d compressed_dds(DXT1,DXT3,DXT5)=%v",
		skinIndirectSupported, bindless, options.DisableBindless, drawParametersSupported, drawParameters, options.DisableDrawParameters, drawParametersSupported, sceneStorageSupported, skinCrowdSupported, vertexBlocks, fragmentBlocks, compressedDDS[1:])
	fallbackDiffuse = nil
	if options.Debug {
		gl.Enable(gl.DEBUG_OUTPUT)
		gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
		gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, _ unsafe.Pointer) {
			if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
				return
			}
			log.Printf("gl debug [%#x] source=%#x type=%#x id=%d: %s", severity, source, gltype, id, message)
		}, nil)
	}
	return nil
}

func compile(name, stage string, kind uint32, src string) (uint32, error) {
	clientlog.Debugf("OpenGL compiling %q (%s shader)", name, stage)
	s := gl.CreateShader(kind)
	cs, free := gl.Strs(src + "\x00")
	gl.ShaderSource(s, 1, cs, nil)
	free()
	gl.CompileShader(s)
	var ok int32
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &ok)
	if ok == gl.FALSE {
		var n int32
		gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &n)
		msg := strings.Repeat("\x00", int(n+1))
		gl.GetShaderInfoLog(s, n, nil, gl.Str(msg))
		gl.DeleteShader(s)
		return 0, fmt.Errorf("compile %q (%s shader): %s", name, stage, strings.TrimRight(msg, "\x00"))
	}
	return s, nil
}

// Program compiles and links a named vertex and fragment program.
func Program(name, vs, fs string) (uint32, error) {
	v, err := compile(name, "vertex", gl.VERTEX_SHADER, vs)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(v)
	f, err := compile(name, "fragment", gl.FRAGMENT_SHADER, fs)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(f)
	return link(name, v, f)
}

func computeProgram(name, src string) (uint32, error) {
	shader, err := compile(name, "compute", gl.COMPUTE_SHADER, src)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(shader)
	return link(name, shader)
}

func link(name string, shaders ...uint32) (uint32, error) {
	p := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(p, shader)
	}
	// Write the program name before entering the driver in case it terminates the process.
	clientlog.Debugf("OpenGL linking %q", name)
	gl.LinkProgram(p)
	var ok int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &ok)
	if ok == gl.FALSE {
		var n int32
		gl.GetProgramiv(p, gl.INFO_LOG_LENGTH, &n)
		msg := strings.Repeat("\x00", int(n+1))
		gl.GetProgramInfoLog(p, n, nil, gl.Str(msg))
		gl.DeleteProgram(p)
		return 0, fmt.Errorf("link %q: %s", name, strings.TrimRight(msg, "\x00"))
	}
	return p, nil
}
