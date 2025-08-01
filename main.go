package main

import (
	"log"
	"runtime"
	"time"

	web "github.com/RDLRPL/Himera/HDS/core/html"
	h "github.com/RDLRPL/Himera/HDS/core/http"
	draw "github.com/RDLRPL/Himera/HGD/Draw"
	"github.com/RDLRPL/Himera/HGD/Draw/TextLIB"
	"github.com/RDLRPL/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var Monitor, _ = utils.GetPrimaryMonitor()

var (
	currentWidth           = Monitor.Width
	currentHeight          = Monitor.Height
	zoom           float32 = 1.0
	scrollOffset   float32 = 0.0
	isFullscreen           = false
	windowedWidth          = Monitor.Width
	windowedHeight         = Monitor.Height

	// HTML рендерер
	htmlRenderer  *web.HTMLRenderer
	contentHeight float32 = 0.0
)

func init() {
	runtime.LockOSThread()
}

func updateProjection(program uint32) {
	projection := [16]float32{
		2.0 / float32(currentWidth), 0, 0, 0,
		0, -2.0 / float32(currentHeight), 0, 0,
		0, 0, -1, 0,
		-1, 1, 0, 1,
	}

	gl.UseProgram(program)
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])
}

func toggleFullscreen(window *glfw.Window) {
	if isFullscreen {
		// Возврат в оконный режим
		window.SetMonitor(nil, 100, 100, windowedWidth, windowedHeight, 0)
		currentWidth = windowedWidth
		currentHeight = windowedHeight
		isFullscreen = false
	} else {
		// Переход в полноэкранный режим
		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		currentWidth = mode.Width
		currentHeight = mode.Height
		isFullscreen = true
	}
	gl.Viewport(0, 0, int32(currentWidth), int32(currentHeight))

	// Пересчитываем высоту контента при изменении размера окна
	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(currentWidth),
			Height: float32(currentHeight),
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
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
		scrollOffset = 0 // Сброс скролла при изменении масштаба

		// Пересчитываем высоту контента
		if htmlRenderer != nil {
			ctx := &web.RenderContext{
				Width:  float32(currentWidth),
				Height: float32(currentHeight),
				Zoom:   zoom,
			}
			contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		}
	}
}

func updateScrollLimits() {
	maxScrollOffset := float32(0.0)
	minScrollOffset := -float32(9900.0)

	if minScrollOffset > 0 {
		minScrollOffset = 0
	}

	// Ограничиваем текущий скролл новыми лимитами
	if scrollOffset > maxScrollOffset {
		scrollOffset = maxScrollOffset
	}
	if scrollOffset < minScrollOffset {
		scrollOffset = minScrollOffset
	}
}

func framebufferSizeCallback(window *glfw.Window, width, height int) {
	currentWidth = width
	currentHeight = height
	if !isFullscreen {
		windowedWidth = width
		windowedHeight = height
	}
	gl.Viewport(0, 0, int32(width), int32(height))

	// Пересчитываем высоту контента при изменении размера окна
	if htmlRenderer != nil {
		ctx := &web.RenderContext{
			Width:  float32(currentWidth),
			Height: float32(currentHeight),
			Zoom:   zoom,
		}
		contentHeight = htmlRenderer.CalculateContentHeight(ctx)
		updateScrollLimits()
	}
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
				scrollOffset = 0
				if htmlRenderer != nil {
					ctx := &web.RenderContext{
						Width:  float32(currentWidth),
						Height: float32(currentHeight),
						Zoom:   zoom,
					}
					contentHeight = htmlRenderer.CalculateContentHeight(ctx)
				}
			}
		case glfw.KeyHome:
			scrollOffset = 0
		case glfw.KeyEnd:
			updateScrollLimits()
			scrollOffset = -(contentHeight - float32(currentHeight)*0.9)
			if scrollOffset > 0 {
				scrollOffset = 0
			}
		case glfw.KeyPageUp:
			scrollOffset += float32(currentHeight) * 0.8
			updateScrollLimits()
		case glfw.KeyPageDown:
			scrollOffset -= float32(currentHeight) * 0.8
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

func scrollCallback(window *glfw.Window, xoff, yoff float64) {
	if window.GetKey(glfw.KeyLeftControl) == glfw.Press ||
		window.GetKey(glfw.KeyRightControl) == glfw.Press {
		// Масштабирование с помощью Ctrl+Scroll
		adjustZoom(float32(yoff) * 0.1)
	} else {
		// Обычный скроллинг
		scrollOffset += float32(yoff) * 25.0
		updateScrollLimits()
	}
}

func renderHTML(program uint32) {
	if htmlRenderer == nil {
		return
	}

	ctx := &web.RenderContext{
		Program:      program,
		X:            10.0 * zoom,
		Y:            10.0 * zoom,
		Width:        float32(currentWidth) - 20.0*zoom,
		Height:       float32(currentHeight),
		ScrollOffset: scrollOffset,
		Zoom:         zoom,
	}

	if err := htmlRenderer.Render(ctx); err != nil {
		log.Printf("HTML render error: %v", err)
		TextLIB.DrawText(program, "HTML Render Error: "+err.Error(),
			10.0*zoom, 10.0*zoom, zoom, utils.RGBToFloat32(255, 100, 100))
	}
}

func main() {
	// Инициализация GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalf("failed to initialize glfw: %v", err)
	}
	defer glfw.Terminate()

	// Настройка окна
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

	// Установка коллбэков
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetKeyCallback(keyCallback)
	window.SetScrollCallback(scrollCallback)

	// Инициализация OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalf("failed to initialize gl: %v", err)
	}

	// Создание шейдерной программы
	program, err := draw.CreateShaderProgram()
	if err != nil {
		log.Fatalf("failed to create shader program: %v", err)
	}

	// Инициализация шрифтов
	if err := TextLIB.InitFont(); err != nil {
		log.Fatalf("failed to initialize font: %v", err)
	}

	// Настройка OpenGL
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	updateProjection(program)

	log.Println("Loading HTML content...")
	req, err := h.GETRequest("https://ai.onlysq.ru", "Himera/0.1B (FURRY PORN_X64 Linux; X64) HDS/001B")
	if err != nil {
		log.Printf("Failed to load HTML: %v", err)
		errorHTML := `<!DOCTYPE html>
			<html>
			<head><title>Error</title></head>
			<body>
			<h1>Failed to load page</h1>
			<p>Error: ` + err.Error() + `</p>
			<p>Please check your internet connection and try again.</p>
			</body>
			</html>`
		htmlRenderer = web.NewHTMLRenderer(errorHTML)
	} else {
		log.Println("HTML content loaded successfully")
		htmlRenderer = web.NewHTMLRenderer(req.Page)
	}

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

	ctx := &web.RenderContext{
		Width:  float32(currentWidth),
		Height: float32(currentHeight),
		Zoom:   zoom,
	}
	contentHeight = htmlRenderer.CalculateContentHeight(ctx)
	updateScrollLimits()

	var lastWidth, lastHeight int = currentWidth, currentHeight
	var lastZoom float32 = zoom

	for !window.ShouldClose() {
		time.Sleep(time.Millisecond * 16)

		// Обновляем проекцию при изменении размеров или масштаба
		if currentWidth != lastWidth || currentHeight != lastHeight || zoom != lastZoom {
			updateProjection(program)
			lastWidth = currentWidth
			lastHeight = currentHeight
			lastZoom = zoom
		}

		// Очищаем экран
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Рендерим HTML
		renderHTML(program)

		// Обмен буферов и обработка событий
		window.SwapBuffers()
		glfw.PollEvents()
	}

	log.Println("Browser closed")
}
