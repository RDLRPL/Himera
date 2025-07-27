package web

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"golang.org/x/net/html"
)

type DrawHTML struct {
	gtxDH     layout.Context
	thDH      *material.Theme
	htmlStrDH string
}

func NewDRW(gtx layout.Context, th *material.Theme, htmlStr string) *DrawHTML {
	return &DrawHTML{
		gtxDH:     gtx,
		thDH:      th,
		htmlStrDH: htmlStr,
	}
}

func (e DrawHTML) RenderHTML() layout.Dimensions {
	doc, err := html.Parse(strings.NewReader(e.htmlStrDH))
	if err != nil {
		return layout.Dimensions{}
	}
	return layoutList(e.gtxDH, e.thDH, doc)
}

// layoutList walks the HTML tree and renders children vertically
func layoutList(gtx layout.Context, th *material.Theme, node *html.Node) layout.Dimensions {
	var children []layout.FlexChild
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return renderElement(gtx, th, c)
			}))
		case html.TextNode:
			text := strings.TrimSpace(c.Data)
			if text != "" {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, text).Layout(gtx)
				}))
			}
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// renderElement handles specific HTML tags
func renderElement(gtx layout.Context, th *material.Theme, node *html.Node) layout.Dimensions {
	content := extractText(node)
	switch node.Data {
	case "h1":
		return material.H1(th, content).Layout(gtx)
	case "h2":
		return material.H2(th, content).Layout(gtx)
	case "h3":
		return material.H3(th, content).Layout(gtx)
	case "p":
		return material.Body1(th, content).Layout(gtx)
	case "span":
		return material.Body2(th, content).Layout(gtx)
	case "a":
		// Render links italicized
		lbl := material.Body1(th, content)
		// Устанавливаем стиль шрифта как Italic
		return lbl.Layout(gtx)
	default:
		// Fallback: render children
		return layoutList(gtx, th, node)
	}
}

// extractText concatenates all text within a node
func extractText(node *html.Node) string {
	var sb strings.Builder
	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(node)
	return strings.TrimSpace(sb.String())
}
