package router

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	appcontainer "github.com/cybernote/md-blog/internal/container"
	apiHandler "github.com/cybernote/md-blog/internal/handler/api"
	webHandler "github.com/cybernote/md-blog/internal/handler/web"
	"github.com/cybernote/md-blog/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func New(c *appcontainer.Container) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)

	web := webHandler.New(c)
	api := apiHandler.New(c)

	assetServer := http.FileServer(http.FS(c.AssetFS))
	r.Handle("/assets/*", http.StripPrefix("/assets/", assetServer))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(c.Config.Storage.LocalDir))))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	r.Route("/api/admin", func(apiRouter chi.Router) {
		apiRouter.Post("/login", api.Login)
		apiRouter.Post("/logout", api.Logout)
		apiRouter.Get("/me", api.Me)
		apiRouter.Group(func(guarded chi.Router) {
			guarded.Use(middleware.AdminOnly(c.Auth))
			guarded.Get("/dashboard", api.Dashboard)
			guarded.Get("/articles", api.ListArticles)
			guarded.Post("/articles", api.SaveArticle)
			guarded.Get("/articles/{id}", api.GetArticle)
			guarded.Post("/articles/{id}/status", api.UpdateArticleStatus)
			guarded.Post("/articles/preview", api.PreviewMarkdown)
			guarded.Get("/categories", api.ListCategories)
			guarded.Post("/categories", api.SaveCategory)
			guarded.Delete("/categories/{id}", api.DeleteCategory)
			guarded.Get("/tags", api.ListTags)
			guarded.Post("/tags", api.SaveTag)
			guarded.Delete("/tags/{id}", api.DeleteTag)
			guarded.Get("/settings", api.GetSettings)
			guarded.Post("/settings", api.SaveSettings)
			guarded.Get("/media", api.ListMedia)
			guarded.Post("/media/upload", api.UploadMedia)
			guarded.Delete("/media/{id}", api.DeleteMedia)
		})
	})

	r.Get("/", web.Home)
	r.Get("/posts/{slug}", web.Article)
	r.Get("/categories/{slug}", web.Category)
	r.Get("/tags/{slug}", web.Tag)
	r.Get("/search", web.Search)
	r.Get("/archives", web.Archives)
	r.Get("/about", web.About)
	r.Get("/rss.xml", web.RSS)
	r.Get("/sitemap.xml", web.Sitemap)
	r.Get("/robots.txt", web.Robots)
	r.Get("/admin", serveAdmin(c.AdminFS))
	r.Get("/admin/*", serveAdmin(c.AdminFS))
	return r
}

func serveAdmin(adminFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentPath := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/admin"), "/")
		if currentPath == "" {
			currentPath = "index.html"
		}
		content, err := fs.ReadFile(adminFS, currentPath)
		if err != nil {
			if ext := path.Ext(currentPath); ext != "" {
				http.NotFound(w, r)
				return
			}
			content, err = fs.ReadFile(adminFS, "index.html")
			if err != nil {
				http.Error(w, "admin not built", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(content)
			return
		}
		w.Header().Set("Content-Type", detectContentType(currentPath))
		_, _ = w.Write(content)
	}
}

func detectContentType(name string) string {
	switch path.Ext(name) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
