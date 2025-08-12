package main

import (
	"encoding/binary"
	"log"
	"os"
	"runtime"
	"unicode"

	web "github.com/RDLxxx/Himera/HDS/core/html"
	h "github.com/RDLxxx/Himera/HDS/core/http"
	draw "github.com/RDLxxx/Himera/HGD/Draw"
	drawer "github.com/RDLxxx/Himera/HGD/Draw/Drawer"
	himera "github.com/RDLxxx/Himera/HGD/Draw/Himera"
	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/RDLxxx/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type RenderState struct {
	needsRedraw   bool
	lastWidth     int
	lastHeight    int
	lastZoom      float32
	lastScroll    float32
	lastInputText string
	lastFocused   bool
	lastCursorPos int
}

var renderState = &RenderState{needsRedraw: true}

var (
	zoom         float32 = 1.0
	scrollOffset float32 = 0.0
	isFullscreen         = false
	isMaximized          = false

	htmlRenderer  *web.HTMLRenderer
	contentHeight float32 = 0.0

	windowedX, windowedY, windowedWidth, windowedHeight int
	wasMaximizedBeforeFullscreen                        bool

	inputBoxFocused bool = false
)

func init() {
	runtime.LockOSThread()
}

func checkNeedsRedraw() bool {
	if renderState.needsRedraw ||
		renderState.lastWidth != core.Browse.CurrentWidth ||
		renderState.lastHeight != core.Browse.CurrentHeight ||
		renderState.lastZoom != zoom ||
		renderState.lastScroll != scrollOffset ||
		renderState.lastInputText != core.Browse.InputText ||
		renderState.lastFocused != inputBoxFocused ||
		renderState.lastCursorPos != core.Browse.CursorPosition {

		renderState.lastWidth = core.Browse.CurrentWidth
		renderState.lastHeight = core.Browse.CurrentHeight
		renderState.lastZoom = zoom
		renderState.lastScroll = scrollOffset
		renderState.lastInputText = core.Browse.InputText
		renderState.lastFocused = inputBoxFocused
		renderState.lastCursorPos = core.Browse.CursorPosition
		renderState.needsRedraw = false

		return true
	}

	if inputBoxFocused {
		return true
	}

	return false
}

func markNeedsRedraw() {
	renderState.needsRedraw = true
}

func updateProjection(program uint32) {
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

func toggleFullscreen(window *glfw.Window) {
	if isFullscreen {
		window.SetMonitor(nil, windowedX, windowedY, windowedWidth, windowedHeight, 0)
		isFullscreen = false

		if wasMaximizedBeforeFullscreen {
			window.Maximize()
			isMaximized = true
		} else {
			isMaximized = false
		}
	} else {
		wasMaximizedBeforeFullscreen = isMaximized
		windowedX, windowedY = window.GetPos()
		windowedWidth, windowedHeight = window.GetSize()

		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		isFullscreen = true
		core.Browse.CurrentWidth = mode.Width
		core.Browse.CurrentHeight = mode.Height
	}

	gl.Viewport(0, 0, int32(core.Browse.CurrentWidth), int32(core.Browse.CurrentHeight))

	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(core.Browse.CurrentWidth),
			Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0,
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		updateScrollLimits()
	}

	markNeedsRedraw()
}

func windowMaximizeCallback(window *glfw.Window, maximized bool) {
	if !isFullscreen {
		isMaximized = maximized
		markNeedsRedraw()
	}
}

func adjustZoom(delta float32) {
	newZoom := zoom + delta
	if newZoom < 0.1 {
		newZoom = 0.1
	} else if newZoom > 5.0 {
		newZoom = 5.0
	}

	if newZoom != zoom {
		zoom = newZoom
		scrollOffset = 0

		if htmlRenderer != nil {
			ctx := &web.RenderContext{
				Width:  float32(core.Browse.CurrentWidth),
				Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0,
				Zoom:   zoom,
			}
			contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		}

		markNeedsRedraw()
	}
}

func updateScrollLimits() {
	if htmlRenderer == nil {
		return
	}

	availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Width:  float32(core.Browse.CurrentWidth) - 20.0*zoom,
		Height: availableHeight,
		X:      10.0 * zoom,
		Y:      core.Browse.InputBoxHeight + 15.0*zoom,
		Zoom:   zoom,
	}
	contentHeight = htmlRenderer.CalculateContentHeight(ctx)

	maxScrollOffset := float32(0.0)
	minScrollOffset := -(contentHeight - availableHeight*0.9)

	if minScrollOffset > 0 {
		minScrollOffset = 0
	}

	if scrollOffset > maxScrollOffset {
		scrollOffset = maxScrollOffset
	}
	if scrollOffset < minScrollOffset {
		scrollOffset = minScrollOffset
	}
}

