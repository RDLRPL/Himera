#version 410
out vec4 FragColor;
uniform vec3 fillColor;
void main() {
    FragColor = vec4(fillColor, 1.0);
}