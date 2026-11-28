package render

import (
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

// Delete releases the scene's streaming buffer and shader on the GL thread.
func (s *Scene3D) Delete() {
	if s == nil || s.prog == 0 {
		return
	}
	if s.instanceBuffer != 0 {
		gl.DeleteBuffers(1, &s.instanceBuffer)
		removeGPUBytes(s.instanceBytes)
		s.instanceBuffer, s.instanceBytes = 0, 0
	}
	if s.effectStream != nil {
		s.effectStream.Delete()
	}
	s.effectPipeline.delete()
	for _, program := range [...]uint32{s.mapPipeline.draw, s.mapPipeline.bound, s.mapPipeline.compute, s.terrainOverlay.prog, s.terrainHalo.program} {
		if program != 0 {
			gl.DeleteProgram(program)
		}
	}
	s.terrainOverlay.indirect.delete()
	s.ReleaseWaterBackground()
	if s.terrainIndirect != 0 {
		gl.DeleteBuffers(1, &s.terrainIndirect)
		removeGPUBytes(s.terrainIndirectBytes)
	}
	gl.DeleteProgram(s.prog)
	s.prog = 0
}

// SetInstances uploads one model's placements once; all its material parts
// reuse them. Camera products are computed before local vertices, preserving
// the precision fix for animated and static models.
func (s *Scene3D) SetInstances(worlds []M4) {
	s.uploadInstances(worlds, nil)
}

func (s *Scene3D) uploadInstances(worlds []M4, colors [][4]float32) {
	s.uploadInstanceData(worlds, colors, nil)
}

func (s *Scene3D) uploadInstanceData(worlds []M4, colors [][4]float32, poses []EffectPose) int32 {
	s.instanceData = s.instanceData[:0]
	for _, world := range worlds {
		mv := s.modelView(world)
		mvp := mv.Mul(s.projection)
		s.instanceData = append(s.instanceData, world[:]...)
		s.instanceData = append(s.instanceData, mv[:]...)
		s.instanceData = append(s.instanceData, mvp[:]...)
	}
	s.instanceCount = len(worlds)
	if s.instanceCount == 0 {
		return 0
	}
	if s.instanceBuffer == 0 {
		gl.CreateBuffers(1, &s.instanceBuffer)
	}
	// The VAO reads the matrix prefix. The shader reads the optional vec4
	// colors at the end of the same upload through its storage binding.
	colorOffset := len(s.instanceData) / 4
	for _, color := range colors {
		s.instanceData = append(s.instanceData, color[:]...)
	}
	poseOffset := int32(len(s.instanceData) / 4)
	for _, pose := range poses {
		s.instanceData = append(s.instanceData, pose[:]...)
	}
	bytes := uint64(len(s.instanceData) * 4)
	gl.NamedBufferData(s.instanceBuffer, int(bytes), unsafe.Pointer(&s.instanceData[0]), gl.STREAM_DRAW)
	removeGPUBytes(s.instanceBytes)
	s.instanceBytes = bytes
	addGPUBytes(bytes)
	if len(colors) != 0 {
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 3, s.instanceBuffer)
		gl.Uniform1i(s.uInstanceColorOffset, int32(colorOffset))
	}
	return poseOffset
}
func (s *Scene3D) DrawInstanced(m *Mesh, t *Texture, alphaRef float32) {
	if s.instanceCount == 0 || m == nil || m.Count == 0 {
		return
	}
	gl.Uniform1i(s.uInstanced, 1)
	gl.Uniform1i(s.uMode, 0)
	gl.Uniform1i(s.uLit, 1)
	gl.Uniform1f(s.uRef, alphaRef)
	s.blend(false)
	bindDiffuse(t)
	s.drawInstances(m)
}

// DrawEquipmentInstanced uses the same raster state as DrawEquipment.
func (s *Scene3D) DrawEquipmentInstanced(m *Mesh, t *Texture, alphaRef float32) {
	beginModelFaces(true)
	s.DrawInstanced(m, t, alphaRef)
	endModelFaces(true)
}

func (s *Scene3D) drawInstances(m *Mesh) {
	s.bindInstances(m)
	gl.DrawElementsInstanced(gl.TRIANGLES, m.Count, m.indexType, nil, int32(s.instanceCount))
	drawCalls++
}

func (s *Scene3D) bindInstances(m *Mesh) {
	if m.instanceBuffer != s.instanceBuffer {
		gl.VertexArrayVertexBuffer(m.vao, 1, s.instanceBuffer, 0, 3*64)
		for i := uint32(0); i < 12; i++ {
			location := 4 + i
			gl.EnableVertexArrayAttrib(m.vao, location)
			gl.VertexArrayAttribFormat(m.vao, location, 4, gl.FLOAT, false, i*16)
			gl.VertexArrayAttribBinding(m.vao, location, 1)
		}
		gl.VertexArrayBindingDivisor(m.vao, 1, 1)
		m.instanceBuffer = s.instanceBuffer
	}
	gl.BindVertexArray(m.vao)
}
