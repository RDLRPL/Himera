package shaders

import (
	"os"
	"path/filepath"

	"github.com/RDLRPL/Himera/HGD/utils"
)

type Shaders struct {
	Vertex string
	Frag   string
}

func ReadShaders(ShadersPath string, VertexShaderFile string, FragShaderFile string) *Shaders {
	vertexShaderPath := filepath.Join(ShadersPath, VertexShaderFile)
	fragShaderPath := filepath.Join(ShadersPath, FragShaderFile)

	fragShader, err := os.ReadFile(fragShaderPath)
	utils.CheckErrors(err)
	vertexShader, err := os.ReadFile(vertexShaderPath)
	utils.CheckErrors(err)

	return &Shaders{
		Vertex: string(vertexShader) + "\x00",
		Frag:   string(fragShader) + "\x00",
	}
}
