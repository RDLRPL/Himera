package browser

import (
	web "github.com/RDLxxx/Himera/HDS/core/html"
	"github.com/RDLxxx/Himera/HGD/utils"
)

var HTMLStyles = &web.StyleConfig{
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
