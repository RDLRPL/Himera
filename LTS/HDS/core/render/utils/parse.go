package utils

import (
	"io"
	"os"
	"sync"
)

type Parse struct {
	templates map[string]string
	funcMap   map[string]func(string) string
	mu        sync.Mutex
}

func NewParse() *Parse {
	return &Parse{
		templates: make(map[string]string),
		funcMap:   make(map[string]func(string) string),
	}
}

/*
	func (e *Parse) ParseHTML(name string) (string, error) {
		e.mu.Lock()
		content, _ := e.templates[name]
		e.mu.Unlock()

		for name, fn := range e.funcMap {
			placeholder := "{{" + name + " ."
			if strings.Contains(content, placeholder) {
				start := strings.Index(content, placeholder)
				if start != -1 {
					end := strings.Index(content[start:], "}}") + start
					if end != -1 {
						expr := content[start+len(placeholder) : end]
						result := fn(expr)
						content = content[:start] + result + content[end+2:]
					}
				}
			}
		}

		return content, nil
	}
*/
func (e *Parse) ParseHTML(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "ParseHTML: ", err
	}
	defer file.Close()

	htmlBytes, err := io.ReadAll(file)
	if err != nil {
		return "ParseHTML: ", err
	}

	htmlStr := string(htmlBytes)
	return htmlStr, nil
}
