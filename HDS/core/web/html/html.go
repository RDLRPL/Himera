package html

import (
	"strings"
	"unicode"

	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"golang.org/x/net/html"
)

func NewHTMLRenderer(htmlContent string) *HTMLRenderer {
	return &HTMLRenderer{
		htmlContent: htmlContent,
		textCache:   make(map[*html.Node]string),
		layoutCache: make(map[*html.Node]*LayoutInfo),
		HTMLstyle:   HTMLcfgStyle,
	}
}

func (r *HTMLRenderer) ensureParsed() error {
	if r.parsed {
		return nil
	}

	if r.htmlContent == "" {
		r.parsed = true
		return nil
	}

	doc, err := html.Parse(strings.NewReader(r.htmlContent))
	if err != nil {
		return err
	}

	r.cachedDoc = doc
	r.bodyNode = findBodyNode(doc)
	r.parsed = true

	return nil
}

func (r *HTMLRenderer) Render(ctx *RenderContext) error {
	if err := r.ensureParsed(); err != nil {
		errorText := "HTML Parse Error: " + err.Error()
		TextLIB.DrawText(ctx.Program, errorText, ctx.X, ctx.Y, ctx.Zoom, [3]float32{1, 0, 0})
		return err
	}

	if r.bodyNode != nil {
		r.renderNode(ctx, r.bodyNode, ctx.X, ctx.Y)
	} else if r.cachedDoc != nil {
		r.renderNode(ctx, r.cachedDoc, ctx.X, ctx.Y)
	}

	return nil
}

func (r *HTMLRenderer) renderNode(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	currentY := y

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			if !shouldSkipElement(child.Data) {
				currentY = r.renderElement(ctx, child, x, currentY)
			}
		case html.TextNode:
			text := cleanText(child.Data)
			if text != "" && child.Parent != nil && !shouldSkipElement(child.Parent.Data) {
				currentY = r.renderText(ctx, text, x, currentY, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
			}
		}
	}

	return currentY
}

func (r *HTMLRenderer) renderElement(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	tag := strings.ToLower(node.Data)

	content := r.getCachedText(node)

	switch tag {
	case "h1":
		if content != "" {
			y += r.HTMLstyle.H1MarginTop * ctx.Zoom
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H1Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.H1MarginBottom * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "h2":
		if content != "" {
			y += r.HTMLstyle.H2MarginTop * ctx.Zoom
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H2Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.H2MarginBottom * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "h3":
		if content != "" {
			y += r.HTMLstyle.H3MarginTop * ctx.Zoom
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H3Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.H3MarginBottom * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "h4":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H4Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "h5":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H5Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "h6":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.H6Size, r.HTMLstyle.HeadingColor)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "p":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.renderNode(ctx, node, x, y)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		}

	case "div":
		y = r.renderNode(ctx, node, x, y)
		y += (r.HTMLstyle.ParagraphSpacing / 2) * ctx.Zoom

	case "span":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "a":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.LinkColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "strong", "b":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "em", "i":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "small":
		if content != "" {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.SmallSize, r.HTMLstyle.TextColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}

	case "br":
		lineHeight := float32(TextLIB.FontMetrics.Height>>6) * ctx.Zoom * r.HTMLstyle.LineSpacing
		y += lineHeight

	case "hr":
		y += 20 * ctx.Zoom

	case "ul", "ol":
		y = r.renderList(ctx, node, x, y, tag == "ol")

	case "li":
		y = r.renderListItem(ctx, node, x, y)

	case "blockquote":
		y = r.renderNode(ctx, node, x+r.HTMLstyle.IndentSize*ctx.Zoom, y)
		y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom

	default:
		if content != "" && len(content) < 1000 {
			y = r.renderText(ctx, content, x, y, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)
		} else {
			y = r.renderNode(ctx, node, x, y)
		}
	}

	return y
}

func (r *HTMLRenderer) renderText(ctx *RenderContext, text string, x, y, scale float32, color [3]float32) float32 {
	if text == "" {
		return y
	}

	effectiveScale := scale * ctx.Zoom
	lines := r.wrapText(text, ctx.Width-x, effectiveScale)
	lineHeight := float32(TextLIB.FontMetrics.Height>>6) * effectiveScale * r.HTMLstyle.LineSpacing

	currentY := y
	for _, line := range lines {
		if currentY+ctx.ScrollOffset > -lineHeight && currentY+ctx.ScrollOffset < ctx.Height+lineHeight {
			TextLIB.DrawText(ctx.Program, line, x, currentY+ctx.ScrollOffset, effectiveScale, color)
		}
		currentY += lineHeight

		if currentY+ctx.ScrollOffset > ctx.Height+lineHeight*10 {
			break
		}
	}

	return currentY
}

func (r *HTMLRenderer) renderList(ctx *RenderContext, node *html.Node, x, y float32, ordered bool) float32 {
	currentY := y
	itemNumber := 1

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && strings.ToLower(child.Data) == "li" {
			var prefix string
			if ordered {
				prefix = strings.Repeat(" ", 2) + string(rune('0'+itemNumber)) + ". "
				itemNumber++
			} else {
				prefix = "  • "
			}

			currentY = r.renderText(ctx, prefix, x, currentY, r.HTMLstyle.BaseSize, r.HTMLstyle.TextColor)

			itemY := currentY - float32(TextLIB.FontMetrics.Height>>6)*ctx.Zoom*r.HTMLstyle.LineSpacing
			currentY = r.renderNode(ctx, child, x+30*ctx.Zoom, itemY)
			currentY += 5 * ctx.Zoom
		}
	}

	return currentY + r.HTMLstyle.ParagraphSpacing*ctx.Zoom
}

