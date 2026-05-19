package migration

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	defaults := defaultSiteSetting(cfg)

	err := db.First(&setting).Error
	if err == nil {
		if !mergeSiteSetting(&setting, defaults) {
			return nil
		}
		return db.Save(&setting).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	setting = defaults
	return db.Create(&setting).Error
}

func seedAdminUser(db *gorm.DB, cfg config.Config) error {
	var count int64
	if err := db.Model(&model.AdminUser{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := security.HashPassword(cfg.BootstrapAdmin.Password)
	if err != nil {
		return err
	}

	admin := model.AdminUser{
		Username:      cfg.BootstrapAdmin.Username,
		PasswordHash:  hash,
		PasswordReset: false,
	}
	return db.Create(&admin).Error
}

func defaultSiteSetting(cfg config.Config) model.SiteSetting {
	publicURL := strings.TrimRight(getSettingEnv("STORAGE_S3_PUBLIC_URL", ""), "/")
	return model.SiteSetting{
		SiteName:            getSettingEnv("APP_NAME", "Cybernote Blog"),
		SiteSubtitle:        "记录技术、思考与生活",
		SiteDescription:     "A modern personal blog built with Go SSR and Vue admin.",
		SiteKeywords:        "blog,golang,vue,markdown",
		AboutContent:        "## 关于我\n\n这里可以在后台维护你的关于页内容。",
		HeroTitle:           "你好，欢迎来到我的博客",
		HeroDescription:     "用 Go 打造的个人博客，支持 Markdown 创作与优雅展示。",
		ThemeDefault:        "system",
		FooterText:          "Built with Go SSR and Vue Admin",
		ICP:                 "",
		GithubURL:           "",
		Logo:                "",
		DefaultOGImage:      "",
		SearchPlaceholder:   "搜索标题、摘要或正文...",
		BaseURL:             strings.TrimRight(getSettingEnv("APP_BASE_URL", "http://localhost:8080"), "/"),
		PreviewSecret:       getSettingEnv("APP_PREVIEW_SECRET", generateSecret()),
		MaxUploadSize:       getSettingInt64Env("STORAGE_MAX_UPLOAD_SIZE", 8<<20),
		StorageDriver:       strings.ToLower(getSettingEnv("STORAGE_DRIVER", "local")),
		StorageLocalPath:    defaultLocalPath(cfg.App.DataDir, getSettingEnv("STORAGE_LOCAL_DIR", filepath.Join(cfg.App.DataDir, "uploads"))),
		StorageLocalBaseURL: trimRightOrDefault(getSettingEnv("STORAGE_LOCAL_BASE_URL", "/uploads"), "/uploads"),
		StorageS3Endpoint:   getSettingEnv("STORAGE_S3_ENDPOINT", "127.0.0.1:9000"),
		StorageS3Region:     getSettingEnv("STORAGE_S3_REGION", "us-east-1"),
		StorageS3Bucket:     getSettingEnv("STORAGE_S3_BUCKET", "md-blog"),
		StorageS3AccessKey:  getSettingEnv("STORAGE_S3_ACCESS_KEY", ""),
		StorageS3SecretKey:  getSettingEnv("STORAGE_S3_SECRET_KEY", ""),
		StorageS3UseSSL:     getSettingBoolEnv("STORAGE_S3_USE_SSL", false),
		StorageS3PublicURL:  publicURL,
		StoragePublicURL:    publicURL,
		AIEnabled:           false,
		AIProvider:          "openai_compatible",
		AIModel:             "",
		AIAPIKey:            "",
		AIBaseURL:           "https://api.openai.com/v1",
		AITimeoutSeconds:    15,
	}
}

func mergeSiteSetting(setting *model.SiteSetting, defaults model.SiteSetting) bool {
	changed := false
	applyString := func(target *string, fallback string) {
		if strings.TrimSpace(*target) == "" && strings.TrimSpace(fallback) != "" {
			*target = fallback
			changed = true
		}
	}

	applyString(&setting.SiteName, defaults.SiteName)
	applyString(&setting.SiteSubtitle, defaults.SiteSubtitle)
	applyString(&setting.SiteDescription, defaults.SiteDescription)
	applyString(&setting.SiteKeywords, defaults.SiteKeywords)
	applyString(&setting.AboutContent, defaults.AboutContent)
	applyString(&setting.HeroTitle, defaults.HeroTitle)
	applyString(&setting.HeroDescription, defaults.HeroDescription)
	applyString(&setting.ThemeDefault, defaults.ThemeDefault)
	applyString(&setting.FooterText, defaults.FooterText)
	applyString(&setting.SearchPlaceholder, defaults.SearchPlaceholder)
	applyString(&setting.BaseURL, defaults.BaseURL)
	applyString(&setting.PreviewSecret, defaults.PreviewSecret)
	applyString(&setting.StorageDriver, defaults.StorageDriver)
	applyString(&setting.StorageLocalPath, defaults.StorageLocalPath)
	applyString(&setting.StorageLocalBaseURL, defaults.StorageLocalBaseURL)
	applyString(&setting.StorageS3Endpoint, defaults.StorageS3Endpoint)
	applyString(&setting.StorageS3Region, defaults.StorageS3Region)
	applyString(&setting.StorageS3Bucket, defaults.StorageS3Bucket)
	applyString(&setting.StorageS3AccessKey, defaults.StorageS3AccessKey)
	applyString(&setting.StorageS3SecretKey, defaults.StorageS3SecretKey)
	if strings.TrimSpace(setting.StorageS3PublicURL) == "" && strings.TrimSpace(setting.StoragePublicURL) != "" {
		setting.StorageS3PublicURL = strings.TrimSpace(setting.StoragePublicURL)
		changed = true
	}
	applyString(&setting.StorageS3PublicURL, defaults.StorageS3PublicURL)
	if strings.TrimSpace(setting.StoragePublicURL) == "" && strings.TrimSpace(setting.StorageS3PublicURL) != "" {
		setting.StoragePublicURL = strings.TrimSpace(setting.StorageS3PublicURL)
		changed = true
	}
	applyString(&setting.AIProvider, defaults.AIProvider)
	applyString(&setting.AIBaseURL, defaults.AIBaseURL)
	if setting.AITimeoutSeconds <= 0 {
		setting.AITimeoutSeconds = defaults.AITimeoutSeconds
		changed = true
	}
	if setting.MaxUploadSize <= 0 {
		setting.MaxUploadSize = defaults.MaxUploadSize
		changed = true
	}
	return changed
}

func getSettingEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getSettingInt64Env(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getSettingBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func trimRightOrDefault(value, fallback string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func defaultLocalPath(dataDir, raw string) string {
	clean := filepath.Clean(strings.TrimSpace(raw))
	if clean == "." || clean == "" {
		return "uploads"
	}
	if rel, err := filepath.Rel(filepath.Clean(dataDir), clean); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	if filepath.IsAbs(clean) {
		return "uploads"
	}
	clean = strings.TrimPrefix(clean, "."+string(filepath.Separator))
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	if clean == "" {
		return "uploads"
	}
	return filepath.ToSlash(clean)
}

func generateSecret() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "preview-secret-change-me"
	}
	return hex.EncodeToString(buf)
}
