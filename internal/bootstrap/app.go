package bootstrap

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	mdblog "github.com/cybernote/md-blog"
	"github.com/cybernote/md-blog/internal/config"
	appcontainer "github.com/cybernote/md-blog/internal/container"
	"github.com/cybernote/md-blog/internal/migration"
	"github.com/cybernote/md-blog/internal/repository"
	"github.com/cybernote/md-blog/internal/router"
	articleSvc "github.com/cybernote/md-blog/internal/service/article"
	authSvc "github.com/cybernote/md-blog/internal/service/auth"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	mediaSvc "github.com/cybernote/md-blog/internal/service/media"
	seoSvc "github.com/cybernote/md-blog/internal/service/seo"
	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
	"github.com/cybernote/md-blog/internal/view"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	Container *appcontainer.Container
	server    *http.Server
}

func New(cfg config.Config) (*App, error) {
	if err := os.MkdirAll(cfg.App.DataDir, 0o755); err != nil {
		return nil, err
	}

	db, sqlDB, err := openDB(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Database.AutoMigrate {
		if err := migration.Run(db, cfg); err != nil {
			return nil, err
		}
	}

	renderer, err := view.NewRenderer(mdblog.TemplateFS)
	if err != nil {
		return nil, err
	}

	sessionStore := sessions.NewCookieStore([]byte(cfg.App.SessionSecret))
	sessionStore.Options = &sessions.Options{Path: "/", MaxAge: 86400 * 7, HttpOnly: true, SameSite: http.SameSiteLaxMode}

	templateFS, _ := fs.Sub(mdblog.TemplateFS, "web/templates")
	assetFS, _ := fs.Sub(mdblog.AssetFS, "web/assets")
	adminFS, _ := fs.Sub(mdblog.AdminFS, "web/admin/dist")

	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	settingsService := settingSvc.New(cfg, settingRepo)

	markdownService := markdownSvc.New()
	articleService := articleSvc.New(articleRepo, categoryRepo, tagRepo, markdownService, settingsService)
	authService := authSvc.New(adminRepo, sessionStore)
	resolvedSettings, err := settingsService.Resolve()
	if err != nil {
		return nil, err
	}
	if resolvedSettings.Storage.Driver == "local" {
		if err := os.MkdirAll(resolvedSettings.Storage.LocalDirAbs, 0o755); err != nil {
			return nil, err
		}
	}
	mediaService, err := mediaSvc.New(cfg, settingsService, mediaRepo)
	if err != nil {
		return nil, err
	}
	seoService := seoSvc.New(settingsService)

	container := &appcontainer.Container{Config: cfg, DB: db, SQLDB: sqlDB, Sessions: sessionStore, Renderer: renderer, TemplateFS: templateFS, AssetFS: assetFS, AdminFS: adminFS, ArticleRepo: articleRepo, CategoryRepo: categoryRepo, TagRepo: tagRepo, MediaRepo: mediaRepo, SettingRepo: settingRepo, AdminRepo: adminRepo, Settings: settingsService, Markdown: markdownService, Article: articleService, Auth: authService, Media: mediaService, SEO: seoService}
	mux := router.New(container)
	server := &http.Server{Addr: cfg.App.Addr, Handler: mux, ReadTimeout: time.Duration(cfg.App.ReadTimeout) * time.Second, WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second}
	return &App{Container: container, server: server}, nil
}

func (a *App) Start() error { return a.server.ListenAndServe() }

func openDB(cfg config.Config) (*gorm.DB, *sql.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.Database.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.SQLitePath), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.Database.MySQLDSN), &gorm.Config{})
	default:
		return nil, nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	return db, sqlDB, nil
}
