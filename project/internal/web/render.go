package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
)

type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{pages: make(map[string]*template.Template)}

	partials, err := template.ParseFS(templateFS, "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	r.partials = partials

	pageFiles, err := fs.Glob(templateFS, "templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("glob pages: %w", err)
	}
	for _, page := range pageFiles {
		p := path.Base(page)
		t, err := template.ParseFS(templateFS,
			"templates/layout.html",
			"templates/partials/*.html",
			page,
		)
		if err != nil {
			return nil, fmt.Errorf("parse page %q: %w", p, err)
		}
		r.pages[p] = t
	}
	return r, nil
}

func (r *Renderer) Page(w http.ResponseWriter, name string, data any) {
	t, ok := r.pages[name]
	if !ok {
		http.Error(w, "page not found: "+name, http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *Renderer) Partial(w http.ResponseWriter, name string, data any) {
	if err := r.partials.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
