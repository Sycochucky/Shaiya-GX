package render

import (
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

// SkinInstance matches CrowdInstance's std430 layout.
type SkinInstance struct {
	World, View, Projection M4
	Palette                 [4]int32
}

type SkinGroup struct {
	Mesh     *SkinMesh
	Texture  *Texture
	AlphaRef float32
	Unlit    bool
	TwoSided bool
	First    int
	Count    int
}

// CrowdPalettes owns the append-only bone buffer used by the scene's camera
// passes. Reset starts a new store before subsequent draw calls use it.
type CrowdPalettes struct {
	Matrices []M4
	buffer   uint32
	capacity int
	uploaded int
}

func (p *CrowdPalettes) Reset() {
	p.Matrices = p.Matrices[:0]
	p.uploaded = 0
}

func (p *CrowdPalettes) bind() int {
	size := len(p.Matrices) * 64
	if size == 0 {
		return 0
	}
	if p.buffer == 0 {
		gl.CreateBuffers(1, &p.buffer)
	}
	if p.uploaded == 0 || size > p.capacity {
		capacity := p.capacity
		if size > capacity {
			capacity = max(size, min(64<<20, 2*capacity))
		}
		gl.NamedBufferData(p.buffer, capacity, nil, gl.STREAM_DRAW)
		removeGPUBytes(uint64(p.capacity))
		addGPUBytes(uint64(capacity))
		p.capacity, p.uploaded = capacity, 0
	}
	written := size - p.uploaded
	if written > 0 {
		gl.NamedBufferSubData(p.buffer, p.uploaded, written, unsafe.Pointer(&p.Matrices[p.uploaded/64]))
		p.uploaded = size
	}
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 5, p.buffer)
	return written
}

func (p *CrowdPalettes) Delete() {
	if p.buffer != 0 {
		gl.DeleteBuffers(1, &p.buffer)
		removeGPUBytes(uint64(p.capacity))
	}
	*p = CrowdPalettes{}
}

func (s *Scene3D) CrowdSkinInstance(world M4) SkinInstance {
	view := s.modelView(world)
	return SkinInstance{World: world, View: view, Projection: view.Mul(s.depthProjection(s.Proj))}
}

func (k *Skinner) uploadCrowdBuffer(i int, size int, data unsafe.Pointer) {
	if k.crowdBuffers[i] == 0 {
		gl.CreateBuffers(1, &k.crowdBuffers[i])
	}
	gl.NamedBufferData(k.crowdBuffers[i], size, data, gl.STREAM_DRAW)
	removeGPUBytes(k.crowdBytes[i])
	k.crowdBytes[i] = uint64(size)
	addGPUBytes(k.crowdBytes[i])
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, [...]uint32{4, 6}[i], k.crowdBuffers[i])
}

func (k *Skinner) uploadCrowdInstances(instances []SkinInstance) {
	k.uploadCrowdBuffer(0, len(instances)*int(unsafe.Sizeof(instances[0])), unsafe.Pointer(&instances[0]))
}

// DrawCrowd preserves independent skeletal palettes while sharing material and
// draw state. Callers queue only opaque or alpha-tested world characters.
func (k *Skinner) DrawCrowd(s *Scene3D, groups []SkinGroup, instances []SkinInstance, palettes *CrowdPalettes, ambient float32) int {
	if len(instances) == 0 {
		return 0
	}
	k.Begin(s, ambient)
	k.uploadCrowdInstances(instances)
	written := palettes.bind()
	gl.FrontFace(gl.CW)
	gl.CullFace(gl.BACK)
	for _, group := range groups {
		gl.Uniform1i(k.uCrowdBase, int32(group.First+1))
		gl.Uniform1f(k.uRef, group.AlphaRef)
		light := ambient
		if group.Unlit {
			light = 1
		}
		k.SetAmbient(light)
		if group.TwoSided {
			gl.Disable(gl.CULL_FACE)
		} else {
			gl.Enable(gl.CULL_FACE)
		}
		bindDiffuse(group.Texture)
		gl.BindVertexArray(group.Mesh.vao)
		gl.DrawElementsInstanced(gl.TRIANGLES, group.Mesh.Count, group.Mesh.indexType, nil, int32(group.Count))
		drawCalls++
	}
	gl.Uniform1i(k.uCrowdBase, 0)
	gl.Disable(gl.CULL_FACE)
	gl.FrontFace(gl.CCW)
	return written
}

// RigidInstanceVisible applies the ordinary model's per-mesh occlusion test
// before a rigid attachment is added to an instanced draw.
func (s *Scene3D) RigidInstanceVisible(mesh *Mesh, world M4) bool {
	return !s.meshOccluded(mesh, world)
}
