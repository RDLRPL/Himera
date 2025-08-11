package main

import (
	"encoding/binary"
	"log"
	"os"
	"runtime"
	"time"
	"unicode"

	web "github.com/RDLRPL/Himera/HDS/core/html"
	h "github.com/RDLRPL/Himera/HDS/core/http"
	draw "github.com/RDLRPL/Himera/HGD/Draw"
	"github.com/RDLRPL/Himera/HGD/Draw/TextLIB"
	"github.com/RDLRPL/Himera/HGD/browser"
	"github.com/RDLRPL/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var Monitor, _ = utils.GetPrimaryMonitor()
var (
	zoom         float32 = 1.0
	scrollOffset float32 = 0.0
	isFullscreen         = false
	isMaximized          = false

	htmlRenderer  *web.HTMLRenderer
	contentHeight float32 = 0.0

	curlink string = "https://example.com"
	gua     string = "(FurryPornox64 HimeraBrowsrx000)"

	windowedX, windowedY, windowedWidth, windowedHeight int
	wasMaximizedBeforeFullscreen                        bool

	inputBoxFocused bool = false

	inputText      string = curlink
	cursorPosition int    = len(curlink)

	inputBoxHeight float32 = 40.0
	blinkTimer     float64 = 0.0
)

var Browse = browser.NewBrowser(Monitor.Width, Monitor.Height)

func init() {
	runtime.LockOSThread()
}

func updateProjection(program uint32) {
	projection := [16]float32{
		2.0 / float32(Browse.CurrentWidth), 0, 0, 0,
		0, -2.0 / float32(Browse.CurrentHeight), 0, 0,
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
		Browse.CurrentWidth = mode.Width
		Browse.CurrentHeight = mode.Height
	}

	gl.Viewport(0, 0, int32(Browse.CurrentWidth), int32(Browse.CurrentHeight))

	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(Browse.CurrentWidth),
			Height: float32(Browse.CurrentHeight) - inputBoxHeight - 20.0,
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		updateScrollLimits()
	}
}

func windowMaximizeCallback(window *glfw.Window, maximized bool) {
	if !isFullscreen {
		isMaximized = maximized
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
				Width:  float32(Browse.CurrentWidth),
				Height: float32(Browse.CurrentHeight) - inputBoxHeight - 20.0,
				Zoom:   zoom,
			}
			contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		}
	}
}

func updateScrollLimits() {
	if htmlRenderer == nil {
		return
	}

	availableHeight := float32(Browse.CurrentHeight) - inputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Width:  float32(Browse.CurrentWidth) - 20.0*zoom,
		Height: availableHeight,
		X:      10.0 * zoom,
		Y:      inputBoxHeight + 15.0*zoom,
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
	Browse.CurrentWidth = width
	Browse.CurrentHeight = height
	gl.Viewport(0, 0, int32(width), int32(height))

	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(Browse.CurrentWidth),
			Height: float32(Browse.CurrentHeight) - inputBoxHeight - 10.0,
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		updateScrollLimits()
	}
}

