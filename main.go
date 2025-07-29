package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/RDLRPL/Himera/HDS/core/render/web"
)

// Функция для создания темы с масштабированием
func createScaledTheme(zoomFactor float32) *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	// Масштабируем размер текста
	th.TextSize = unit.Sp(14 * zoomFactor) // базовый размер * зум

	return th
}

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Himera"))

		var ops op.Ops
		var htmlString string
		var htmlMutex sync.RWMutex
		var dataLoaded bool

		// Переменные для зума
		var zoomFactor float32 = 1.0
		const minZoom = 0.5
		const maxZoom = 3.0
		const zoomStep = 0.2

		// Кешируем тему
		var cachedTheme *material.Theme
		var cachedZoom float32 = -1 // инициализируем недопустимым значением

		// Создаем виджет списка для прокрутки
		var list widget.List
		list.Axis = layout.Vertical

		go func() {
			res, err := http.Get("https://polytech.alabuga.ru")
			if err != nil {
				log.Printf("HTTP error: %v", err)
				return
			}
			defer res.Body.Close()

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}

			htmlMutex.Lock()
			htmlString = string(bodyBytes)
			dataLoaded = true
			htmlMutex.Unlock()

			w.Invalidate()
		}()

		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				// Обработка клавиш зума
				zoomChanged := false
				for {
					if e, ok := gtx.Event(key.Filter{Required: key.ModCtrl}); ok {
						if ev, ok := e.(key.Event); ok && ev.State == key.Press {
							switch ev.Name {
							case "+", "=": // + и = на одной клавише
								newZoom := zoomFactor + zoomStep
								if newZoom <= maxZoom {
									zoomFactor = newZoom
									zoomChanged = true
								}
							case "-":
								newZoom := zoomFactor - zoomStep
								if newZoom >= minZoom {
									zoomFactor = newZoom
									zoomChanged = true
								}
							case "0": // Сброс зума
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

				// Пересоздаем тему только если зум изменился
				if cachedTheme == nil || cachedZoom != zoomFactor || zoomChanged {
					cachedTheme = createScaledTheme(zoomFactor)
					cachedZoom = zoomFactor
				}

				htmlMutex.RLock()
				if dataLoaded {
					// Используем кешированную тему
					material.List(cachedTheme, &list).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
						// Добавляем отступы (тоже масштабируем)
						return layout.Inset{
							Top:    unit.Dp(10 * zoomFactor),
							Bottom: unit.Dp(10 * zoomFactor),
							Left:   unit.Dp(10 * zoomFactor),
							Right:  unit.Dp(10 * zoomFactor),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Рендеринг HTML контента с кешированной темой
							htmlDrEngine := web.NewDRW(gtx, cachedTheme, htmlString)
							return htmlDrEngine.RenderHTML()
						})
					})
				} else {
					// Показываем индикатор загрузки
					if cachedTheme == nil {
						cachedTheme = createScaledTheme(1.0)
					}
					layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Body1(cachedTheme, "Loading...").Layout(gtx)
					})
				}
				htmlMutex.RUnlock()

				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
