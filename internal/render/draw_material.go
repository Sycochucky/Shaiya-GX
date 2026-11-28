package render

import "github.com/Spelunkern/Shaiya-GX/internal/render/gl"

// The indirect command's baseInstance selects a material shared by every
// instance of that command. The divisor exceeds the largest instance ID.
func bindDrawMaterial(vao, buffer uint32, stride int32) {
	const location, binding = 4, 1
	gl.VertexArrayVertexBuffer(vao, binding, buffer, 0, stride)
	gl.EnableVertexArrayAttrib(vao, location)
	gl.VertexArrayAttribIFormat(vao, location, 4, gl.UNSIGNED_INT, 0)
	gl.VertexArrayAttribBinding(vao, location, binding)
	gl.VertexArrayBindingDivisor(vao, binding, 1<<31)
}

// Each indirect command has one material for all its primitives. Constant
// sampler indices preserve ordinary implicit texture filtering in that draw.
const drawMaterialFragment = `
layout(binding=16) uniform sampler2D drawTextures[16];
vec4 sampleDrawTexture(uint slot, vec2 uv) {
 switch (slot) {
 case 0u: return texture(drawTextures[0], uv);
 case 1u: return texture(drawTextures[1], uv);
 case 2u: return texture(drawTextures[2], uv);
 case 3u: return texture(drawTextures[3], uv);
 case 4u: return texture(drawTextures[4], uv);
 case 5u: return texture(drawTextures[5], uv);
 case 6u: return texture(drawTextures[6], uv);
 case 7u: return texture(drawTextures[7], uv);
 case 8u: return texture(drawTextures[8], uv);
 case 9u: return texture(drawTextures[9], uv);
 case 10u: return texture(drawTextures[10], uv);
 case 11u: return texture(drawTextures[11], uv);
 case 12u: return texture(drawTextures[12], uv);
 case 13u: return texture(drawTextures[13], uv);
 case 14u: return texture(drawTextures[14], uv);
 case 15u: return texture(drawTextures[15], uv);
 default: return vec4(0.0);
 }
}
`
