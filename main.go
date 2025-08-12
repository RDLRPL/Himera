package main

import (
	"encoding/binary"
	"log"
	"os"
	"runtime"

	web "github.com/RDLxxx/Himera/HDS/core/html"
	draw "github.com/RDLxxx/Himera/HGD/Draw"
	drawer "github.com/RDLxxx/Himera/HGD/Draw/Drawer"
	himera "github.com/RDLxxx/Himera/HGD/Draw/Himera"
	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/RDLxxx/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
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
	window.SetMaximizeCallback(himera.WindowMaximizeCallback)
	window.SetFramebufferSizeCallback(himera.FramebufferSizeCallback)
	window.SetKeyCallback(himera.KeyCallback)
	window.SetCharCallback(himera.CharCallback)
	window.SetMouseButtonCallback(himera.MouseButtonCallback)
	window.SetScrollCallback(himera.ScrollCallback)

	himera.InitializeWindowState(window)

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

	himera.UpdateProjection(textPrg)
	htmlRenderer := himera.UpdateContent(core.Browse.Link, core.Browse.Ua)
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
	himera.UpdateScrollLimits()

	glfw.SwapInterval(1)

	for !window.ShouldClose() {
		glfw.WaitEventsTimeout(0.016)

		if himera.CheckNeedsRedraw() {
			if core.Browse.RState.LastWidth != core.Browse.CurrentWidth ||
				core.Browse.RState.LastHeight != core.Browse.CurrentHeight ||
				core.Browse.RState.LastZoom != core.Browse.Zoom {
				himera.UpdateProjection(textPrg)
			}

			if core.Browse.InputBoxFocused {
				window.SetCursor(glfw.CreateStandardCursor(glfw.IBeamCursor))
			} else {
				window.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))
			}

			gl.Clear(gl.COLOR_BUFFER_BIT)
			himera.RenderHTML(textPrg)
			himera.DrawURLBox(data.RectShaderProgram, textPrg)
			window.SwapBuffers()
		}
	}
}