func (r *HTMLRenderer) renderListItem(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	return r.renderNode(ctx, node, x, y)
}

func (r *HTMLRenderer) wrapText(text string, maxWidth, scale float32) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine strings.Builder

	for i, word := range words {
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		averageCharWidth := float32(TextLIB.FontMetrics.Height>>6) * 0.6
		estimatedWidth := float32(len(testLine)) * averageCharWidth * scale

		if estimatedWidth > maxWidth && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			if i > 0 && currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

func (r *HTMLRenderer) getCachedText(node *html.Node) string {
	if cached, exists := r.textCache[node]; exists {
		return cached
	}

	result := extractTextOptimized(node)

	if len(r.textCache) > 1000 {
		r.textCache = make(map[*html.Node]string)
	}

	r.textCache[node] = result
	return result
}

func (r *HTMLRenderer) CalculateContentHeight(ctx *RenderContext) float32 {
	if err := r.ensureParsed(); err != nil {
		return 100.0
	}

	tempCtx := *ctx
	tempCtx.ScrollOffset = 0

	var endY float32
	if r.bodyNode != nil {
		endY = r.calculateNodeHeight(&tempCtx, r.bodyNode, ctx.X, ctx.Y)
	} else if r.cachedDoc != nil {
		endY = r.calculateNodeHeight(&tempCtx, r.cachedDoc, ctx.X, ctx.Y)
	}

	return endY - ctx.Y
}

func (r *HTMLRenderer) calculateNodeHeight(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	currentY := y

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			if !shouldSkipElement(child.Data) {
				currentY = r.calculateElementHeight(ctx, child, x, currentY)
			}
		case html.TextNode:
			text := cleanText(child.Data)
			if text != "" && child.Parent != nil && !shouldSkipElement(child.Parent.Data) {
				parentTag := strings.ToLower(child.Parent.Data)
				if !r.isTextHandledByElement(parentTag) {
					currentY = r.calculateTextHeight(ctx, text, x, currentY, r.HTMLstyle.BaseSize)
				}
			}
		}
	}

	return currentY
}

