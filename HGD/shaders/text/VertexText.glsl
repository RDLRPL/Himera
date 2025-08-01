#version 410
layout (location = 0) in vec4 vertex;
out vec2 TexCoords;
uniform mat4 projection;

void main() {
    vec4 pos = vec4(vertex.x, vertex.y, 0.0, 1.0);
    gl_Position = projection * pos;
    TexCoords = vertex.zw;
}
