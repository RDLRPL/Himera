package render

import (
	"html/template"
	"path/filepath"
)

// Engine - основной тип движка рендеринга
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	debug     bool
	dir       string
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

func (e *Engine) LoadTemplates() error {
	if e.debug {
		e.templates = make(map[string]*template.Template)
	}

	layouts, err := filepath.Glob(filepath.Join(e.dir, "layouts/*.html"))
	if err != nil {
		return err
	}

	pages, err := filepath.Glob(filepath.Join(e.dir, "pages/*.html"))
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Копируем карту функций для каждого шаблона
		funcMap := make(template.FuncMap)
		for k, v := range e.funcMap {
			funcMap[k] = v
		}

		tmpl := template.New(name).Funcs(funcMap)

		// Парсим страницу вместе с лейаутами
		files := append(layouts, page)
		tmpl, err = tmpl.ParseFiles(files...)
		if err != nil {
			return err
		}

		e.templates[name] = tmpl
	}

	return nil
}

func (e *Engine) AddFunc(name string, fn interface{}) {
	e.funcMap[name] = fn
}
