package himera

import (
	web "github.com/RDLxxx/Himera/HDS/core/web/html"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		needsRedraw := false

		if core.Browse.InputBoxFocused {
			switch key {
			case glfw.KeyEnter:
				core.Browse.Link = core.Browse.InputText
				UpdateContent(core.Browse.Link, core.Browse.Ua)
				core.Browse.InputBoxFocused = false
				needsRedraw = true
			case glfw.KeyEscape:
				core.Browse.InputBoxFocused = false
				needsRedraw = true
			case glfw.KeyBackspace:
				if core.Browse.CursorPosition > 0 {
					core.Browse.InputText = core.Browse.InputText[:core.Browse.CursorPosition-1] + core.Browse.InputText[core.Browse.CursorPosition:]
					core.Browse.CursorPosition--
					needsRedraw = true
				}
			case glfw.KeyDelete:
				if core.Browse.CursorPosition < len(core.Browse.InputText) {
					core.Browse.InputText = core.Browse.InputText[:core.Browse.CursorPosition] + core.Browse.InputText[core.Browse.CursorPosition+1:]
					needsRedraw = true
				}
			case glfw.KeyLeft:
				if core.Browse.CursorPosition > 0 {
					core.Browse.CursorPosition--
					needsRedraw = true
				}
			case glfw.KeyRight:
				if core.Browse.CursorPosition < len(core.Browse.InputText) {
					core.Browse.CursorPosition++
					needsRedraw = true
				}
			case glfw.KeyHome:
				core.Browse.CursorPosition = 0
				needsRedraw = true
			case glfw.KeyEnd:
				core.Browse.CursorPosition = len(core.Browse.InputText)
				needsRedraw = true
			case glfw.KeyA:
				if mods&glfw.ModControl != 0 {
					core.Browse.CursorPosition = len(core.Browse.InputText)
					needsRedraw = true
				}
			}
		}

		switch key {
		case glfw.KeyF5:
			UpdateContent(core.Browse.Link, core.Browse.Ua)
			needsRedraw = true
		case glfw.KeyF11:
			ToggleFullscreen(window)
			needsRedraw = true
		case glfw.KeyL:
			if mods&glfw.ModControl != 0 {
				core.Browse.InputBoxFocused = true
				core.Browse.CursorPosition = len(core.Browse.InputText)
				needsRedraw = true
			}
		case glfw.KeyEqual, glfw.KeyKPAdd:
			if mods&glfw.ModControl != 0 {
				AdjustZoom(0.1)
				needsRedraw = true
			}
		case glfw.KeyMinus, glfw.KeyKPSubtract:
			if mods&glfw.ModControl != 0 {
				AdjustZoom(-0.1)
				needsRedraw = true
			}
		case glfw.Key0, glfw.KeyKP0:
			if mods&glfw.ModControl != 0 {
				core.Browse.Zoom = 1.0
				core.Browse.ScrollOffset = 0
				if core.Browse.HtmlRenderer != nil {
					ctx := &web.RenderContext{
						Width:  float32(core.Browse.CurrentWidth),
						Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 10.0,
						Zoom:   core.Browse.Zoom,
					}
					core.Browse.ContentHeight = core.Browse.HtmlRenderer.CalculateContentHeight(ctx)
				}
				needsRedraw = true
			}
		}

		if !core.Browse.InputBoxFocused {
			switch key {
			case glfw.KeyHome:
				core.Browse.ScrollOffset = 0
				needsRedraw = true
			case glfw.KeyEnd:
				UpdateScrollLimits()
				availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
				core.Browse.ScrollOffset = -(core.Browse.ContentHeight - availableHeight*0.9)
				if core.Browse.ScrollOffset > 0 {
					core.Browse.ScrollOffset = 0
				}
				needsRedraw = true
			case glfw.KeyPageUp:
				core.Browse.ScrollOffset += float32(core.Browse.CurrentHeight) * 0.8
				UpdateScrollLimits()
				needsRedraw = true
			case glfw.KeyPageDown:
				core.Browse.ScrollOffset -= float32(core.Browse.CurrentHeight) * 0.8
				UpdateScrollLimits()
				needsRedraw = true
			case glfw.KeyUp:
				core.Browse.ScrollOffset += 50.0
				UpdateScrollLimits()
				needsRedraw = true
			case glfw.KeyDown:
				core.Browse.ScrollOffset -= 50.0
				UpdateScrollLimits()
				needsRedraw = true
			}
		}

		if needsRedraw {
			MarkNeedsRedraw()
		}
	}
}