func framebufferSizeCallback(window *glfw.Window, width, height int) {
	core.Browse.CurrentWidth = width
	core.Browse.CurrentHeight = height
	gl.Viewport(0, 0, int32(width), int32(height))

	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(core.Browse.CurrentWidth),
			Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 10.0,
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		updateScrollLimits()
	}

	markNeedsRedraw()
}

func charCallback(window *glfw.Window, char rune) {
	if inputBoxFocused && unicode.IsPrint(char) {
		core.Browse.InputText = core.Browse.InputText[:core.Browse.CursorPosition] + string(char) + core.Browse.InputText[core.Browse.CursorPosition:]
		core.Browse.CursorPosition++
		markNeedsRedraw()
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		needsRedraw := false

		if inputBoxFocused {
			switch key {
			case glfw.KeyEnter:
				core.Browse.Link = core.Browse.InputText
				updateContent(core.Browse.Link, core.Browse.Ua)
				inputBoxFocused = false
				needsRedraw = true
			case glfw.KeyEscape:
				inputBoxFocused = false
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
			updateContent(core.Browse.Link, core.Browse.Ua)
			needsRedraw = true
		case glfw.KeyF11:
			toggleFullscreen(window)
			needsRedraw = true
		case glfw.KeyL:
			if mods&glfw.ModControl != 0 {
				inputBoxFocused = true
				core.Browse.CursorPosition = len(core.Browse.InputText)
				needsRedraw = true
			}
		case glfw.KeyEqual, glfw.KeyKPAdd:
			if mods&glfw.ModControl != 0 {
				adjustZoom(0.1)
				needsRedraw = true
			}
		case glfw.KeyMinus, glfw.KeyKPSubtract:
			if mods&glfw.ModControl != 0 {
				adjustZoom(-0.1)
				needsRedraw = true
			}
		case glfw.Key0, glfw.KeyKP0:
			if mods&glfw.ModControl != 0 {
				zoom = 1.0
				scrollOffset = 0
				if htmlRenderer != nil {
					ctx := &web.RenderContext{
						Width:  float32(core.Browse.CurrentWidth),
						Height: float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 10.0,
						Zoom:   zoom,
					}
					contentHeight = htmlRenderer.CalculateContentHeight(ctx)
				}
				needsRedraw = true
			}
		}

		if !inputBoxFocused {
			switch key {
			case glfw.KeyHome:
				scrollOffset = 0
				needsRedraw = true
			case glfw.KeyEnd:
				updateScrollLimits()
				availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
				scrollOffset = -(contentHeight - availableHeight*0.9)
				if scrollOffset > 0 {
					scrollOffset = 0
				}
				needsRedraw = true
			case glfw.KeyPageUp:
				scrollOffset += float32(core.Browse.CurrentHeight) * 0.8
				updateScrollLimits()
				needsRedraw = true
			case glfw.KeyPageDown:
				scrollOffset -= float32(core.Browse.CurrentHeight) * 0.8
				updateScrollLimits()
				needsRedraw = true
			case glfw.KeyUp:
				scrollOffset += 50.0
				updateScrollLimits()
				needsRedraw = true
			case glfw.KeyDown:
				scrollOffset -= 50.0
				updateScrollLimits()
				needsRedraw = true
			}
		}

		if needsRedraw {
			markNeedsRedraw()
		}
	}
}

func mouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && button == glfw.MouseButtonLeft {
		xpos, ypos := window.GetCursorPos()

		inputBoxY := float32(5.0)
		if float32(ypos) >= inputBoxY && float32(ypos) <= inputBoxY+core.Browse.InputBoxHeight &&
			float32(xpos) >= 10.0 && float32(xpos) <= float32(core.Browse.CurrentWidth)-10.0 {
			inputBoxFocused = true
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
			inputBoxFocused = false
		}
		markNeedsRedraw()
	}
}

func scrollCallback(window *glfw.Window, xoff, yoff float64) {
	if !inputBoxFocused {
		if window.GetKey(glfw.KeyLeftControl) == glfw.Press ||
			window.GetKey(glfw.KeyRightControl) == glfw.Press {
			adjustZoom(float32(yoff) * 0.1)
		} else {
			scrollOffset += float32(yoff) * 25.0
			updateScrollLimits()
		}
		markNeedsRedraw()
	}
}

