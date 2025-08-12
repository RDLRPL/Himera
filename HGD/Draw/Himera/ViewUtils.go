package himera

import (
	web "github.com/RDLxxx/Himera/HDS/core/html"
	"github.com/RDLxxx/Himera/HGD/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func UpdateScrollLimits() {
	if core.Browse.HtmlRenderer == nil {
		return
	}

	availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Width:  float32(core.Browse.CurrentWidth) - 20.0*core.Browse.Zoom,
		Height: availableHeight,
		X:      10.0 * core.Browse.Zoom,
		Y:      core.Browse.InputBoxHeight + 15.0*core.Browse.Zoom,
		Zoom:   core.Browse.Zoom,
	}
	core.Browse.ContentHeight = core.Browse.HtmlRenderer.CalculateContentHeight(ctx)

	maxScrollOffset := float32(0.0)
	minScrollOffset := -(core.Browse.ContentHeight - availableHeight*0.9)

	if minScrollOffset > 0 {
		minScrollOffset = 0
	}

	if core.Browse.ScrollOffset > maxScrollOffset {
		core.Browse.ScrollOffset = maxScrollOffset
	}
	if core.Browse.ScrollOffset < minScrollOffset {
		core.Browse.ScrollOffset = minScrollOffset
	}
}

func AdjustZoom(delta float32) {
	newZoom := core.Browse.Zoom + delta
	if newZoom < 0.1 {
		newZoom = 0.1
	} else if newZoom > 5.0 {
		newZoom = 5.0
	}

	if newZoom != core.Browse.Zoom {
		core.Browse.Zoom = newZoom
		core.Browse.ScrollOffset = 0

		if core.Browse.HtmlRenderer != nil {
			ctx := &web.RenderContext{
				Width:  float32(core.Browse.CurrentWidth),
				Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0,
				Zoom:   core.Browse.Zoom,
			}
			core.Browse.ContentHeight = core.Browse.HtmlRenderer.CalculateContentHeight(ctx)
		}

		MarkNeedsRedraw()
	}
}

func MarkNeedsRedraw() {
	core.Browse.RState.NeedsRedraw = true
}

func ToggleFullscreen(window *glfw.Window) {
	if core.Browse.IsFullscreen {
		window.SetMonitor(nil, core.Browse.WindowedX, core.Browse.WindowedY, core.Browse.WindowedWidth, core.Browse.WindowedHeight, 0)
		core.Browse.IsFullscreen = false

		if core.Browse.WasMaximizedBeforeFullscreen {
			window.Maximize()
			core.Browse.IsMaximized = true
		} else {
			core.Browse.IsMaximized = false
		}
	} else {
		core.Browse.WasMaximizedBeforeFullscreen = core.Browse.IsMaximized
		core.Browse.WindowedX, core.Browse.WindowedY = window.GetPos()
		core.Browse.WindowedWidth, core.Browse.WindowedHeight = window.GetSize()

		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		core.Browse.IsFullscreen = true
		core.Browse.CurrentWidth = mode.Width
		core.Browse.CurrentHeight = mode.Height
	}

	gl.Viewport(0, 0, int32(core.Browse.CurrentWidth), int32(core.Browse.CurrentHeight))

	if core.Browse.HtmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(core.Browse.CurrentWidth),
			Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0,
			Zoom:   core.Browse.Zoom,
		}
		core.Browse.ContentHeight = core.Browse.HtmlRenderer.CalculateContentHeight(ctx)
		UpdateScrollLimits()
	}

	MarkNeedsRedraw()
}

func CheckNeedsRedraw() bool {
	if core.Browse.RState.NeedsRedraw ||
		core.Browse.RState.LastWidth != core.Browse.CurrentWidth ||
		core.Browse.RState.LastHeight != core.Browse.CurrentHeight ||
		core.Browse.RState.LastZoom != core.Browse.Zoom ||
		core.Browse.RState.LastScroll != core.Browse.ScrollOffset ||
		core.Browse.RState.LastInputText != core.Browse.InputText ||
		core.Browse.RState.LastFocused != core.Browse.InputBoxFocused ||
		core.Browse.RState.LastCursorPos != core.Browse.CursorPosition {

		core.Browse.RState.LastWidth = core.Browse.CurrentWidth
		core.Browse.RState.LastHeight = core.Browse.CurrentHeight
		core.Browse.RState.LastZoom = core.Browse.Zoom
		core.Browse.RState.LastScroll = core.Browse.ScrollOffset
		core.Browse.RState.LastInputText = core.Browse.InputText
		core.Browse.RState.LastFocused = core.Browse.InputBoxFocused
		core.Browse.RState.LastCursorPos = core.Browse.CursorPosition
		core.Browse.RState.NeedsRedraw = false

		return true
	}

	if core.Browse.InputBoxFocused {
		return true
	}

	return false
}

func UpdateProjection(program uint32) {
	projection := [16]float32{
		2.0 / float32(core.Browse.CurrentWidth), 0, 0, 0,
		0, -2.0 / float32(core.Browse.CurrentHeight), 0, 0,
		0, 0, -1, 0,
		-1, 1, 0, 1,
	}

	gl.UseProgram(program)
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])
}
