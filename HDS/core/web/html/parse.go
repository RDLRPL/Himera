package html

import (
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(content string) *html.Node {
	doc, _ := html.Parse(strings.NewReader(content))

	return doc
}
