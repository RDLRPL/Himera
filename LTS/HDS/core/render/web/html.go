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

	// Кэшированные данные для оптимизации
	cachedDoc    *html.Node
	cachedLayout []layout.FlexChild
	bodyNode     *html.Node
	parsed       bool
}

func NewDRW(gtx layout.Context, th *material.Theme, htmlStr string) *DrawHTML {
	return &DrawHTML{
		gtxDH:     gtx,
		thDH:      th,
		htmlStrDH: htmlStr,
	}
}

// Кэшированный парсинг HTML
func (e *DrawHTML) ensureParsed() error {
	if e.parsed {
		return nil
	}

	if e.htmlStrDH == "" {
		e.parsed = true
		return nil
	}

	doc, err := html.Parse(strings.NewReader(e.htmlStrDH))
	if err != nil {
		return err
	}

	e.cachedDoc = doc
	e.bodyNode = findBodyNode(doc)
	e.parsed = true

	// Предварительная подготовка структуры для рендеринга
	e.prepareLayout()

	return nil
}

// Предварительная подготовка layout структуры
func (e *DrawHTML) prepareLayout() {
	if e.bodyNode != nil {
		e.cachedLayout = e.buildLayoutChildren(e.bodyNode)
	} else if e.cachedDoc != nil {
		e.cachedLayout = e.buildLayoutChildren(e.cachedDoc)
	}
}

// Создает layout children один раз вместо создания на каждом рендере
func (e *DrawHTML) buildLayoutChildren(node *html.Node) []layout.FlexChild {
	var children []layout.FlexChild

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			if !shouldSkipElement(c.Data) {
				child := c // Захватываем переменную
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return e.renderElement(gtx, child)
				}))
			}
		case html.TextNode:
			text := cleanText(c.Data)
			if text != "" && c.Parent != nil && !shouldSkipElement(c.Parent.Data) {
				textCopy := text // Захватываем переменную
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Body1(e.thDH, textCopy).Layout(gtx)
				}))
			}
		}
	}

	return children
}

// Получаем базовый размер отступа (без сложных вычислений)
func (e *DrawHTML) getInset(base float32) unit.Dp {
	return unit.Dp(base)
}

func (e *DrawHTML) RenderHTML() layout.Dimensions {
	if err := e.ensureParsed(); err != nil {
		return material.Body1(e.thDH, "Ошибка парсинга HTML: "+err.Error()).Layout(e.gtxDH)
	}

	if len(e.cachedLayout) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceSides,
	}.Layout(e.gtxDH, e.cachedLayout...)
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
	switch nodeName {
	case "head", "title", "meta", "link", "script", "style", "noscript", "comment":
		return true
	default:
		return false
	}
}

// Оптимизированная очистка текста
func cleanText(text string) string {
	if text == "" {
		return ""
	}

	// Быстрая проверка - если нет пробельных символов, возвращаем как есть
	hasWhitespace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			hasWhitespace = true
			break
		}
	}

	if !hasWhitespace {
		return text
	}

	// Удаляем лишние пробелы и переносы строк
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Заменяем множественные пробелы одним (оптимизированная версия)
	result := make([]rune, 0, len(text))
	prevSpace := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result = append(result, ' ')
				prevSpace = true
			}
		} else {
			result = append(result, r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(string(result))
}

var textCache = make(map[*html.Node]string)

func (e *DrawHTML) renderElement(gtx layout.Context, node *html.Node) layout.Dimensions {
	content := e.getCachedText(node)

	switch strings.ToLower(node.Data) {
	case "h1":
		if content != "" {
			h1 := material.H1(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(16)}.Layout(gtx, h1.Layout)
		}
		return e.layoutListOptimized(gtx, node)

	case "h2":
		if content != "" {
			h2 := material.H2(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(12), Top: e.getInset(8)}.Layout(gtx, h2.Layout)
		}
		return e.layoutListOptimized(gtx, node)

	case "h3":
		if content != "" {
			h3 := material.H3(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(8), Top: e.getInset(4)}.Layout(gtx, h3.Layout)
		}
		return e.layoutListOptimized(gtx, node)
	case "p":
		if content != "" {
			body := material.Body1(e.thDH, content)
			return layout.Inset{Bottom: e.getInset(12)}.Layout(gtx, body.Layout)
		}
		return layout.Inset{Bottom: e.getInset(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutListOptimized(gtx, node)
		})

	case "div":
		return layout.Inset{Bottom: e.getInset(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return e.layoutListOptimized(gtx, node)
		})

	case "span":
		if content != "" {
			return material.Body2(e.thDH, content).Layout(gtx)
		}
		return e.layoutListOptimized(gtx, node)

	case "a":
		if content != "" {
			link := material.Body1(e.thDH, content)
			return link.Layout(gtx)
		}
		return e.layoutListOptimized(gtx, node)

	case "br":
		return layout.Spacer{Height: e.getInset(8)}.Layout(gtx)

	case "hr":
		return layout.Spacer{Height: e.getInset(16)}.Layout(gtx)

	default:
		if content != "" && len(content) < 1000 {
			return material.Body1(e.thDH, content).Layout(gtx)
		}
		return e.layoutListOptimized(gtx, node)
	}
}

func (e *DrawHTML) getCachedText(node *html.Node) string {
	if cached, exists := textCache[node]; exists {
		return cached
	}

	result := extractTextOptimized(node)

	if len(textCache) > 1000 {
		textCache = make(map[*html.Node]string)
	}

	textCache[node] = result
	return result
}

func (e *DrawHTML) layoutListOptimized(gtx layout.Context, node *html.Node) layout.Dimensions {
	children := e.buildLayoutChildren(node)

	if len(children) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceSides,
	}.Layout(gtx, children...)
}

func extractTextOptimized(node *html.Node) string {
	var parts []string

	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := cleanText(n.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(node)

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return strings.Join(parts, " ")
}
