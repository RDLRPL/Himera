package himera

import (
	"unicode"

	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func CharCallback(window *glfw.Window, char rune) {
	if core.Browse.InputBoxFocused && unicode.IsPrint(char) {
		core.Browse.InputText = core.Browse.InputText[:core.Browse.CursorPosition] + string(char) + core.Browse.InputText[core.Browse.CursorPosition:]
		core.Browse.CursorPosition++
		MarkNeedsRedraw()
	}
}

func MarkNeedsRedraw() {
	core.Browse.RState.NeedsRedraw = true
}
