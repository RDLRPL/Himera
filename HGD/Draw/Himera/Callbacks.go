package himera

import (
	"unicode"

	web "github.com/RDLxxx/Himera/HDS/core/html"
	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func CharCallback(window *glfw.Window, char rune) {
	if core.Browse.InputBoxFocused && unicode.IsPrint(char) {
		core.Browse.InputText = core.Browse.InputText[:core.Browse.CursorPosition] + string(char) + core.Browse.InputText[core.Browse.CursorPosition:]
		core.Browse.CursorPosition++
		MarkNeedsRedraw()
	}
}

func InitializeWindowState(window *glfw.Window) {
	core.Browse.IsMaximized = window.GetAttrib(glfw.Maximized) == glfw.True
	core.Browse.WindowedWidth, core.Browse.WindowedHeight = window.GetSize()
	core.Browse.WindowedX, core.Browse.WindowedY = window.GetPos()
}

func ScrollCallback(window *glfw.Window, xoff, yoff float64) {
	if !core.Browse.InputBoxFocused {
		if window.GetKey(glfw.KeyLeftControl) == glfw.Press ||
			window.GetKey(glfw.KeyRightControl) == glfw.Press {
			AdjustZoom(float32(yoff) * 0.1)
		} else {
			core.Browse.ScrollOffset += float32(yoff) * 25.0
			UpdateScrollLimits()
		}
		MarkNeedsRedraw()
	}
}

func MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && button == glfw.MouseButtonLeft {
		xpos, ypos := window.GetCursorPos()

		inputBoxY := float32(5.0)
		if float32(ypos) >= inputBoxY && float32(ypos) <= inputBoxY+core.Browse.InputBoxHeight &&
			float32(xpos) >= 10.0 && float32(xpos) <= float32(core.Browse.CurrentWidth)-10.0 {
			core.Browse.InputBoxFocused = true
			textWidth, _ := TextLIB.GetTextDimensions(core.Browse.InputText, 1.0)
			relativeX := float32(xpos) - 15.0
			if relativeX < 0 {
				core.Browse.CursorPosition = 0
			} else if relativeX > textWidth {
				core.Browse.CursorPosition = len(core.Browse.InputText)
			} else {
				core.Browse.CursorPosition = int(float32(len(core.Browse.InputText)) * (relativeX / textWidth))
				if core.Browse.CursorPosition < 0 {
					core.Browse.CursorPosition = 0
				}
				if core.Browse.CursorPosition > len(core.Browse.InputText) {
					core.Browse.CursorPosition = len(core.Browse.InputText)
				}
			}
		} else {
			core.Browse.InputBoxFocused = false
		}
		MarkNeedsRedraw()
	}
}

func WindowMaximizeCallback(window *glfw.Window, maximized bool) {
	if !core.Browse.IsFullscreen {
		core.Browse.IsMaximized = maximized
		MarkNeedsRedraw()
	}
}
func FramebufferSizeCallback(window *glfw.Window, width, height int) {
	core.Browse.CurrentWidth = width
	core.Browse.CurrentHeight = height
	gl.Viewport(0, 0, int32(width), int32(height))

	if core.Browse.HtmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(core.Browse.CurrentWidth),
			Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 10.0,
			Zoom:   core.Browse.Zoom,
		}
		core.Browse.ContentHeight = core.Browse.HtmlRenderer.CalculateContentHeight(ctx)
		UpdateScrollLimits()
	}

	MarkNeedsRedraw()
}
