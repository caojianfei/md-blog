package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer(templateFS fs.FS) (*Renderer, error) {
	funcs := template.FuncMap{
		"safe": func(value string) template.HTML { return template.HTML(value) },
		"formatDate": func(t *time.Time, layout string) string {
			if t == nil {
				return ""
			}
			return t.Format(layout)
		},
		"join": strings.Join,
		"pathEscape": func(value string) string {
			return url.PathEscape(value)
		},
		"split": strings.Split,
		"inc": func(v int) int { return v + 1 },
		"dec": func(v int) int {
			if v <= 1 {
				return 1
			}
			return v - 1
		},
		"tagWeight": func(count int64) string {
			if count >= 5 {
				return "high"
			}
			if count >= 2 {
				return "medium"
			}
			return "low"
		},
	}

	pageFiles, err := fs.Glob(templateFS, "web/templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	templates := make(map[string]*template.Template, len(pageFiles))
	for _, pageFile := range pageFiles {
		tpl, err := template.New("base").Funcs(funcs).ParseFS(templateFS,
			"web/templates/layouts/*.html",
			"web/templates/partials/*.html",
			pageFile,
		)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(pageFile), filepath.Ext(pageFile))
		templates[name] = tpl
	}
	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any, status int) {
	if status == 0 {
		status = http.StatusOK
	}
	tpl, ok := r.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("render error: template %q not found", name), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, fmt.Sprintf("render error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
