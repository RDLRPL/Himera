package himera

import (
	web "github.com/RDLxxx/Himera/HDS/core/html"
	h "github.com/RDLxxx/Himera/HDS/core/http"
	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/RDLxxx/Himera/HGD/utils"
)

func RenderHTML(program uint32) {
	if core.Browse.HtmlRenderer == nil {
		return
	}

	availableHeight := float32(core.Browse.CurrentHeight) - core.Browse.InputBoxHeight - 20.0
	ctx := &web.RenderContext{
		Program:      program,
		X:            10.0 * core.Browse.Zoom,
		Y:            core.Browse.InputBoxHeight + 15.0*core.Browse.Zoom,
		Width:        float32(core.Browse.CurrentWidth) - 20.0*core.Browse.Zoom,
		Height:       availableHeight,
		ScrollOffset: core.Browse.ScrollOffset,
		Zoom:         core.Browse.Zoom,
	}

	if err := core.Browse.HtmlRenderer.Render(ctx); err != nil {
		TextLIB.DrawText(program, "HTML Render Error: "+err.Error(),
			10.0*core.Browse.Zoom, core.Browse.InputBoxHeight+15.0*core.Browse.Zoom, core.Browse.Zoom, utils.RGBToFloat32(255, 100, 100))
	}
}

func UpdateContent(link string, ua string) web.HTMLRenderer {
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
		core.Browse.HtmlRenderer = web.NewHTMLRenderer(errorHTML)
	} else {
		core.Browse.HtmlRenderer = web.NewHTMLRenderer(req.Page)
	}
	return *core.Browse.HtmlRenderer
}
