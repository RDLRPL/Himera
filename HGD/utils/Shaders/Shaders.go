package shaders

import (
	"path/filepath"

	"github.com/RDLRPL/Himera/HGD/utils"
)

var TextShaders = ReadShaders(filepath.Join(utils.GetExecPath(), "../../../HGD/shaders/text"), "VertexText.glsl", "FragText.frag")
