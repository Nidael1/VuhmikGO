package delivery

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// pageData contiene los datos comunes de todas las páginas.
// El nombre de la app es configurable para soportar white label.
type pageData struct {
	ErrorCode    string
	ErrorMessage string
	UXErrors     []UXValidationError
	AppName      string
	PageTitle    string
	CurrentPath  string
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	render(w, "layout.html", pageData{
		AppName:     appName(),
		PageTitle:   "Inicio",
		CurrentPath: "/",
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	render(w, "layout.html", pageData{
		AppName:     appName(),
		PageTitle:   "Dashboard",
		CurrentPath: "/dashboard",
	})
}

func handlePacientes(w http.ResponseWriter, r *http.Request) {
	render(w, "layout.html", pageData{
		AppName:     appName(),
		PageTitle:   "Pacientes",
		CurrentPath: "/pacientes",
	})
}

func render(w http.ResponseWriter, tmpl string, data pageData) {
	path := filepath.Join("internal", "delivery", "http", "templates", tmpl)
	t, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "error al cargar plantilla", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "error al renderizar plantilla", http.StatusInternalServerError)
	}
}

// appName retorna el nombre de la aplicación.
// En producción se cargará desde configuración de tenant.
func appName() string {
	return "VUHMÍK"
}
