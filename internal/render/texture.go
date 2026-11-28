package render

import (
	"slices"
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/clientlog"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"

	img "github.com/Spelunkern/Shaiya-GX/internal/data/image"
)

type Texture struct {
	Opaque        bool // proven full alpha coverage at every authored mip
	EmptyMask     bool // all terrain alpha samples are below the overlay threshold
	bytes         uint64
	ID            uint32
	W, H          int
	repeat        bool
	samplingViews [2]textureSamplingView
}

type textureSamplingView struct {
	id     uint32
	handle uint64
}

// Keep the ordinary scene's texture units intact. OpenGL 4.5 guarantees
// sixteen fragment samplers and enough combined units for this range.
const drawTextureUnit = 16
const drawTextureLimit = 16

type drawTextures []uint32

func (t *drawTextures) slot(id uint32, limit int) (int, bool) {
	if index := slices.Index(*t, id); index >= 0 {
		return index, true
	}
	index := len(*t)
	if index == limit {
		return 0, false
	}
	*t = append(*t, id)
	return index, true
}

func (t drawTextures) bind() {
	if len(t) != 0 {
		gl.BindTextures(drawTextureUnit, int32(len(t)), &t[0])
	}
}

var fallbackDiffuse *Texture
var compressedDDS [img.FormatDXT5 + 1]bool

func configureTextureFormats(s3tc, dxt1 bool) {
	for _, format := range [...]img.Format{img.FormatDXT1, img.FormatDXT3, img.FormatDXT5} {
		var supported int32
		if s3tc || (format == img.FormatDXT1 && dxt1) {
			gl.GetInternalformativ(gl.TEXTURE_2D, dxtFormat(format), gl.INTERNALFORMAT_SUPPORTED, 1, &supported)
		}
		compressedDDS[format] = supported != 0
	}
}

// A missing/untextured material must never inherit the preceding draw's
// sampler. D3D's untextured diffuse stage corresponds to multiplying by white.
func bindDiffuse(t *Texture) {
	gl.BindTextureUnit(0, diffuseTexture(t).ID)
}

func diffuseTexture(t *Texture) *Texture {
	if t == nil {
		if fallbackDiffuse == nil {
			fallbackDiffuse = Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: []byte{255, 255, 255, 255}}})
		}
		t = fallbackDiffuse
	}
	return t
}

// Views share the original texel storage and give each wrap mode its own
// texture state. Ordinary draws can still change wrapping on the original.
func (t *Texture) samplingView(repeat bool) *textureSamplingView {
	index := 0
	mode := int32(gl.CLAMP_TO_EDGE)
	if repeat {
		index = 1
		mode = gl.REPEAT
	}
	v := &t.samplingViews[index]
	if v.id != 0 {
		return v
	}
	if t.ID == 0 {
		return nil
	}
	var levels, format int32
	gl.GetTextureParameteriv(t.ID, gl.TEXTURE_IMMUTABLE_LEVELS, &levels)
	if levels < 1 {
		return nil
	}
	gl.GetTextureLevelParameteriv(t.ID, 0, gl.TEXTURE_INTERNAL_FORMAT, &format)
	gl.GenTextures(1, &v.id)
	gl.TextureView(v.id, gl.TEXTURE_2D, t.ID, uint32(format), 0, uint32(levels), 0, 1)
	if !gl.IsTexture(v.id) {
		gl.DeleteTextures(1, &v.id)
		v.id = 0
		return nil
	}
	for _, parameter := range [...]uint32{gl.TEXTURE_MIN_FILTER, gl.TEXTURE_MAG_FILTER, gl.TEXTURE_MAX_LEVEL} {
		var value int32
		gl.GetTextureParameteriv(t.ID, parameter, &value)
		gl.TextureParameteri(v.id, parameter, value)
	}
	gl.TextureParameteri(v.id, gl.TEXTURE_WRAP_S, mode)
	gl.TextureParameteri(v.id, gl.TEXTURE_WRAP_T, mode)
	return v
}

func (t *Texture) bindlessHandle(repeat bool) uint64 {
	v := t.samplingView(repeat)
	if v == nil {
		return 0
	}
	if v.handle != 0 {
		return v.handle
	}
	v.handle = gl.GetTextureHandleARB(v.id)
	if v.handle != 0 {
		gl.MakeTextureHandleResidentARB(v.handle)
		if gl.IsTextureHandleResidentARB(v.handle) {
			return v.handle
		}
	}
	if v.id != 0 {
		gl.DeleteTextures(1, &v.id)
	}
	*v = textureSamplingView{}
	return 0
}

// Upload creates an immutable texture (DSA). Supported DXT stays compressed.
func Upload(t img.Texture) *Texture { return uploadTexture(t, false) }

