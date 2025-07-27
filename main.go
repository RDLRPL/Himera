package main

import (
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/RDLRPL/Himera/HDS/core/render"
)

func main() {
	// Получаем путь к директории с шаблонами
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "templates")

	// Создаем движок
	engine := render.New(dir, true) // true - режим отладки

	// Добавляем пользовательские функции
	engine.AddFunc("uppercase", func(s string) string {
		return strings.ToUpper(s)
	})

	// Загружаем шаблоны
	if err := engine.LoadTemplates(); err != nil {
		panic(err)
	}

	// Роутинг
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title":   "Главная страница",
			"Message": "Привет, мир!",
		}
		engine.RenderHTTP(w, "index.html", data)
	})

	http.ListenAndServe(":8080", nil)
}
