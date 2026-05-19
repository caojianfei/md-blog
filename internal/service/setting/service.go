package setting

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cybernote/md-blog/internal/config"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
)

type ResolvedStorage struct {
	Driver       string
	LocalPath    string
	LocalDirAbs  string
	LocalBaseURL string
	S3Endpoint   string
	S3Region     string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	S3UseSSL     bool
	S3PublicURL  string
}

type ResolvedAI struct {
	Enabled        bool
	Provider       string
	Model          string
	APIKey         string
	BaseURL        string
	TimeoutSeconds int
}

type ResolvedSettings struct {
	Site          *model.SiteSetting
	BaseURL       string
	PreviewSecret string
	MaxUploadSize int64
	Storage       ResolvedStorage
	AI            ResolvedAI
}

type Service struct {
	boot config.Config
	repo *repository.SettingRepository
}

func New(boot config.Config, repo *repository.SettingRepository) *Service {
	return &Service{boot: boot, repo: repo}
}

func (s *Service) Get() (*model.SiteSetting, error) {
	return s.repo.GetOrCreate(s.defaultSiteSetting())
}

func (s *Service) Resolve() (*ResolvedSettings, error) {
	site, err := s.Get()
	if err != nil {
		return nil, err
	}

	normalized := s.normalize(site)
	if err := s.Validate(normalized); err != nil {
		return nil, err
	}

	return &ResolvedSettings{
		Site:          normalized,
		BaseURL:       normalized.BaseURL,
		PreviewSecret: normalized.PreviewSecret,
		MaxUploadSize: normalized.MaxUploadSize,
		Storage: ResolvedStorage{
			Driver:       normalized.StorageDriver,
			LocalPath:    normalized.StorageLocalPath,
			LocalDirAbs:  filepath.Join(filepath.Clean(s.boot.App.DataDir), filepath.FromSlash(normalized.StorageLocalPath)),
			LocalBaseURL: normalized.StorageLocalBaseURL,
			S3Endpoint:   normalized.StorageS3Endpoint,
			S3Region:     normalized.StorageS3Region,
			S3Bucket:     normalized.StorageS3Bucket,
			S3AccessKey:  normalized.StorageS3AccessKey,
			S3SecretKey:  normalized.StorageS3SecretKey,
			S3UseSSL:     normalized.StorageS3UseSSL,
			S3PublicURL:  normalized.StorageS3PublicURL,
		},
		AI: ResolvedAI{
			Enabled:        normalized.AIEnabled,
			Provider:       normalized.AIProvider,
			Model:          normalized.AIModel,
			APIKey:         normalized.AIAPIKey,
			BaseURL:        normalized.AIBaseURL,
			TimeoutSeconds: normalized.AITimeoutSeconds,
		},
	}, nil
}