// UploadWorld requests the full mip chain like Game.exe 0x5aff00/0x5ac500.
// Existing authored levels remain byte-identical; only missing levels are built.
func UploadWorld(t img.Texture) *Texture { return uploadTexture(t, true) }

func fullMipCount(w, h int) int32 {
	n := int32(1)
	for w > 1 || h > 1 {
		w, h = max(w/2, 1), max(h/2, 1)
		n++
	}
	return n
}

func uploadTexture(t img.Texture, world bool) *Texture {
	var id uint32
	gl.CreateTextures(gl.TEXTURE_2D, 1, &id)
	tex := &Texture{ID: id}
	if world {
		tex.Opaque = img.Opaque(t)
		tex.EmptyMask = !tex.Opaque && img.TerrainMaskEmpty(t)
	}
	if t.DDS != nil {
		d := t.DDS
		tex.W, tex.H = d.W, d.H
		levels := int32(len(d.Levels))
		authored := levels
		if world {
			levels = fullMipCount(d.W, d.H)
		}
		block := 0
		compressed := compressedDDS[d.Format]
		if compressed {
			block = 16
			if d.Format == img.FormatDXT1 {
				block = 8
			}
		}
		tex.bytes = textureStorageBytes(d.W, d.H, int(levels), block)
		if compressed {
			gl.TextureStorage2D(id, levels, dxtFormat(d.Format), int32(d.W), int32(d.H))
		} else {
			gl.TextureStorage2D(id, levels, gl.RGBA8, int32(d.W), int32(d.H))
		}
		w, h := d.W, d.H
		for i, l := range d.Levels {
			if compressed {
				gl.CompressedTextureSubImage2D(id, int32(i), 0, 0, int32(w), int32(h), dxtFormat(d.Format), int32(len(l)), unsafe.Pointer(&l[0]))
			} else {
				pixels, err := d.LevelRGBA(i)
				if err != nil {
					clientlog.Warnf("OpenGL upload DDS %dx%d format %d mip %d: %v", d.W, d.H, d.Format, i, err)
					pixels = make([]byte, w*h*4)
				}
				gl.TextureSubImage2D(id, int32(i), 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[0]))
			}
			w, h = max(w/2, 1), max(h/2, 1)
		}
		gl.TextureParameteri(id, gl.TEXTURE_MAX_LEVEL, levels-1)
		if levels > authored {
			gl.TextureParameteri(id, gl.TEXTURE_BASE_LEVEL, authored-1)
			gl.GenerateTextureMipmap(id)
			gl.TextureParameteri(id, gl.TEXTURE_BASE_LEVEL, 0)
		}
		if levels > 1 {
			gl.TextureParameteri(id, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
		} else {
			gl.TextureParameteri(id, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		}
	} else {
		r := t.RGBA
		tex.W, tex.H = r.W, r.H
		levels := int32(1)
		if world {
			levels = fullMipCount(r.W, r.H)
		}
		tex.bytes = textureStorageBytes(r.W, r.H, int(levels), 0)
		gl.TextureStorage2D(id, levels, gl.RGBA8, int32(r.W), int32(r.H))
		gl.TextureSubImage2D(id, 0, 0, 0, int32(r.W), int32(r.H), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&r.Pix[0]))
		gl.TextureParameteri(id, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TextureParameteri(id, gl.TEXTURE_MAX_LEVEL, levels-1)
		if levels > 1 {
			gl.GenerateTextureMipmap(id)
		}
	}
	if world {
		gl.TextureParameteri(id, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	}
	gl.TextureParameteri(id, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TextureParameteri(id, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(id, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	addGPUBytes(tex.bytes)
	return tex
}

func dxtFormat(f img.Format) uint32 {
	switch f {
	case img.FormatDXT1:
		return gl.COMPRESSED_RGBA_S3TC_DXT1_EXT
	case img.FormatDXT3:
		return gl.COMPRESSED_RGBA_S3TC_DXT3_EXT
	default:
		return gl.COMPRESSED_RGBA_S3TC_DXT5_EXT
	}
}

func (t *Texture) Delete() {
	if t == nil {
		return
	}
	for i := range t.samplingViews {
		v := &t.samplingViews[i]
		if v.handle != 0 {
			gl.MakeTextureHandleNonResidentARB(v.handle)
		}
		if v.id != 0 {
			gl.DeleteTextures(1, &v.id)
			*v = textureSamplingView{}
		}
	}
	gl.DeleteTextures(1, &t.ID)
	removeGPUBytes(t.bytes)
	t.bytes = 0
	t.ID = 0
}

func textureStorageBytes(w, h, levels, block int) uint64 {
	var n uint64
	for i := 0; i < levels; i++ {
		if block == 0 {
			n += uint64(w * h * 4)
		} else {
			n += uint64(((w + 3) / 4) * ((h + 3) / 4) * block)
		}
		w, h = max(1, w/2), max(1, h/2)
	}
	return n
}
