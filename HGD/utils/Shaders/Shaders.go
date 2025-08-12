package shaders

import (
	"path/filepath"

	"github.com/RDLxxx/Himera/HGD/utils"
)

var TextShaders = ReadShaders(filepath.Join(utils.GetExecPath(), "../../../HGD/shaders/text"), "VertexText.glsl", "FragText.frag")
var RectShaders = ReadShaders(filepath.Join(utils.GetExecPath(), "../../../HGD/shaders/figures/rect"), "VertexRect.glsl", "FragRect.frag")
