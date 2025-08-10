package draw

import (
	"fmt"
	"strings"

	shaders "github.com/RDLRPL/Himera/HGD/utils/Shaders"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type CompiledShaders struct {
	TextShaderVertex uint32
	TextShaderFrag   uint32
	RectShaderVertex uint32
	RectShaderFrag   uint32
}

type ShadersPrograms struct {
	TextShaderProgram uint32
	RectShaderProgram uint32
}

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
		return 0, fmt.Errorf("failed ? %v: %v", source, log)
	}

	return shader, nil
}

func CompileShaders() (CompiledShaders, error) {
	tvs, err := CompileShader(shaders.TextShaders.Vertex, gl.VERTEX_SHADER)
	if err != nil {
		return CompiledShaders{}, fmt.Errorf("vertex ? %v", err)
	}
	tfs, err := CompileShader(shaders.TextShaders.Frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return CompiledShaders{}, fmt.Errorf("frag ? %v", err)
	}
	rsv, err := CompileShader(shaders.RectShaders.Vertex, gl.VERTEX_SHADER)
	if err != nil {
		return CompiledShaders{}, fmt.Errorf("vertex ? %v", err)
	}
	rsf, err := CompileShader(shaders.RectShaders.Frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return CompiledShaders{}, fmt.Errorf("frag ? %v", err)
	}
	return CompiledShaders{
		TextShaderVertex: tvs,
		TextShaderFrag:   tfs,
		RectShaderVertex: rsv,
		RectShaderFrag:   rsf,
	}, nil

}

func MakeShadersPrgs() (ShadersPrograms, error) {
	Shaders, err := CompileShaders()
	if err != nil {
		return ShadersPrograms{}, fmt.Errorf("shaders ? %v", err)
	}

	// Text
	Tp := gl.CreateProgram()
	gl.AttachShader(Tp, Shaders.TextShaderVertex)
	gl.AttachShader(Tp, Shaders.TextShaderFrag)
	gl.LinkProgram(Tp)
	var statusText int32
	gl.GetProgramiv(Tp, gl.LINK_STATUS, &statusText)
	if statusText == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(Tp, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(Tp, logLength, nil, gl.Str(log))
		return ShadersPrograms{}, fmt.Errorf("link program ? %v", log)
	}

	// Rect
	Rp := gl.CreateProgram()
	gl.AttachShader(Rp, Shaders.RectShaderVertex)
	gl.AttachShader(Rp, Shaders.RectShaderFrag)
	gl.LinkProgram(Rp)
	var status int32
	gl.GetProgramiv(Rp, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(Rp, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(Rp, logLength, nil, gl.Str(log))
		return ShadersPrograms{}, fmt.Errorf("link program ? %v", log)
	}

	// Optimization ♥
	// Text
	gl.DeleteShader(Shaders.TextShaderVertex)
	gl.DeleteShader(Shaders.TextShaderFrag)
	gl.DetachShader(Tp, Shaders.TextShaderVertex)
	gl.DetachShader(Tp, Shaders.TextShaderFrag)
	// Vertex
	gl.DeleteShader(Shaders.RectShaderVertex)
	gl.DeleteShader(Shaders.RectShaderFrag)
	gl.DetachShader(Rp, Shaders.RectShaderVertex)
	gl.DetachShader(Rp, Shaders.RectShaderFrag)

	return ShadersPrograms{
		TextShaderProgram: Tp,
		RectShaderProgram: Rp,
	}, nil
}