func (s *Service) Save(input *model.SiteSetting) (*model.SiteSetting, error) {
	current, err := s.Get()
	if err != nil {
		return nil, err
	}

	normalized := s.normalize(input)
	normalized.ID = current.ID
	normalized.CreatedAt = current.CreatedAt
	normalized.UpdatedAt = current.UpdatedAt
	normalized.DeletedAt = current.DeletedAt
	if err := s.Validate(normalized); err != nil {
		return nil, err
	}
	if err := s.repo.Save(normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func (s *Service) Validate(site *model.SiteSetting) error {
	if site == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if strings.TrimSpace(site.SiteName) == "" {
		return fmt.Errorf("siteName cannot be empty")
	}
	if strings.TrimSpace(site.BaseURL) == "" {
		return fmt.Errorf("baseUrl cannot be empty")
	}
	parsed, err := url.Parse(site.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("baseUrl must be a valid absolute URL")
	}
	if strings.TrimSpace(site.PreviewSecret) == "" {
		return fmt.Errorf("previewSecret cannot be empty")
	}
	if site.MaxUploadSize <= 0 {
		return fmt.Errorf("maxUploadSize must be greater than 0")
	}
	if site.StorageDriver != "local" && site.StorageDriver != "s3" {
		return fmt.Errorf("storageDriver must be local or s3")
	}
	if !isValidRelativePath(site.StorageLocalPath) {
		return fmt.Errorf("storageLocalPath must be a safe relative path")
	}
	if !isValidLocalBaseURL(site.StorageLocalBaseURL) {
		return fmt.Errorf("storageLocalBaseUrl must start with /uploads")
	}
	if site.StorageDriver == "s3" {
		if strings.TrimSpace(site.StorageS3Endpoint) == "" {
			return fmt.Errorf("storageS3Endpoint cannot be empty")
		}
		if strings.TrimSpace(site.StorageS3Bucket) == "" {
			return fmt.Errorf("storageS3Bucket cannot be empty")
		}
		if strings.TrimSpace(site.StorageS3AccessKey) == "" {
			return fmt.Errorf("storageS3AccessKey cannot be empty")
		}
		if strings.TrimSpace(site.StorageS3SecretKey) == "" {
			return fmt.Errorf("storageS3SecretKey cannot be empty")
		}
	}
	if !site.AIEnabled {
		return nil
	}
	if site.AIProvider != "openai_compatible" && site.AIProvider != "anthropic" && site.AIProvider != "gemini" {
		return fmt.Errorf("aiProvider must be openai_compatible, anthropic or gemini")
	}
	if strings.TrimSpace(site.AIModel) == "" {
		return fmt.Errorf("aiModel cannot be empty when ai is enabled")
	}
	if strings.TrimSpace(site.AIAPIKey) == "" {
		return fmt.Errorf("aiApiKey cannot be empty when ai is enabled")
	}
	if site.AITimeoutSeconds <= 0 {
		return fmt.Errorf("aiTimeoutSeconds must be greater than 0 when ai is enabled")
	}
	if site.AIProvider == "openai_compatible" && strings.TrimSpace(site.AIBaseURL) == "" {
		return fmt.Errorf("aiBaseUrl cannot be empty when aiProvider is openai_compatible")
	}
	return nil
}

func (s *Service) defaultSiteSetting() *model.SiteSetting {
	return s.normalize(&model.SiteSetting{
		SiteName:            "Cybernote Blog",
		SiteSubtitle:        "记录技术、思考与生活",
		SiteDescription:     "A modern personal blog built with Go SSR and Vue admin.",
		SiteKeywords:        "blog,golang,vue,markdown",
		AboutContent:        "## 关于我\n\n这里可以在后台维护你的关于页内容。",
		HeroTitle:           "你好，欢迎来到我的博客",
		HeroDescription:     "用 Go 打造的个人博客，支持 Markdown 创作与优雅展示。",
		ThemeDefault:        "system",
		FooterText:          "Built with Go SSR and Vue Admin",
		SearchPlaceholder:   "搜索标题、摘要或正文...",
		BaseURL:             "http://localhost:8080",
		PreviewSecret:       "preview-secret-change-me",
		MaxUploadSize:       8 << 20,
		StorageDriver:       "local",
		StorageLocalPath:    "uploads",
		StorageLocalBaseURL: "/uploads",
		StorageS3Region:     "us-east-1",
		StorageS3Bucket:     "md-blog",
		AIEnabled:           false,
		AIProvider:          "openai_compatible",
		AIBaseURL:           "https://api.openai.com/v1",
		AITimeoutSeconds:    15,
	})
}

func (s *Service) normalize(input *model.SiteSetting) *model.SiteSetting {
	if input == nil {
		input = &model.SiteSetting{}
	}
	normalized := *input
	normalized.SiteName = strings.TrimSpace(normalized.SiteName)
	normalized.SiteSubtitle = strings.TrimSpace(normalized.SiteSubtitle)
	normalized.SiteDescription = strings.TrimSpace(normalized.SiteDescription)
	normalized.SiteKeywords = strings.TrimSpace(normalized.SiteKeywords)
	normalized.HeroTitle = strings.TrimSpace(normalized.HeroTitle)
	normalized.HeroDescription = strings.TrimSpace(normalized.HeroDescription)
	normalized.ThemeDefault = strings.TrimSpace(normalized.ThemeDefault)
	normalized.FooterText = strings.TrimSpace(normalized.FooterText)
	normalized.ICP = strings.TrimSpace(normalized.ICP)
	normalized.GithubURL = strings.TrimSpace(normalized.GithubURL)
	normalized.Logo = strings.TrimSpace(normalized.Logo)
	normalized.DefaultOGImage = strings.TrimSpace(normalized.DefaultOGImage)
	normalized.SearchPlaceholder = strings.TrimSpace(normalized.SearchPlaceholder)
	normalized.BaseURL = trimTrailingSlash(normalized.BaseURL, "http://localhost:8080")
	normalized.PreviewSecret = strings.TrimSpace(normalized.PreviewSecret)
	normalized.StorageDriver = strings.ToLower(strings.TrimSpace(normalized.StorageDriver))
	normalized.StorageLocalPath = normalizeRelativePath(normalized.StorageLocalPath, "uploads")
	normalized.StorageLocalBaseURL = normalizeBasePath(normalized.StorageLocalBaseURL, "/uploads")
	normalized.StorageS3Endpoint = strings.TrimSpace(normalized.StorageS3Endpoint)
	normalized.StorageS3Region = strings.TrimSpace(normalized.StorageS3Region)
	normalized.StorageS3Bucket = strings.TrimSpace(normalized.StorageS3Bucket)
	normalized.StorageS3AccessKey = strings.TrimSpace(normalized.StorageS3AccessKey)
	normalized.StorageS3SecretKey = strings.TrimSpace(normalized.StorageS3SecretKey)
	normalized.StorageS3PublicURL = trimTrailingSlash(normalized.StorageS3PublicURL, "")
	normalized.AIProvider = strings.ToLower(strings.TrimSpace(normalized.AIProvider))
	normalized.AIModel = strings.TrimSpace(normalized.AIModel)
	normalized.AIAPIKey = strings.TrimSpace(normalized.AIAPIKey)
	normalized.AIBaseURL = trimTrailingSlash(normalized.AIBaseURL, "")
	if normalized.StorageS3PublicURL == "" {
		normalized.StorageS3PublicURL = trimTrailingSlash(normalized.StoragePublicURL, "")
	}
	normalized.StoragePublicURL = normalized.StorageS3PublicURL
	if normalized.SiteName == "" {
		normalized.SiteName = "Cybernote Blog"
	}
	if normalized.ThemeDefault == "" {
		normalized.ThemeDefault = "system"
	}
	if normalized.SearchPlaceholder == "" {
		normalized.SearchPlaceholder = "搜索标题、摘要或正文..."
	}
	if normalized.PreviewSecret == "" {
		normalized.PreviewSecret = "preview-secret-change-me"
	}
	if normalized.StorageDriver == "" {
		normalized.StorageDriver = "local"
	}
	if normalized.StorageS3Region == "" {
		normalized.StorageS3Region = "us-east-1"
	}
	if normalized.StorageS3Bucket == "" {
		normalized.StorageS3Bucket = "md-blog"
	}
	if normalized.AIProvider == "" {
		normalized.AIProvider = "openai_compatible"
	}
	if normalized.AIBaseURL == "" {
		normalized.AIBaseURL = "https://api.openai.com/v1"
	}
	if normalized.AITimeoutSeconds <= 0 {
		normalized.AITimeoutSeconds = 15
	}
	if normalized.MaxUploadSize <= 0 {
		normalized.MaxUploadSize = 8 << 20
	}
	return &normalized
}

func trimTrailingSlash(value, fallback string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func normalizeBasePath(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	trimmed = strings.TrimRight(trimmed, "/")
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func normalizeRelativePath(value, fallback string) string {
	trimmed := filepath.ToSlash(strings.TrimSpace(value))
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = pathClean(trimmed)
	if trimmed == "" || trimmed == "." || strings.HasPrefix(trimmed, "../") {
		return fallback
	}
	return trimmed
}

func isValidRelativePath(value string) bool {
	trimmed := normalizeRelativePath(value, "")
	return trimmed != ""
}

func isValidLocalBaseURL(value string) bool {
	return value == "/uploads" || strings.HasPrefix(value, "/uploads/")
}

func pathClean(value string) string {
	if value == "" {
		return ""
	}
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(value)), "./")
}
