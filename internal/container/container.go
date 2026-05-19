package container

import (
	"database/sql"
	"io/fs"

	"github.com/cybernote/md-blog/internal/config"
	"github.com/cybernote/md-blog/internal/repository"
	articleSvc "github.com/cybernote/md-blog/internal/service/article"
	authSvc "github.com/cybernote/md-blog/internal/service/auth"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	mediaSvc "github.com/cybernote/md-blog/internal/service/media"
	seoSvc "github.com/cybernote/md-blog/internal/service/seo"
	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
	"github.com/cybernote/md-blog/internal/view"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

type Container struct {
	Config       config.Config
	DB           *gorm.DB
	SQLDB        *sql.DB
	Sessions     *sessions.CookieStore
	Renderer     *view.Renderer
	TemplateFS   fs.FS
	AssetFS      fs.FS
	AdminFS      fs.FS
	ArticleRepo  *repository.ArticleRepository
	CategoryRepo *repository.CategoryRepository
	TagRepo      *repository.TagRepository
	MediaRepo    *repository.MediaRepository
	SettingRepo  *repository.SettingRepository
	AdminRepo    *repository.AdminRepository
	Settings     *settingSvc.Service
	Markdown     *markdownSvc.Service
	Article      *articleSvc.Service
	Auth         *authSvc.Service
	Media        *mediaSvc.Service
	SEO          *seoSvc.Service
}
