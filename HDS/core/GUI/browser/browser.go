package browser

import (
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	h "github.com/RDLRPL/Himera/HDS/core/http"
)

type Browser struct {
	resp     *h.Response
	maxZoom  float32
	minZoom  float32
	ZoomStep float32
}

func CreateScaledTheme(zoomFactor float32) *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	th.TextSize = unit.Sp(14 * zoomFactor)

	return th
}
