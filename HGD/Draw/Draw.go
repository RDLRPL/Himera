package draw

import (
	"fmt"
	"strings"

	shaders "github.com/RDLRPL/Himera/HGD/utils/Shaders"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func CompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func CreateShaderProgram() (uint32, error) {
	vertexShader, err := CompileShader(shaders.TextShaders.Vertex, gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("failed to compile vertex shader: %v", err)
	}

	fragmentShader, err := CompileShader(shaders.TextShaders.Frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, fmt.Errorf("failed to compile fragment shader: %v", err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}
