package html

import "golang.org/x/net/html"

type HTMLConfig struct {
	TextColor    [3]float32
	LinkColor    [3]float32
	HeadingColor [3]float32

	H1Size    float32
	H2Size    float32
	H3Size    float32
	H4Size    float32
	H5Size    float32
	H6Size    float32
	BaseSize  float32
	SmallSize float32

	ParagraphSpacing float32
	LineSpacing      float32
	IndentSize       float32

	H1MarginTop    float32
	H1MarginBottom float32
	H2MarginTop    float32
	H2MarginBottom float32
	H3MarginTop    float32
	H3MarginBottom float32
}

type LayoutInfo struct {
	X, Y          float32
	Width, Height float32
	LineHeight    float32
}

type RenderContext struct {
	Program      uint32
	X, Y         float32
	Width        float32
	Height       float32
	ScrollOffset float32
	Zoom         float32
}

type HTMLRenderer struct {
	htmlContent string
	cachedDoc   *html.Node
	bodyNode    *html.Node
	parsed      bool

	HTMLstyle *HTMLConfig

	textCache   map[*html.Node]string
	layoutCache map[*html.Node]*LayoutInfo
}
