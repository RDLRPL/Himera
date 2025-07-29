package web

import (
	"strings"
	"unicode"

	"gioui.org/layout"
	"gioui.org/unit"
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

// Получаем базовый размер отступа (без сложных вычислений)
func (e *DrawHTML) getInset(base float32) unit.Dp {
	return unit.Dp(base)
}

func (e DrawHTML) RenderHTML() layout.Dimensions {
	if e.htmlStrDH == "" {
		return layout.Dimensions{}
	}

	doc, err := html.Parse(strings.NewReader(e.htmlStrDH))
	if err != nil {
		// Показываем ошибку парсинга
		return material.Body1(e.thDH, "Ошибка парсинга HTML: "+err.Error()).Layout(e.gtxDH)
	}

	// Ищем body элемент для лучшего рендеринга
	bodyNode := findBodyNode(doc)
	if bodyNode != nil {
		return e.layoutList(e.gtxDH, bodyNode)
	}

	return e.layoutList(e.gtxDH, doc)
}

// Находит body элемент в HTML документе
func findBodyNode(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "body" {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if result := findBodyNode(c); result != nil {
			return result
		}
	}

	return nil
}

// Проверяет, является ли элемент скрытым или служебным
func shouldSkipElement(nodeName string) bool {
	skipElements := map[string]bool{
		"head":     true,
		"title":    true,
		"meta":     true,
		"link":     true,
		"script":   true,
		"style":    true,
		"noscript": true,
		"comment":  true,
	}
	return skipElements[nodeName]
}

// Очищает и нормализует текст
func cleanText(text string) string {
	// Удаляем лишние пробелы и переносы строк
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Заменяем множественные пробелы одним
	var result strings.Builder
	var prevSpace bool

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

func (e *DrawHTML) layoutList(gtx layout.Context, node *html.Node) layout.Dimensions {
	var children []layout.FlexChild

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			if !shouldSkipElement(c.Data) {
				child := c // Захватываем переменную для замыкания
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return e.renderElement(gtx, child)
				}))
			}
		case html.TextNode:
			text := cleanText(c.Data)
			if text != "" && c.Parent != nil && !shouldSkipElement(c.Parent.Data) {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Body1(e.thDH, text).Layout(gtx)
				}))
			}
		}
	}

	if len(children) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceSides,
	}.Layout(gtx, children...)
}

func (e *DrawHTML) renderElement(gtx layout.Context, node *html.Node) layout.Dimensions {
	// Получаем текстовое содержимое элемента
	content := extractText(node)

	switch strings.ToLower(node.Data) {
	case "h1":
		if content != "" {
			h1 := material.H1(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(16)}.Layout(gtx, h1.Layout)
		}
		return e.layoutList(gtx, node)

	case "h2":
		if content != "" {
			h2 := material.H2(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(12), Top: e.getInset(8)}.Layout(gtx, h2.Layout)
		}
		return e.layoutList(gtx, node)

	case "h3":
		if content != "" {
			h3 := material.H3(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(8), Top: e.getInset(4)}.Layout(gtx, h3.Layout)
		}
		return e.layoutList(gtx, node)

	case "h4":
		if content != "" {
			h4 := material.H4(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(6), Top: e.getInset(2)}.Layout(gtx, h4.Layout)
		}
		return e.layoutList(gtx, node)

	case "h5":
		if content != "" {
			h5 := material.H5(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(4)}.Layout(gtx, h5.Layout)
		}
		return e.layoutList(gtx, node)

	case "h6":
		if content != "" {
			h6 := material.H6(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(4)}.Layout(gtx, h6.Layout)
		}
		return e.layoutList(gtx, node)

	case "p":
		if content != "" {
			body := material.Body1(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(12)}.Layout(gtx, body.Layout)
		}
		return layout.Inset{Bottom: e.getInset(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutList(gtx, node)
		})

	case "div":
		return layout.Inset{Bottom: e.getInset(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutList(gtx, node)
		})

	case "span":
		if content != "" {
			return material.Body2(e.thDH, content).Layout(gtx)
		}
		return e.layoutList(gtx, node)

	case "a":
		if content != "" {
			link := material.Body1(e.thDH, content)
			// Можно добавить другой цвет для ссылок
			return link.Layout(gtx)
		}
		return e.layoutList(gtx, node)

	case "ul", "ol":
		return layout.Inset{Bottom: e.getInset(8), Left: e.getInset(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutList(gtx, node)
		})

	case "li":
		if content != "" {
			listItem := material.Body1(e.thDH, "• "+content)
			return layout.Inset{Bottom: e.getInset(4)}.Layout(gtx, listItem.Layout)
		}
		return layout.Inset{Left: e.getInset(16), Bottom: e.getInset(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutList(gtx, node)
		})

	case "br":
		return layout.Spacer{Height: e.getInset(8)}.Layout(gtx)

	case "hr":
		return layout.Spacer{Height: e.getInset(16)}.Layout(gtx)

	case "blockquote":
		return layout.Inset{
			Left:   e.getInset(16),
			Right:  e.getInset(16),
			Bottom: e.getInset(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if content != "" {
				quote := material.Body2(e.thDH, content)
				return quote.Layout(gtx)
			}
			return e.layoutList(gtx, node)
		})

	case "pre", "code":
		if content != "" {
			code := material.Body2(e.thDH, content)
			return layout.Inset{
				Top:    e.getInset(8),
				Bottom: e.getInset(8),
				Left:   e.getInset(12),
				Right:  e.getInset(12),
			}.Layout(gtx, code.Layout)
		}
		return e.layoutList(gtx, node)

	default:
		// Для неизвестных элементов просто рендерим их содержимое
		if content != "" && len(content) < 1000 { // Ограничиваем длинный текст
			return material.Body1(e.thDH, content).Layout(gtx)
		}
		return e.layoutList(gtx, node)
	}
}

func extractText(node *html.Node) string {
	var sb strings.Builder
	var walker func(*html.Node)

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := cleanText(n.Data)
			if text != "" {
				if sb.Len() > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(node)
	result := strings.TrimSpace(sb.String())

	return result
}
