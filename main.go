package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	h "github.com/RDLRPL/Himera/HDS/core/http"
	"github.com/RDLRPL/Himera/HDS/core/utils"
	"github.com/RDLRPL/Himera/HGD/Draw/TextLIB"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var Monitor, _ = utils.GetPrimaryMonitor()

var (
	currentWidth   = Monitor.Width
	currentHeight  = Monitor.Height
	zoom           = float32(1.0)
	isFullscreen   = false
	windowedWidth  = Monitor.Width
	windowedHeight = Monitor.Height
)

var vertexShaderSource = `
#version 410
layout (location = 0) in vec4 vertex;
out vec2 TexCoords;
uniform mat4 projection;

void main() {
    gl_Position = projection * vec4(vertex.xy, 0.0, 1.0);
    TexCoords = vertex.zw;
}
` + "\x00"

var fragmentShaderSource = `
#version 410
in vec2 TexCoords;
out vec4 color;
uniform sampler2D text;
uniform vec3 textColor;
void main() {
    vec4 sampled = vec4(1.0, 1.0, 1.0, texture(text, TexCoords).r);
    color = vec4(textColor, 1.0) * sampled;
}
` + "\x00"

func init() {
	runtime.LockOSThread()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func createShaderProgram() (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("failed to compile vertex shader: %v", err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, fmt.Errorf("failed to compile fragment shader: %v", err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

// Функция для рендеринга многострочного текста
func renderMultilineText(program uint32, text string, x, y float32, scale float32, color [3]float32, lineSpacing float32) {
	lines := strings.Split(text, "\n")
	lineHeight := float32(TextLIB.FontMetrics.Height>>6) * scale * lineSpacing

	for i, line := range lines {
		TextLIB.DrawText(program, line, x, y+float32(i)*lineHeight, scale, color)
	}
}

// Функция для обновления матрицы проекции
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

// Функция для переключения полноэкранного режима
func toggleFullscreen(window *glfw.Window) {
	if isFullscreen {
		// Выход из полноэкранного режима
		window.SetMonitor(nil, 100, 100, windowedWidth, windowedHeight, 0)
		currentWidth = windowedWidth
		currentHeight = windowedHeight
		isFullscreen = false
	} else {
		// Вход в полноэкранный режим
		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		currentWidth = mode.Width
		currentHeight = mode.Height
		isFullscreen = true
	}
	gl.Viewport(0, 0, int32(currentWidth), int32(currentHeight))
}

// Функция для изменения масштаба
func adjustZoom(delta float32) {
	zoom += delta
	if zoom < 0.1 {
		zoom = 0.1
	} else if zoom > 5.0 {
		zoom = 5.0
	}
}

// Callback для изменения размера окна
func framebufferSizeCallback(window *glfw.Window, width, height int) {
	currentWidth = width
	currentHeight = height
	if !isFullscreen {
		windowedWidth = width
		windowedHeight = height
	}
	gl.Viewport(0, 0, int32(width), int32(height))
}

// Callback для клавиатуры
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
		case glfw.KeyEqual, glfw.KeyKPAdd: // + или = для увеличения
			if mods&glfw.ModControl != 0 {
				adjustZoom(0.1)
			}
		case glfw.KeyMinus, glfw.KeyKPSubtract: // - для уменьшения
			if mods&glfw.ModControl != 0 {
				adjustZoom(-0.1)
			}
		case glfw.Key0, glfw.KeyKP0: // 0 для сброса масштаба
			if mods&glfw.ModControl != 0 {
				zoom = 1.0
			}
		}
	}
}

// Callback для скролла мыши
func scrollCallback(window *glfw.Window, xoff, yoff float64) {
	// Ctrl + scroll для масштабирования
	if window.GetKey(glfw.KeyLeftControl) == glfw.Press || window.GetKey(glfw.KeyRightControl) == glfw.Press {
		adjustZoom(float32(yoff) * 0.1)
	}
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalf("failed to initialize glfw: %v", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True) // Разрешаем изменение размера
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(Monitor.Width, Monitor.Height, "System Info - F11: Fullscreen, Ctrl+/-: Zoom, Ctrl+0: Reset Zoom", nil, nil)
	if err != nil {
		log.Fatalf("failed to create window: %v", err)
	}
	window.MakeContextCurrent()

	// Устанавливаем callback'и
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetKeyCallback(keyCallback)
	window.SetScrollCallback(scrollCallback)

	if err := gl.Init(); err != nil {
		log.Fatalf("failed to initialize gl: %v", err)
	}

	program, err := createShaderProgram()
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

	req, _ := h.GETRequest("https://darq-project.ru/", "Himera/0.1B (Furry♥ X64; PurryForno*x86_64; x64; ver:=001B) HDS/001B Himera/0.1B")

	var lastWidth, lastHeight int = currentWidth, currentHeight
	var lastZoom float32 = zoom

	for !window.ShouldClose() {
		if currentWidth != lastWidth || currentHeight != lastHeight || zoom != lastZoom {
			updateProjection(program)
			lastWidth = currentWidth
			lastZoom = zoom
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)

		effectiveScale := zoom * 1.0
		renderMultilineText(program, req.Page, 50*zoom, 50*zoom, effectiveScale, [3]float32{0.9, 0.6, 0.1}, 1.2)

		zoomInfo := fmt.Sprintf("Zoom: %.1fx", zoom)
		TextLIB.DrawText(program, zoomInfo, float32(currentWidth)-200*zoom, 30*zoom, zoom, utils.RGBToFloat32(255, 255, 255))

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
