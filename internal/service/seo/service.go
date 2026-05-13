package seo

import (
    "fmt"
    "strings"

    "github.com/cybernote/md-blog/internal/config"
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
    cfg      config.Config
    settings *repository.SettingRepository
}

func New(cfg config.Config, settings *repository.SettingRepository) *Service {
    return &Service{cfg: cfg, settings: settings}
}

func (s *Service) Build(title, description, keywords, path string) Meta {
    siteTitle := s.cfg.SEO.DefaultTitle
    if strings.TrimSpace(title) != "" {
        siteTitle = fmt.Sprintf("%s | %s", title, s.cfg.SEO.DefaultTitle)
    }
    if strings.TrimSpace(description) == "" {
        description = s.cfg.SEO.DefaultDescription
    }
    if strings.TrimSpace(keywords) == "" {
        keywords = s.cfg.SEO.DefaultKeywords
    }
    return Meta{
        Title:       siteTitle,
        Description: description,
        Keywords:    keywords,
        Canonical:   strings.TrimRight(s.cfg.App.BaseURL, "/") + path,
        OGImage:     s.cfg.SEO.DefaultCover,
        Type:        "website",
    }
}
