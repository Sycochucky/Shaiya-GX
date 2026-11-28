package render

import (
	"unsafe"

	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

// The shared scene fragment shader reserves eight of the sixteen required
// fragment samplers for its other drawing modes.
const skinDrawTextureLimit = 8

const indexedDrawStride = int(unsafe.Sizeof(indexedDrawCommand{}))

type indexedDrawCommand struct {
	count, instances, first, baseVertex, baseInstance uint32
}

type skinDrawMaterial struct {
	handleLow, handleHigh uint32
	alphaRef, ambient     float32
}

type skinLocation struct {
	vertex, index int
}

type skinDrawRun struct {
	first     int
	indexType uint32
	twoSided  bool
	textures  drawTextures
}

// Capacity increases on every replacement until deletion, which also drops the
// VAO. Use capacity to detect a change: OpenGL can reuse a deleted buffer name
// while an unbound VAO still references the previous allocation.
type packedMeshBuffer struct {
	id             uint32
	used, capacity int
}

func (b *packedMeshBuffer) copy(source uint32, offset, size, limit int) {
	if need := offset + size; need > b.capacity {
		capacity := min(limit, max(need, 2*b.capacity, 256<<10))
		var buffer uint32
		gl.CreateBuffers(1, &buffer)
		gl.NamedBufferStorage(buffer, capacity, nil, 0)
		if b.id != 0 {
			gl.CopyNamedBufferSubData(b.id, buffer, 0, 0, b.used)
			gl.DeleteBuffers(1, &b.id)
		}
		removeGPUBytes(uint64(b.capacity))
		addGPUBytes(uint64(capacity))
		b.id, b.capacity = buffer, capacity
	}
	gl.CopyNamedBufferSubData(source, b.id, 0, offset, size)
	b.used = offset + size
}

func (b *packedMeshBuffer) delete() {
	if b.id != 0 {
		gl.DeleteBuffers(1, &b.id)
		removeGPUBytes(uint64(b.capacity))
		*b = packedMeshBuffer{}
	}
}

// SkinCrowd retains packed geometry for one scene. Meshes are immutable and
// their assets must outlive the cache. The ordinary VAOs remain available.
type SkinCrowd struct {
	vao, indirect                           uint32
	boundVertexCapacity, boundIndexCapacity int
	commandBytes                            uint64
	vertices, indices                       packedMeshBuffer
	meshes                                  map[*SkinMesh]skinLocation
	commands                                []indexedDrawCommand
	materials                               []skinDrawMaterial
	runs                                    []skinDrawRun
}

func (b *SkinCrowd) meshLocation(m *SkinMesh) (skinLocation, bool) {
	if at, ok := b.meshes[m]; ok {
		return at, true
	}
	const maxVertices, maxIndices = 48 << 20, 16 << 20
	vertexSize := m.vertexCount * skinVertexStride
	indexSize := int(m.bytes) - vertexSize
	vertex, index := b.vertices.used, (b.indices.used+3)&^3
	if vertexSize == 0 || indexSize <= 0 || vertex+vertexSize > maxVertices || index+indexSize > maxIndices {
		return skinLocation{}, false
	}
	b.vertices.copy(m.vbo, vertex, vertexSize, maxVertices)
	b.indices.copy(m.ebo, index, indexSize, maxIndices)
	at := skinLocation{vertex / skinVertexStride, index}
	b.meshes[m] = at
	return at, true
}

func (b *SkinCrowd) plan(groups []SkinGroup, ambient float32, bound bool) bool {
	if b.meshes == nil {
		b.meshes = make(map[*SkinMesh]skinLocation)
	}
	b.commands, b.materials, b.runs = b.commands[:0], b.materials[:0], b.runs[:0]
	for _, group := range groups {
		texture := diffuseTexture(group.Texture)
		at, ok := b.meshLocation(group.Mesh)
		if !ok {
			return false
		}
		indexSize := 2
		if group.Mesh.indexType == gl.UNSIGNED_INT {
			indexSize = 4
		}
		if n := len(b.runs); n == 0 || b.runs[n-1].indexType != group.Mesh.indexType || b.runs[n-1].twoSided != group.TwoSided {
			b.newRun(group)
		}
		var textureRef uint64
		if bound {
			slot, ok := b.runs[len(b.runs)-1].textures.slot(texture.ID, skinDrawTextureLimit)
			if !ok {
				b.newRun(group)
				slot, _ = b.runs[len(b.runs)-1].textures.slot(texture.ID, skinDrawTextureLimit)
			}
			textureRef = uint64(slot)
		} else {
			textureRef = texture.bindlessHandle(texture.repeat)
			if textureRef == 0 {
				return false
			}
		}
		b.commands = append(b.commands, indexedDrawCommand{uint32(group.Mesh.Count), uint32(group.Count), uint32(at.index / indexSize), uint32(at.vertex), uint32(group.First)})
		light := ambient
		if group.Unlit {
			light = 1
		}
		b.materials = append(b.materials, skinDrawMaterial{uint32(textureRef), uint32(textureRef >> 32), group.AlphaRef, light})
	}
	return len(b.commands) > len(b.runs)
}

func (b *SkinCrowd) newRun(group SkinGroup) {
	run := skinDrawRun{first: len(b.commands), indexType: group.Mesh.indexType, twoSided: group.TwoSided}
	if len(b.runs) < cap(b.runs) {
		i := len(b.runs)
		b.runs = b.runs[:i+1]
		run.textures = b.runs[i].textures[:0]
		b.runs[i] = run
	} else {
		b.runs = append(b.runs, run)
	}
}

// DrawCrowdIndirect submits consecutive groups with compatible raster state
// together. Indirect commands retain the same mesh and instance ordering.
func (k *Skinner) DrawCrowdIndirect(s *Scene3D, b *SkinCrowd, groups []SkinGroup, instances []SkinInstance, palettes *CrowdPalettes, ambient float32) (bool, int) {
	if k.indirect == nil || (s.Shadow && k.indirect.boundCrowdTextures) ||
		len(instances) == 0 || len(groups) < 2 || !b.plan(groups, ambient, k.indirect.boundCrowdTextures) {
		return false, 0
	}
	if b.vao == 0 {
		gl.CreateVertexArrays(1, &b.vao)
		gl.CreateBuffers(1, &b.indirect)
	}
	if b.boundVertexCapacity != b.vertices.capacity || b.boundIndexCapacity != b.indices.capacity {
		bindSkinMeshLayout(b.vao, b.vertices.id, b.indices.id)
		b.boundVertexCapacity, b.boundIndexCapacity = b.vertices.capacity, b.indices.capacity
	}
	commandBytes := len(b.commands) * indexedDrawStride
	gl.NamedBufferData(b.indirect, commandBytes, unsafe.Pointer(&b.commands[0]), gl.STREAM_DRAW)
	removeGPUBytes(b.commandBytes)
	b.commandBytes = uint64(commandBytes)
	addGPUBytes(b.commandBytes)

	k.indirect.Begin(s, ambient)
	k.uploadCrowdInstances(instances)
	written := palettes.bind()
	k.uploadCrowdBuffer(1, len(b.materials)*int(unsafe.Sizeof(b.materials[0])), unsafe.Pointer(&b.materials[0]))
	gl.BindVertexArray(b.vao)
	gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, b.indirect)
	gl.FrontFace(gl.CW)
	gl.CullFace(gl.BACK)
	for i, run := range b.runs {
		if run.twoSided {
			gl.Disable(gl.CULL_FACE)
		} else {
			gl.Enable(gl.CULL_FACE)
		}
		end := len(b.commands)
		if i+1 < len(b.runs) {
			end = b.runs[i+1].first
		}
		gl.Uniform1i(k.indirect.uCrowdDrawOffset, int32(run.first))
		run.textures.bind()
		gl.MultiDrawElementsIndirectWithOffset(gl.TRIANGLES, run.indexType, uintptr(run.first*indexedDrawStride), int32(end-run.first), 0)
		drawCalls++
	}
	gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, 0)
	gl.Disable(gl.CULL_FACE)
	gl.FrontFace(gl.CCW)
	return true, written
}

func (b *SkinCrowd) Delete() {
	if b.vao != 0 {
		gl.DeleteVertexArrays(1, &b.vao)
		gl.DeleteBuffers(1, &b.indirect)
	}
	b.vertices.delete()
	b.indices.delete()
	removeGPUBytes(b.commandBytes)
	*b = SkinCrowd{}
}
