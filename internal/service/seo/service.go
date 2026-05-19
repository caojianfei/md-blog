package seo

import (
	"fmt"
	"strings"

	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
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
	settings *settingSvc.Service
}

func New(settings *settingSvc.Service) *Service {
	return &Service{settings: settings}
}

func (s *Service) Build(title, description, keywords, path string) Meta {
	resolved, _ := s.settings.Resolve()
	if resolved == nil || resolved.Site == nil {
		return Meta{Title: title, Description: description, Keywords: keywords, Canonical: path, Type: "website"}
	}
	site := resolved.Site

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
		Canonical:   strings.TrimRight(resolved.BaseURL, "/") + path,
		OGImage:     site.DefaultOGImage,
		Type:        "website",
	}
}
