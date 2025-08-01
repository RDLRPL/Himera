package main

import (
	"log"
	"runtime"
	"strings"
	"time"

	h "github.com/RDLRPL/Himera/HDS/core/http"
	draw "github.com/RDLRPL/Himera/HGD/Draw"
	"github.com/RDLRPL/Himera/HGD/Draw/TextLIB"
	"github.com/RDLRPL/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var Monitor, _ = utils.GetPrimaryMonitor()

var (
	currentWidth               = Monitor.Width
	currentHeight              = Monitor.Height
	zoom                       = float32(1.0)
	scrollOffset       float32 = 0.0
	changeScrollOffset float32 = 0.0
	isFullscreen               = false
	windowedWidth              = Monitor.Width
	windowedHeight             = Monitor.Height
)

func init() {
	runtime.LockOSThread()
}

func renderMultilineText(program uint32, text string, x, y, scale float32, color [3]float32, lineSpacing float32) {
	lines := strings.Split(text, "\n")
	lineHeight := float32(TextLIB.FontMetrics.Height>>6) * scale * lineSpacing

	startY := y + scrollOffset

	for i, line := range lines {
		TextLIB.DrawText(program, line, x, startY+float32(i)*lineHeight, scale, color)
	}
}
func updateProjection(program uint32) {
	projection := [16]float32{
		2.0 / float32(currentWidth), 0, 0, 0,
		0, -2.0 / float32(currentHeight), 0, 0,
		0, 0, 1, 0,
		-1, 1, 0, 1,
	}

	gl.UseProgram(program)
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])
}

func toggleFullscreen(window *glfw.Window) {
	if isFullscreen {
		window.SetMonitor(nil, 100, 100, windowedWidth, windowedHeight, 0)
		currentWidth = windowedWidth
		currentHeight = windowedHeight
		isFullscreen = false
	} else {
		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		currentWidth = mode.Width
		currentHeight = mode.Height
		isFullscreen = true
	}
	gl.Viewport(0, 0, int32(currentWidth), int32(currentHeight))
}

func adjustZoom(delta float32) {
	zoom += delta
	if zoom < 0.1 {
		zoom = 0.1
	} else if zoom > 5.0 {
		zoom = 5.0
	}
	scrollOffset = 0
}

func framebufferSizeCallback(window *glfw.Window, width, height int) {
	currentWidth = width
	currentHeight = height
	if !isFullscreen {
		windowedWidth = width
		windowedHeight = height
	}
	gl.Viewport(0, 0, int32(width), int32(height))
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.KeyEscape:
			if isFullscreen {
				toggleFullscreen(window)
			} else {
				window.SetShouldClose(true)
			}
		case glfw.KeyF11:
			toggleFullscreen(window)
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
			}
		}
	}
}

func scrollCallback(window *glfw.Window, xoff, yoff float64) {
	if window.GetKey(glfw.KeyLeftControl) == glfw.Press ||
		window.GetKey(glfw.KeyRightControl) == glfw.Press {

		adjustZoom(float32(yoff) * 0.1)
	} else {
		changeScrollOffset = scrollOffset + float32(yoff)*25.0
		scrollOffset = changeScrollOffset

	}

}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalf("failed to initialize glfw: %v", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(Monitor.Width, Monitor.Height, "Himera", nil, nil)
	if err != nil {
		log.Fatalf("failed to create window: %v", err)
	}
	window.MakeContextCurrent()

	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetKeyCallback(keyCallback)
	window.SetScrollCallback(scrollCallback)

	if err := gl.Init(); err != nil {
		log.Fatalf("failed to initialize gl: %v", err)
	}

	program, err := draw.CreateShaderProgram()
	if err != nil {
		log.Fatalf("failed to create shader program: %v", err)
	}

	if err := TextLIB.InitFont(); err != nil {
		log.Fatalf("failed to initialize font: %v", err)
	}

	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	updateProjection(program)

	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	req, _ := h.GETRequest("https://darq-project.ru", "Himera/0.1B (Furry♥ X64; PurryForno*x86_64; x64; ver:=001B) HDS/001B Himera/0.1B")

	var lastWidth, lastHeight int = currentWidth, currentHeight
	var lastZoom float32 = zoom

	for !window.ShouldClose() {
		time.Sleep(time.Millisecond * 16)
		if currentWidth != lastWidth || currentHeight != lastHeight || zoom != lastZoom {
			updateProjection(program)
			lastWidth = currentWidth
			lastZoom = zoom
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)

		effectiveScale := zoom * 1.0
		renderMultilineText(program, req.Page, zoom+10, zoom, effectiveScale, utils.RGBToFloat32(255, 255, 255), 1.2)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
