package main

import (
	"image"
	"os"
	"sync"
	"time"

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

type AppState struct {
	htmlstr              string
	dataLoaded           bool
	loading              bool
	cachedTheme          *material.Theme
	mutex                sync.RWMutex
	zoomFactor           float32
	scrollList           widget.List
	LCanvasWidthPercent  float32
	RCanvasWidthPercent  float32
	UCanvasHeightPercent float32
	DCanvasHeightPercent float32
}

func (s *AppState) SetLoading(loading bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loading = loading
}

func (s *AppState) SetData(htmlstr string, loaded bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.htmlstr = htmlstr
	s.dataLoaded = loaded
	s.loading = false
}

func (s *AppState) GetState() (string, bool, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.htmlstr, s.dataLoaded, s.loading
}

func (s *AppState) SetZoom(zoom float32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.zoomFactor = zoom
}

func (s *AppState) GetZoom() float32 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.zoomFactor
}

func (s *AppState) GetScrollList() *widget.List {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return &s.scrollList
}

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

		var ops op.Ops
		appState := &AppState{
			zoomFactor:           1.0,
			scrollList:           widget.List{List: layout.List{Axis: layout.Vertical}},
			LCanvasWidthPercent:  0.0,
			RCanvasWidthPercent:  0.0,
			UCanvasHeightPercent: 0.04,
			DCanvasHeightPercent: 0.0,
		}

		const minZoom = 0.5
		const maxZoom = 3.0
		const zoomStep = 0.2

		updateChan := make(chan struct{}, 1)

		appState.cachedTheme = createScaledTheme(1.0)

		go func() {
			appState.SetLoading(true)
			select {
			case updateChan <- struct{}{}:
			default:
			}

			req, err := h.GETRequest("https://max.ru/", "Himera/0.1B (Furry♥ X64; PurryForno*x86_64; x64; ver:=001B) HDS/001B Himera/0.1B")
			if err != nil {
				appState.SetData("", false)
			} else {
				appState.SetData(req.Page, req.Done)
			}

			select {
			case updateChan <- struct{}{}:
			default:
			}
			w.Invalidate()
		}()

		go func() {
			ticker := time.NewTicker(16 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-updateChan:
					w.Invalidate()
				case <-ticker.C:
					htmlstr, dataLoaded, loading := appState.GetState()
					if loading || (!dataLoaded && htmlstr == "") {
						w.Invalidate()
					}
				}
			}
		}()

		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				zoomChanged := false
				currentZoom := appState.GetZoom()

				for {
					if e, ok := gtx.Event(key.Filter{Required: key.ModCtrl}); ok {
						if ev, ok := e.(key.Event); ok && ev.State == key.Press {
							switch ev.Name {
							case "+", "Num+", "=":
								newZoom := currentZoom + zoomStep
								if newZoom <= maxZoom {
									appState.SetZoom(newZoom)
									zoomChanged = true
								}
							case "-", "Num-":
								newZoom := currentZoom - zoomStep
								if newZoom >= minZoom {
									appState.SetZoom(newZoom)
									zoomChanged = true
								}
							case "0":
								if currentZoom != 1.0 {
									appState.SetZoom(1.0)
									zoomChanged = true
								}
							}
						}
						continue
					}
					break
				}

				if appState.cachedTheme == nil || zoomChanged {
					appState.cachedTheme = createScaledTheme(appState.GetZoom())
				}

				htmlstr, dataLoaded, loading := appState.GetState()

				layout.Stack{}.Layout(gtx,
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						appState.mutex.RLock()
						left := appState.LCanvasWidthPercent
						right := appState.RCanvasWidthPercent
						top := appState.UCanvasHeightPercent
						bottom := appState.DCanvasHeightPercent
						appState.mutex.RUnlock()

						totalWidth := gtx.Constraints.Max.X
						totalHeight := gtx.Constraints.Max.Y

						xOffset := int(float32(totalWidth) * left)
						yOffset := int(float32(totalHeight) * top)
						width := int(float32(totalWidth) * (1.0 - left - right))
						height := int(float32(totalHeight) * (1.0 - top - bottom))

						gtx.Constraints.Min.X = 0
						gtx.Constraints.Min.Y = 0
						gtx.Constraints.Max.X = width
						gtx.Constraints.Max.Y = height

						call := op.Offset(image.Pt(xOffset, yOffset)).Push(gtx.Ops)
						defer call.Pop()

						if dataLoaded && htmlstr != "" {
							scrollList := appState.GetScrollList()
							return material.List(appState.cachedTheme, scrollList).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
								return layout.Inset{
									Top:    unit.Dp(10 * appState.GetZoom()),
									Bottom: unit.Dp(10 * appState.GetZoom()),
									Left:   unit.Dp(10 * appState.GetZoom()),
									Right:  unit.Dp(10 * appState.GetZoom()),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									htmlDrEngine := web.NewDRW(gtx, appState.cachedTheme, htmlstr)
									return htmlDrEngine.RenderHTML()
								})
							})
						}

						if loading {
							return material.H2(appState.cachedTheme, "Sending Request 🔄 Nya>.<").Layout(gtx)
						}
						return material.H2(appState.cachedTheme, "Loading failed").Layout(gtx)
					}),
				)

				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
