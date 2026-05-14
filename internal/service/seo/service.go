package seo

import (
	"fmt"
	"strings"

	"github.com/cybernote/md-blog/internal/repository"
)

type Meta struct {
	Title       string
	Description string
	Keywords    string
	Canonical   string
	OGImage     string
	Type        string
}

type Service struct {
	baseURL  string
	settings *repository.SettingRepository
}

func New(baseURL string, settings *repository.SettingRepository) *Service {
	return &Service{baseURL: baseURL, settings: settings}
}

func (s *Service) Build(title, description, keywords, path string) Meta {
	site, _ := s.settings.Get()

	siteTitle := site.SiteName
	if strings.TrimSpace(title) != "" {
		siteTitle = fmt.Sprintf("%s | %s", title, site.SiteName)
	}
	if strings.TrimSpace(description) == "" {
		description = site.SiteDescription
	}
	if strings.TrimSpace(keywords) == "" {
		keywords = site.SiteKeywords
	}
	return Meta{
		Title:       siteTitle,
		Description: description,
		Keywords:    keywords,
		Canonical:   strings.TrimRight(s.baseURL, "/") + path,
		OGImage:     site.DefaultOGImage,
		Type:        "website",
	}
}
