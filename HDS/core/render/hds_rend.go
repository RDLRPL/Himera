package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"sync"
)

// Engine - основной тип движка рендеринга
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	debug     bool
	dir       string
	mu        sync.Mutex
}

// New создает новый экземпляр движка
func New(dir string, debug bool) *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
		debug:     debug,
		dir:       dir,
	}
}

// LoadTemplates загружает все шаблоны из указанной директории
func (e *Engine) LoadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.debug {
		e.templates = make(map[string]*template.Template)
	}

	// Получаем все файлы шаблонов
	files, err := filepath.Glob(filepath.Join(e.dir, "*.html"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no templates found in %s", e.dir)
	}

	// Парсим все шаблоны
	for _, file := range files {
		name := filepath.Base(file)

		// Копируем карту функций для каждого шаблона
		funcMap := make(template.FuncMap)
		for k, v := range e.funcMap {
			funcMap[k] = v
		}

		tmpl := template.New(name).Funcs(funcMap)

		tmpl, err = tmpl.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %v", file, err)
		}

		e.templates[name] = tmpl
	}

	return nil
}

// AddFunc добавляет функцию в шаблоны
func (e *Engine) AddFunc(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcMap[name] = fn
}

// Render выполняет рендеринг шаблона
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	e.mu.Lock()
	tmpl, ok := e.templates[name]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	// В режиме отладки перезагружаем шаблоны перед каждым рендером
	if e.debug {
		if err := e.LoadTemplates(); err != nil {
			return err
		}
		e.mu.Lock()
		tmpl = e.templates[name]
		e.mu.Unlock()
	}

	return tmpl.Execute(w, data)
}

// RenderHTTP рендерит шаблон в http.ResponseWriter
func (e *Engine) RenderHTTP(w http.ResponseWriter, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := e.Render(buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}
