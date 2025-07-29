package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	h "github.com/RDLRPL/Himera/HDS/core/http"
	"github.com/RDLRPL/Himera/HDS/core/render/web"
)

func createScaledTheme(zoomFactor float32) *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	th.TextSize = unit.Sp(14 * zoomFactor)

	return th
}

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Himera"))
		var htmlstr string
		var ops op.Ops
		var dataLoaded bool

		var zoomFactor float32 = 1.0
		const minZoom = 0.5
		const maxZoom = 3.0
		const zoomStep = 0.2

		var cachedTheme *material.Theme
		var cachedZoom float32 = -1

		var list widget.List
		list.Axis = layout.Vertical

		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				zoomChanged := false
				for {
					if e, ok := gtx.Event(key.Filter{Required: key.ModCtrl}); ok {
						if ev, ok := e.(key.Event); ok && ev.State == key.Press {
							switch ev.Name {
							case "+", "Num+":
								newZoom := zoomFactor + zoomStep
								if newZoom <= maxZoom {
									zoomFactor = newZoom
									zoomChanged = true
								}
							case "-", "Num-":
								newZoom := zoomFactor - zoomStep
								if newZoom >= minZoom {
									zoomFactor = newZoom
									zoomChanged = true
								}
							case "0":
								if zoomFactor != 1.0 {
									zoomFactor = 1.0
									zoomChanged = true
								}
							}
						}
						continue
					}
					break
				}

				if cachedTheme == nil || cachedZoom != zoomFactor || zoomChanged {
					cachedTheme = createScaledTheme(zoomFactor)
					cachedZoom = zoomFactor
				}
				go func() {
					req, _ := h.GETRequest("https://ria.ru/lenta/", "Himera/0.1B (Furry♥ X64; PurryForno_x86_64; x64; ver:=001B) HDS/001B Himera/0.1B")

					dataLoaded = req.Done
					htmlstr = req.Page
				}()
				if dataLoaded {
					material.List(cachedTheme, &list).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
						return layout.Inset{
							Top:    unit.Dp(10 * zoomFactor),
							Bottom: unit.Dp(10 * zoomFactor),
							Left:   unit.Dp(10 * zoomFactor),
							Right:  unit.Dp(10 * zoomFactor),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							htmlDrEngine := web.NewDRW(gtx, cachedTheme, htmlstr)
							return htmlDrEngine.RenderHTML()
						})
					})
				} else {
					if cachedTheme == nil {
						cachedTheme = createScaledTheme(1.0)
					}
					layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Body1(cachedTheme, "...").Layout(gtx)
					})
				}

				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