func (r *HTMLRenderer) calculateElementHeight(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	tag := strings.ToLower(node.Data)
	content := r.getCachedText(node)

	switch tag {
	case "h1":
		if content != "" {
			y += r.HTMLstyle.H1MarginTop * ctx.Zoom
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H1Size)
			y += r.HTMLstyle.H1MarginBottom * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "h2":
		if content != "" {
			y += r.HTMLstyle.H2MarginTop * ctx.Zoom
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H2Size)
			y += r.HTMLstyle.H2MarginBottom * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "h3":
		if content != "" {
			y += r.HTMLstyle.H3MarginTop * ctx.Zoom
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H3Size)
			y += r.HTMLstyle.H3MarginBottom * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "h4":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H4Size)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "h5":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H5Size)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "h6":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.H6Size)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "p":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
			y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom
		}

	case "div":
		y = r.calculateNodeHeight(ctx, node, x, y)
		y += (r.HTMLstyle.ParagraphSpacing / 2) * ctx.Zoom

	case "span":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "a":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "strong", "b":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "em", "i":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "small":
		if content != "" {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.SmallSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}

	case "br":
		lineHeight := float32(TextLIB.FontMetrics.Height>>6) * ctx.Zoom * r.HTMLstyle.LineSpacing
		y += lineHeight

	case "hr":
		y += 20 * ctx.Zoom

	case "ul", "ol":
		y = r.calculateListHeight(ctx, node, x, y, tag == "ol")

	case "li":
		y = r.calculateListItemHeight(ctx, node, x, y)

	case "blockquote":
		y = r.calculateNodeHeight(ctx, node, x+r.HTMLstyle.IndentSize*ctx.Zoom, y)
		y += r.HTMLstyle.ParagraphSpacing * ctx.Zoom

	default:
		if content != "" && len(content) < 1000 {
			y = r.calculateTextHeight(ctx, content, x, y, r.HTMLstyle.BaseSize)
		} else {
			y = r.calculateNodeHeight(ctx, node, x, y)
		}
	}

	return y
}

func (r *HTMLRenderer) calculateTextHeight(ctx *RenderContext, text string, x, y, scale float32) float32 {
	if text == "" {
		return y
	}

	effectiveScale := scale * ctx.Zoom
	lines := r.wrapText(text, ctx.Width-x, effectiveScale)
	lineHeight := float32(TextLIB.FontMetrics.Height>>6) * effectiveScale * r.HTMLstyle.LineSpacing

	return y + float32(len(lines))*lineHeight
}

func (r *HTMLRenderer) calculateListHeight(ctx *RenderContext, node *html.Node, x, y float32, ordered bool) float32 {
	currentY := y
	itemNumber := 1

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && strings.ToLower(child.Data) == "li" {
			var prefix string
			if ordered {
				prefix = strings.Repeat(" ", 2) + string(rune('0'+itemNumber)) + ". "
				itemNumber++
			} else {
				prefix = "  • "
			}

			currentY = r.calculateTextHeight(ctx, prefix, x, currentY, r.HTMLstyle.BaseSize)

			itemY := currentY - float32(TextLIB.FontMetrics.Height>>6)*ctx.Zoom*r.HTMLstyle.LineSpacing
			currentY = r.calculateNodeHeight(ctx, child, x+30*ctx.Zoom, itemY)
			currentY += 5 * ctx.Zoom
		}
	}

	return currentY + r.HTMLstyle.ParagraphSpacing*ctx.Zoom
}

func (r *HTMLRenderer) calculateListItemHeight(ctx *RenderContext, node *html.Node, x, y float32) float32 {
	return r.calculateNodeHeight(ctx, node, x, y)
}

func (r *HTMLRenderer) isTextHandledByElement(tag string) bool {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6", "p", "span", "a", "strong", "b", "em", "i", "small":
		return true
	default:
		return false
	}
}

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

func shouldSkipElement(nodeName string) bool {
	switch nodeName {
	case "head", "title", "meta", "link", "script", "style", "noscript", "comment":
		return true
	default:
		return false
	}
}

func cleanText(text string) string {
	if text == "" {
		return ""
	}

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

	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

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
