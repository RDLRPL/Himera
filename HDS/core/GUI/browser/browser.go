package browser

import (
	h "github.com/RDLRPL/Himera/HDS/core/http"
)

type Browser struct {
	resp     *h.Response
	maxZoom  float32
	minZoom  float32
	ZoomStep float32
}
