package render

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Engine struct {
	templates map[string]string
	funcMap   map[string]func(string) string
	dir       string
	mu        sync.Mutex
}

func New(dir string) *Engine {
	return &Engine{
		templates: make(map[string]string),
		funcMap:   make(map[string]func(string) string),
		dir:       dir,
	}
}

func (e *Engine) LoadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return filepath.Walk(e.dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			e.templates[info.Name()] = string(content)
		}

		return nil
	})
}

func (e *Engine) AddFunc(name string, fn func(string) string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcMap[name] = fn
}

func (e *Engine) Render(name string, data map[string]interface{}) (string, error) {
	e.mu.Lock()
	content, ok := e.templates[name]
	e.mu.Unlock()

	if !ok {
		return "", ErrTemplateNotFound
	}

	for k, v := range data {
		placeholder := "{{." + k + "}}"
		content = strings.ReplaceAll(content, placeholder, v.(string))
	}

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

var ErrTemplateNotFound = errors.New("template not found")