func renderHTML(program uint32) {
	if htmlRenderer == nil {
		return
	}

	availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Program:      program,
		X:            10.0 * zoom,
		Y:            core.Browse.InputBoxHeight + 15.0*zoom,
		Width:        float32(core.Browse.CurrentWidth) - 20.0*zoom,
		Height:       availableHeight,
		ScrollOffset: scrollOffset,
		Zoom:         zoom,
	}

	if err := htmlRenderer.Render(ctx); err != nil {
		TextLIB.DrawText(program, "HTML Render Error: "+err.Error(),
			10.0*zoom, core.Browse.InputBoxHeight+15.0*zoom, zoom, utils.RGBToFloat32(255, 100, 100))
	}
}

func updateContent(link string, ua string) web.HTMLRenderer {
	req, err := h.GETRequest(link, ua)
	if err != nil {
		errorHTML := `
						<!DOCTYPE html>
						<html>
							<head>
								<title>Error</title>
							</head>
							<body>
								<h1>Failed to load page</h1>
								<p>Error: ` + err.Error() + `</p>
								<p>Please check your internet connection and try again.</p>
							</body>
						</html>
					`
		htmlRenderer = web.NewHTMLRenderer(errorHTML)
	} else {
		htmlRenderer = web.NewHTMLRenderer(req.Page)
	}
	return *htmlRenderer
}

func initializeWindowState(window *glfw.Window) {
	isMaximized = window.GetAttrib(glfw.Maximized) == glfw.True
	windowedWidth, windowedHeight = window.GetSize()
	windowedX, windowedY = window.GetPos()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalf("glfw ? %v", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Maximized, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(1280, 720, "Himera", nil, nil)
	if err != nil {
		log.Fatalf("glfw Window ? %v", err)
	}
	defer drawer.CleanupGLResources()
	window.MakeContextCurrent()
	window.SetMaximizeCallback(windowMaximizeCallback)
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetKeyCallback(keyCallback)
	window.SetCharCallback(charCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	window.SetScrollCallback(scrollCallback)

	initializeWindowState(window)

	if err := gl.Init(); err != nil {
		log.Fatalf("init gl ? %v", err)
	}

	prgs, err := draw.MakeShadersPrgs()
	if err != nil {
		log.Fatalf("shader program ? %v", err)
	}

	draw.ToBinare(prgs)

	f, err := os.Open("Shaders.bin")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	data := draw.ShadersPrograms{}
	binary.Read(f, binary.LittleEndian, &data)
	textPrg := data.TextShaderProgram

	if err := TextLIB.InitFont(); err != nil {
		log.Fatalf("init font ? %v", err)
	}

	drawer.InitGLResources()

	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	updateProjection(textPrg)
	htmlRenderer := updateContent(core.Browse.Link, core.Browse.Ua)
	styles := &web.StyleConfig{
		TextColor:    utils.RGBToFloat32(240, 240, 240),
		LinkColor:    utils.RGBToFloat32(100, 149, 237),
		HeadingColor: utils.RGBToFloat32(255, 255, 255),

		H1Size:    2.0,
		H2Size:    1.5,
		H3Size:    1.17,
		BaseSize:  1.0,
		SmallSize: 0.8,

		ParagraphSpacing: 16.0,
		LineSpacing:      1.4,
		IndentSize:       20.0,

		H1MarginTop:    24.0,
		H1MarginBottom: 16.0,
		H2MarginTop:    20.0,
		H2MarginBottom: 12.0,
		H3MarginTop:    16.0,
		H3MarginBottom: 8.0,
	}

	htmlRenderer.SetStyles(styles)
	updateScrollLimits()

	glfw.SwapInterval(1)

	for !window.ShouldClose() {
		glfw.WaitEventsTimeout(0.016)

		if checkNeedsRedraw() {
			if renderState.lastWidth != core.Browse.CurrentWidth ||
				renderState.lastHeight != core.Browse.CurrentHeight ||
				renderState.lastZoom != zoom {
				updateProjection(textPrg)
			}

			if inputBoxFocused {
				window.SetCursor(glfw.CreateStandardCursor(glfw.IBeamCursor))
			} else {
				window.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))
			}

			gl.Clear(gl.COLOR_BUFFER_BIT)
			renderHTML(textPrg)
			himera.DrawInputBox(data.RectShaderProgram, textPrg)
			window.SwapBuffers()
		}
	}
}
