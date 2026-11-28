package render

import (
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

// Sprites batches textured quads in pixel coordinates (origin top-left),
// the space the original UI is laid out in. Blending is the D3D9 UI state
// SRCALPHA / INVSRCALPHA.
type Sprites struct {
	clipRect               [4]float32
	clipped                bool
	prog                   uint32
	vao, vbo               uint32
	uScreen, uMasked       int32
	uPixelScale            int32
	uCircle                int32
	uGrayscale             int32
	nameDepth              float32
	depth                  *Texture
	depthValid             bool
	depthReversed          bool
	depthReadValid         bool
	depthReadX, depthReadY int
	depthReadValue         float32
	depthFrame             uint64
	hoverDepth             hoverDepthReadback
	verts                  []spriteVertex
	tex                    *Texture
	cap                    int
}

type spriteVertex struct {
	x, y, u, v  float32
	color       [4]float32
	depth, flat float32
}

const spriteVS = `#version 450 core
layout(location=0) in vec2 pos;
layout(location=1) in vec2 uv;
layout(location=2) in vec4 color;
layout(location=3) in vec2 labelState;
uniform vec2 screen;
out vec2 vUV;
out vec4 vColor;
flat out vec2 vLabelState;
void main() {
	vUV = uv;
	vColor = color;
	vLabelState = labelState;
	gl_Position = vec4(pos.x / screen.x * 2.0 - 1.0, 1.0 - pos.y / screen.y * 2.0, 0.0, 1.0);
}`

const spriteFS = `#version 450 core
layout(binding=0) uniform sampler2D tex;
layout(binding=1) uniform sampler2D mask;
layout(binding=2) uniform sampler2D worldDepth;
uniform int masked;
uniform int grayscale;
uniform int reversedDepth;
uniform vec2 screen;
uniform float pixelScale; // framebuffer pixels per UI unit (below 1 on small windows)
uniform vec3 circle; // UI-unit centre and radius; zero radius disables clipping
in vec2 vUV;
in vec4 vColor;
flat in vec2 vLabelState;
out vec4 frag;
void main() {
	// Phoenix classic_overlay.gdshader, adapted from reverse Z to GL ZERO_TO_ONE.
	if (vLabelState.x >= 0.0) {
 float sampledDepth = texelFetch(worldDepth, ivec2(gl_FragCoord.xy), 0).r;
 if (reversedDepth != 0) sampledDepth = 1.0 - sampledDepth;
 if (sampledDepth + 0.000001 < vLabelState.x) discard;
 }
	vec2 uv = masked != 0 ? vec2(vUV.x, 1.0 - vUV.y) : vUV;
	if (vLabelState.y != 0.0) {
		frag = vColor;
	} else if (masked != 0) {
		// Phoenix shaders/retail_portrait.gdshader: integrate the pixel
		// footprint when reducing the 384px snapshot to the HUD rectangle.
		vec2 dx = dFdx(uv), dy = dFdy(uv);
		vec4 face = vec4(0.0);
		for (int y = 0; y < 4; y++) {
			for (int x = 0; x < 4; x++) {
				vec2 offset = (vec2(x,y) + 0.5) / 4.0 - 0.5;
				face += texture(tex, clamp(uv + offset.x*dx + offset.y*dy, vec2(0.0), vec2(1.0)));
			}
		}
		frag = face * (1.0/16.0) * vColor;
		if (masked == 1) frag.a *= texture(mask, vUV).a;
 else frag.a *= 1.0-smoothstep(0.88,0.96,length((vUV-vec2(0.5))*2.0));
	} else {
		frag = texture(tex, uv) * vColor;
	}
 if (grayscale != 0) {
  float gray = dot(frag.rgb, vec3(0.299,0.587,0.114));
  frag.rgb = vec3(gray);
 }
 if (circle.z > 0.0) {
		// gl_FragCoord is in framebuffer pixels while sprites, screen and circle
		// are in UI units: convert before comparing.
		float k = pixelScale > 0.0 ? pixelScale : 1.0;
		vec2 unit = vec2(gl_FragCoord.x, screen.y * k - gl_FragCoord.y) / k;
		float radius = length(unit - circle.xy);
		frag.a *= 1.0 - smoothstep(circle.z - 1.0, circle.z, radius);
	}
}`

const spriteVertexBytes = int(unsafe.Sizeof(spriteVertex{}))

func NewSprites(maxQuads int) (*Sprites, error) {
	p, err := Program("UI sprites", spriteVS, spriteFS)
	if err != nil {
		return nil, err
	}
	s := &Sprites{prog: p, uScreen: gl.GetUniformLocation(p, gl.Str("screen\x00")), uPixelScale: gl.GetUniformLocation(p, gl.Str("pixelScale\x00")), cap: maxQuads, nameDepth: -1}
	s.uMasked = gl.GetUniformLocation(p, gl.Str("masked\x00"))
	s.uCircle = gl.GetUniformLocation(p, gl.Str("circle\x00"))
	s.uGrayscale = gl.GetUniformLocation(p, gl.Str("grayscale\x00"))
	gl.CreateBuffers(1, &s.vbo)
	gl.NamedBufferStorage(s.vbo, maxQuads*6*spriteVertexBytes, nil, gl.DYNAMIC_STORAGE_BIT)
	addGPUBytes(uint64(maxQuads * 6 * spriteVertexBytes))
	gl.CreateVertexArrays(1, &s.vao)
	gl.VertexArrayVertexBuffer(s.vao, 0, s.vbo, 0, int32(spriteVertexBytes))
	var vertex spriteVertex
	offsets := [4]uintptr{unsafe.Offsetof(vertex.x), unsafe.Offsetof(vertex.u), unsafe.Offsetof(vertex.color), unsafe.Offsetof(vertex.depth)}
	for i, n := range []int32{2, 2, 4, 2} {
		gl.EnableVertexArrayAttrib(s.vao, uint32(i))
		gl.VertexArrayAttribFormat(s.vao, uint32(i), n, gl.FLOAT, false, uint32(offsets[i]))
		gl.VertexArrayAttribBinding(s.vao, uint32(i), 0)
	}
	return s, nil
}

func (s *Sprites) Delete() {
	if s == nil || s.prog == 0 {
		return
	}
	s.ResetHoverDepth()
	s.depth.Delete()
	gl.DeleteVertexArrays(1, &s.vao)
	gl.DeleteBuffers(1, &s.vbo)
	gl.DeleteProgram(s.prog)
	removeGPUBytes(uint64(s.cap * 6 * spriteVertexBytes))
	*s = Sprites{}
}

// Begin starts a frame of w x h pixels.
func (s *Sprites) Begin(w, h int) {
	s.clipped = false
	gl.UseProgram(s.prog)
	gl.Uniform2f(s.uScreen, float32(w), float32(h))
	gl.Uniform1f(s.uPixelScale, float32(uiPixelScale))
	gl.BindVertexArray(s.vao)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.DEPTH_TEST)
}

