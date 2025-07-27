package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/RDLRPL/Himera/HDS/core/render"
	"golang.org/x/net/html"
)

func main() {
	go func() {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Join(filepath.Dir(filename), "tests")

		engine := render.New(dir)
		engine.AddFunc("uppercase", strings.ToUpper)

		if err := engine.LoadTemplates(); err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}

		data := map[string]interface{}{
			"Title":   "Главная страница",
			"Message": "Привет, мир!",
		}

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
				// Создаем контекст для раскладки
				gtx := app.NewContext(&ops, e)

				htmlStr, err := engine.Render("index.html", data)
				if err != nil {
					log.Printf("Render error: %v", err)
				} else {
					renderHTML(gtx, th, htmlStr)
				}

				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}

// renderHTML parses the HTML into nodes and renders recursively
func renderHTML(gtx layout.Context, th *material.Theme, htmlStr string) layout.Dimensions {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return layout.Dimensions{}
	}
	return layoutList(gtx, th, doc)
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
