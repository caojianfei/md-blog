package article

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
)

type Service struct {
	articles   *repository.ArticleRepository
	categories *repository.CategoryRepository
	tags       *repository.TagRepository
	markdown   *markdownSvc.Service
}

type SaveInput struct {
	ID             uint                `json:"id"`
	Title          string              `json:"title"`
	Slug           string              `json:"slug"`
	Excerpt        string              `json:"excerpt"`
	Content        string              `json:"content"`
	CoverImage     string              `json:"coverImage"`
	CategoryID     *uint               `json:"categoryId"`
	TagIDs         []uint              `json:"tagIds"`
	Status         model.ArticleStatus `json:"status"`
	SEODescription string              `json:"seoDescription"`
	SEOKeywords    string              `json:"seoKeywords"`
}

func New(articles *repository.ArticleRepository, categories *repository.CategoryRepository, tags *repository.TagRepository, markdown *markdownSvc.Service) *Service {
	return &Service{articles: articles, categories: categories, tags: tags, markdown: markdown}
}

func (s *Service) Save(input SaveInput) (*model.Article, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var article *model.Article
	if input.ID > 0 {
		existing, err := s.articles.FindByID(input.ID)
		if err != nil {
			return nil, err
		}
		article = existing
	} else {
		article = &model.Article{}
	}

	rendered, err := s.markdown.Render(content)
	if err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		status = model.ArticleStatusDraft
	}

	article.Title = title
	article.Slug = slugify(input.Slug, title)
	article.Excerpt = strings.TrimSpace(input.Excerpt)
	if article.Excerpt == "" {
		article.Excerpt = excerptFrom(content)
	}
	article.Content = content
	article.HTMLContent = rendered.HTML
	article.CoverImage = strings.TrimSpace(input.CoverImage)
	article.CategoryID = input.CategoryID
	article.SEODescription = strings.TrimSpace(input.SEODescription)
	article.SEOKeywords = strings.TrimSpace(input.SEOKeywords)
	article.Status = status
	article.PublishedAt = repository.PublishedAtForStatus(status, article.PublishedAt)

	if err := s.articles.Save(article); err != nil {
		return nil, err
	}

	tags, err := s.tags.FindByIDs(input.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := s.articles.SetTags(article, tags); err != nil {
		return nil, err
	}
	return s.articles.FindByID(article.ID)
}

func (s *Service) FindByID(id uint) (*model.Article, error)       { return s.articles.FindByID(id) }
func (s *Service) FindBySlug(slug string) (*model.Article, error) { return s.articles.FindBySlug(slug) }
func (s *Service) List(filter repository.ArticleFilter) ([]model.Article, int64, error) {
	return s.articles.List(filter)
}
func (s *Service) DashboardStats() (*repository.DashboardStats, error) {
	return s.articles.DashboardStats()
}
func (s *Service) Archives() ([]repository.ArchiveItem, error) { return s.articles.Archives() }
func (s *Service) PrevNext(article *model.Article) (*model.Article, *model.Article, error) {
	return s.articles.PrevNext(article)
}
func (s *Service) Preview(content string) (*markdownSvc.RenderResult, error) {
	return s.markdown.Render(content)
}

func (s *Service) UpdateStatus(id uint, status model.ArticleStatus) (*model.Article, error) {
	article, err := s.articles.FindByID(id)
	if err != nil {
		return nil, err
	}
	article.Status = status
	article.PublishedAt = repository.PublishedAtForStatus(status, article.PublishedAt)
	if err := s.articles.Save(article); err != nil {
		return nil, err
	}
	return s.articles.FindByID(id)
}

func slugify(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	if decoded, err := url.PathUnescape(value); err == nil {
		value = decoded
	}
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", ",", "", "，", "", ":", "", "：", "", "?", "", "？", "", "!", "", "！", "", "#", "", "@", "", "(", "", ")", "", "[", "", "]", "")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "post"
	}
	return value
}

func excerptFrom(content string) string {
	compact := strings.ReplaceAll(content, "\n", " ")
	compact = strings.TrimSpace(compact)
	runes := []rune(compact)
	if len(runes) > 140 {
		return string(runes[:140]) + "..."
	}
	return compact
}