// Draw queues a quad at (x,y) of size (w,h) with texture rect (u0,v0)-(u1,v1).
func (s *Sprites) Draw(t *Texture, x, y, w, h, u0, v0, u1, v1 float32, rgba [4]float32) {
	s.draw(t, x, y, w, h, u0, v0, u1, v1, rgba, 0)
}

// DrawFlat draws a solid quad without interrupting the active texture batch.
func (s *Sprites) DrawFlat(fallback *Texture, x, y, w, h float32, rgba [4]float32) {
	t := s.tex
	if t == nil {
		t = fallback
	}
	s.draw(t, x, y, w, h, 0, 0, 1, 1, rgba, 1)
}

func (s *Sprites) draw(t *Texture, x, y, w, h, u0, v0, u1, v1 float32, rgba [4]float32, flat float32) {
	if s.clipped {
		left, top := max(x, s.clipRect[0]), max(y, s.clipRect[1])
		right, bottom := min(x+w, s.clipRect[2]), min(y+h, s.clipRect[3])
		if right <= left || bottom <= top {
			return
		}
		du, dv := u1-u0, v1-v0
		u0, u1 = u0+du*(left-x)/w, u0+du*(right-x)/w
		v0, v1 = v0+dv*(top-y)/h, v0+dv*(bottom-y)/h
		x, y, w, h = left, top, right-left, bottom-top
	}
	s.reserve(t, 6)
	v := func(px, py, u, vv float32) {
		s.verts = append(s.verts, spriteVertex{px, py, u, vv, rgba, s.nameDepth, flat})
	}
	v(x, y, u0, v0)
	v(x+w, y, u1, v0)
	v(x+w, y+h, u1, v1)
	v(x, y, u0, v0)
	v(x+w, y+h, u1, v1)
	v(x, y+h, u0, v1)
}

func (s *Sprites) reserve(t *Texture, vertices int) {
	if s.tex != t || len(s.verts)+vertices > s.cap*6 {
		s.Flush()
		s.tex = t
	}
}

// DrawImage draws a whole texture at its pixel size.
func (s *Sprites) DrawImage(t *Texture, x, y float32) {
	s.Draw(t, x, y, float32(t.W), float32(t.H), 0, 0, 1, 1, [4]float32{1, 1, 1, 1})
}

// DrawMasked draws a framebuffer snapshot through the original UI alpha mask.
func (s *Sprites) DrawMasked(t, mask *Texture, x, y float32) {
	s.Flush()
	gl.Uniform1i(s.uMasked, 1)
	gl.BindTextureUnit(1, mask.ID)
	s.Draw(t, x, y, float32(mask.W), float32(mask.H), 0, 0, 1, 1, [4]float32{1, 1, 1, 1})
	s.Flush()
	gl.Uniform1i(s.uMasked, 0)
}

// DrawPortrait uses Phoenix's final 55x53 ellipse; the offscreen snapshot
// remains square. The original 64px mask includes unrelated rim padding.
func (s *Sprites) DrawPortrait(t *Texture, x, y, w, h float32) {
	s.Flush()
	gl.Uniform1i(s.uMasked, 2)
	s.Draw(t, x, y, w, h, 0, 0, 1, 1, [4]float32{1, 1, 1, 1})
	s.Flush()
	gl.Uniform1i(s.uMasked, 0)
}

func (s *Sprites) Flush() {
	if len(s.verts) == 0 || s.tex == nil {
		s.verts = s.verts[:0]
		return
	}
	gl.NamedBufferSubData(s.vbo, 0, len(s.verts)*spriteVertexBytes, unsafe.Pointer(&s.verts[0]))
	gl.BindTextureUnit(0, s.tex.ID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(s.verts)))
	drawCalls++
	s.verts = s.verts[:0]
}

// End flushes the remaining quads.
func (s *Sprites) End() { s.Flush() }

// DrawGrayscale scopes desaturation to this quad and restores color batching.
func (s *Sprites) DrawGrayscale(t *Texture, x, y, w, h, u0, v0, u1, v1 float32, rgba [4]float32) {
	s.Flush()
	gl.ProgramUniform1i(s.prog, s.uGrayscale, 1)
	s.Draw(t, x, y, w, h, u0, v0, u1, v1, rgba)
	s.Flush()
	gl.ProgramUniform1i(s.prog, s.uGrayscale, 0)
}
