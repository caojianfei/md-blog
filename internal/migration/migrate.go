package migration

import (
	"errors"

	"github.com/cybernote/md-blog/internal/config"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	"github.com/cybernote/md-blog/internal/security"
	"gorm.io/gorm"
)

func Run(db *gorm.DB, cfg config.Config) error {
	if err := db.AutoMigrate(
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleTag{},
		&model.SiteSetting{},
		&model.AdminUser{},
		&model.Media{},
	); err != nil {
		return err
	}

	articleRepo := repository.NewArticleRepository(db)
	if err := articleRepo.RefreshCategoryCounts(nil); err != nil {
		return err
	}
	if err := articleRepo.RefreshTagCounts(nil); err != nil {
		return err
	}

	if err := seedSiteSetting(db, cfg); err != nil {
		return err
	}
	return seedAdminUser(db, cfg)
}

func seedSiteSetting(db *gorm.DB, cfg config.Config) error {
	var setting model.SiteSetting
	err := db.First(&setting).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	setting = model.SiteSetting{
		SiteName:          cfg.App.Name,
		SiteSubtitle:      "记录技术、思考与生活",
		SiteDescription:   "A modern personal blog built with Go SSR and Vue admin.",
		SiteKeywords:      "blog,golang,vue,markdown",
		HeroTitle:         "你好，欢迎来到我的博客",
		HeroDescription:   "用 Go 打造的个人博客，支持 Markdown 创作与优雅展示。",
		ThemeDefault:      "system",
		FooterText:        "Built with Go SSR and Vue Admin",
		DefaultOGImage:    "",
		StorageDriver:     cfg.Storage.Driver,
		StoragePublicURL:  cfg.Storage.S3PublicURL,
		SearchPlaceholder: "搜索标题、摘要或正文...",
		AboutContent:      "## 关于我\n\n这里可以在后台维护你的关于页内容。",
	}
	return db.Create(&setting).Error
}

func seedAdminUser(db *gorm.DB, cfg config.Config) error {
	var admin model.AdminUser
	err := db.Where("username = ?", cfg.Admin.Username).First(&admin).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := security.HashPassword(cfg.Admin.Password)
	if err != nil {
		return err
	}

	admin = model.AdminUser{
		Username:      cfg.Admin.Username,
		PasswordHash:  hash,
		PasswordReset: false,
	}
	return db.Create(&admin).Error
}
