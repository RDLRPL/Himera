package main

import (
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/RDLRPL/Himera/HDS/core/ffs"
	"github.com/RDLRPL/Himera/HDS/core/render/utils"
	"github.com/RDLRPL/Himera/HDS/core/render/utils/goutils"
	"github.com/RDLRPL/Himera/HDS/core/render/web"
)

func main() {
	go func() {
		dir := goutils.GetExecPath()

		fullPath := filepath.Join(dir, "tests", "index.html")

		fileFinder := ffs.NewFFST(fullPath)

		w := new(app.Window)
		w.Option(app.Title("Gio HTML Renderer"))

		th := material.NewTheme()
		th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				parseEng := utils.NewParse()
				htmlStr, err := parseEng.ParseHTML(fileFinder.Dir)
				if err != nil {
					log.Printf("Render error: %v", err)
				} else {
					htmlDrEngine := web.NewDRW(gtx, th, htmlStr)
					htmlDrEngine.RenderHTML()
				}

				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