func charCallback(window *glfw.Window, char rune) {
	if inputBoxFocused && unicode.IsPrint(char) {
		inputText = inputText[:cursorPosition] + string(char) + inputText[cursorPosition:]
		cursorPosition++
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		if inputBoxFocused {
			switch key {
			case glfw.KeyEnter:
				curlink = inputText
				updateContent(curlink, gua)
				inputBoxFocused = false
				return
			case glfw.KeyEscape:
				inputBoxFocused = false
				return
			case glfw.KeyBackspace:
				if cursorPosition > 0 {
					inputText = inputText[:cursorPosition-1] + inputText[cursorPosition:]
					cursorPosition--
				}
				return
			case glfw.KeyDelete:
				if cursorPosition < len(inputText) {
					inputText = inputText[:cursorPosition] + inputText[cursorPosition+1:]
				}
				return
			case glfw.KeyLeft:
				if cursorPosition > 0 {
					cursorPosition--
				}
				return
			case glfw.KeyRight:
				if cursorPosition < len(inputText) {
					cursorPosition++
				}
				return
			case glfw.KeyHome:
				cursorPosition = 0
				return
			case glfw.KeyEnd:
				cursorPosition = len(inputText)
				return
			case glfw.KeyA:
				if mods&glfw.ModControl != 0 {
					cursorPosition = len(inputText)
					return
				}
			case glfw.KeyV:
				if mods&glfw.ModControl != 0 {
					return
				}
			}
		}

		switch key {
		case glfw.KeyF5:
			updateContent(curlink, gua)
		case glfw.KeyF11:
			toggleFullscreen(window)
		case glfw.KeyL:
			if mods&glfw.ModControl != 0 {
				inputBoxFocused = true
				cursorPosition = len(inputText)
				return
			}
		case glfw.KeyEqual, glfw.KeyKPAdd:
			if mods&glfw.ModControl != 0 {
				adjustZoom(0.1)
			}
		case glfw.KeyMinus, glfw.KeyKPSubtract:
			if mods&glfw.ModControl != 0 {
				adjustZoom(-0.1)
			}
		case glfw.Key0, glfw.KeyKP0:
			if mods&glfw.ModControl != 0 {
				zoom = 1.0
				scrollOffset = 0
				if htmlRenderer != nil {
					ctx := &web.RenderContext{
						Width:  float32(Browse.CurrentWidth),
						Height: float32(Browse.CurrentHeight) - inputBoxHeight - 10.0,
						Zoom:   zoom,
					}
					contentHeight = htmlRenderer.CalculateContentHeight(ctx)
				}
			}
		}

		if !inputBoxFocused {
			switch key {
			case glfw.KeyHome:
				scrollOffset = 0
			case glfw.KeyEnd:
				updateScrollLimits()
				availableHeight := float32(Browse.CurrentHeight) - inputBoxHeight - 20.0
				scrollOffset = -(contentHeight - availableHeight*0.9)
				if scrollOffset > 0 {
					scrollOffset = 0
				}
			case glfw.KeyPageUp:
				scrollOffset += float32(Browse.CurrentHeight) * 0.8
				updateScrollLimits()
			case glfw.KeyPageDown:
				scrollOffset -= float32(Browse.CurrentHeight) * 0.8
				updateScrollLimits()
			case glfw.KeyUp:
				scrollOffset += 50.0
				updateScrollLimits()
			case glfw.KeyDown:
				scrollOffset -= 50.0
				updateScrollLimits()
			}
		}
	}
}

func mouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && button == glfw.MouseButtonLeft {
		xpos, ypos := window.GetCursorPos()

		inputBoxY := float32(5.0)
		if float32(ypos) >= inputBoxY && float32(ypos) <= inputBoxY+inputBoxHeight &&
			float32(xpos) >= 10.0 && float32(xpos) <= float32(Browse.CurrentWidth)-10.0 {
			inputBoxFocused = true
			textWidth, _ := TextLIB.GetTextDimensions(inputText, 1.0)
			relativeX := float32(xpos) - 15.0
			if relativeX < 0 {
				cursorPosition = 0
			} else if relativeX > textWidth {
				cursorPosition = len(inputText)
			} else {

				cursorPosition = int(float32(len(inputText)) * (relativeX / textWidth))
				if cursorPosition < 0 {
					cursorPosition = 0
				}
				if cursorPosition > len(inputText) {
					cursorPosition = len(inputText)
				}
			}
		} else {
			inputBoxFocused = false
		}
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
	}

}
func drawRect(program uint32, x, y, width, height float32, color [3]float32) {
	vertices := []float32{
		x, y,
		x, y + height,
		x + width, y + height,
		x, y,
		x + width, y + height,
		x + width, y,
	}

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.UseProgram(program)

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	if projectionLoc >= 0 {
		projection := [16]float32{
			2.0 / float32(Browse.CurrentWidth), 0, 0, 0,
			0, -2.0 / float32(Browse.CurrentHeight), 0, 0,
			0, 0, -1, 0,
			-1, 1, 0, 1,
		}
		gl.UniformMatrix4fv(projectionLoc, 1, false, &projection[0])
	}

	colorLoc := gl.GetUniformLocation(program, gl.Str("fillColor\x00"))
	if colorLoc >= 0 {
		gl.Uniform3f(colorLoc, color[0], color[1], color[2])

	}

	if width > 0 && height > 0 {
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		if err := gl.GetError(); err != gl.NO_ERROR {
			log.Printf("OpenGL error: %d", err)
		}
	}

	gl.BindVertexArray(0)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
}

func drawInputBox(rectProgram uint32, textProgram uint32) {
	inputBoxY := float32(5.0)
	inputBoxWidth := float32(Browse.CurrentWidth) - 20.0

	drawRect(rectProgram, 10.0, inputBoxY, inputBoxWidth, inputBoxHeight,
		utils.RGBToFloat32(38, 38, 38))

	borderColor := utils.RGBToFloat32(0, 0, 0)
	if inputBoxFocused {
		borderColor = utils.RGBToFloat32(74, 74, 74)
	}
	drawRectOutline(rectProgram, 10.0, inputBoxY, inputBoxWidth, inputBoxHeight, 2.0, borderColor)

	gl.UseProgram(textProgram)

	textY := inputBoxY + inputBoxHeight/2 - TextLIB.GetLineHeight(1.0)/2 + TextLIB.GetFontAscent(1.0)
	TextLIB.DrawText(textProgram, inputText, 15.0, textY, 1.0,
		utils.RGBToFloat32(0, 0, 0))

	if inputBoxFocused {
		blinkTimer += 16.0
		if int(blinkTimer/500)%2 == 0 {
			cursorText := inputText[:cursorPosition]
			cursorX, _ := TextLIB.GetTextDimensions(cursorText, 1.0)
			drawRect(rectProgram, 15.0+cursorX, inputBoxY+5.0, 2.0, inputBoxHeight-10.0,
				[3]float32{0.0, 0.0, 0.0})
			gl.UseProgram(textProgram)
		}
	}

	if inputText == "" {
		TextLIB.DrawText(textProgram, "Url", 15.0, textY, 1.0,
			utils.RGBToFloat32(150, 150, 150))
	}
}

func drawRectOutline(program uint32, x, y, width, height, lineWidth float32, color [3]float32) {
	drawRect(program, x, y, width, lineWidth, color)
	drawRect(program, x, y+height-lineWidth, width, lineWidth, color)
	drawRect(program, x, y, lineWidth, height, color)
	drawRect(program, x+width-lineWidth, y, lineWidth, height, color)
}

func renderHTML(program uint32) {
	if htmlRenderer == nil {
		return
	}

	availableHeight := float32(Browse.CurrentHeight) - inputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Program:      program,
		X:            10.0 * zoom,
		Y:            inputBoxHeight + 15.0*zoom,
		Width:        float32(Browse.CurrentWidth) - 20.0*zoom,
		Height:       availableHeight,
		ScrollOffset: scrollOffset,
		Zoom:         zoom,
	}

	if err := htmlRenderer.Render(ctx); err != nil {
		TextLIB.DrawText(program, "HTML Render Error: "+err.Error(),
			10.0*zoom, inputBoxHeight+15.0*zoom, zoom, utils.RGBToFloat32(255, 100, 100))
	}
}

func updateContent(link string, ua string) web.HTMLRenderer {
	req, err := h.GETRequest(link, ua)
	if err != nil {
		log.Printf("Failed to load HTML: %v", err)
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
	log.Println(data)
	textPrg := data.TextShaderProgram

	if err := TextLIB.InitFont(); err != nil {
		log.Fatalf("init font ? %v", err)
	}

	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	updateProjection(textPrg)
	htmlRenderer := updateContent(curlink, gua)
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

	var lastWidth, lastHeight int = Browse.CurrentWidth, Browse.CurrentHeight
	var lastZoom float32 = zoom

	for !window.ShouldClose() {
		time.Sleep(time.Millisecond * 16)

		if Browse.CurrentWidth != lastWidth || Browse.CurrentHeight != lastHeight || zoom != lastZoom {
			updateProjection(textPrg)
			lastWidth = Browse.CurrentWidth
			lastHeight = Browse.CurrentHeight
			lastZoom = zoom
		}

		if inputBoxFocused {
			window.SetCursor(glfw.CreateStandardCursor(glfw.IBeamCursor))
		} else {

			window.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))
		}
		gl.Clear(gl.COLOR_BUFFER_BIT)

		renderHTML(textPrg)

		drawInputBox(data.RectShaderProgram, textPrg)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
