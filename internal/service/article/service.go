package article

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	aiSvc "github.com/cybernote/md-blog/internal/service/ai"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
)

type Service struct {
	articles         *repository.ArticleRepository
	categories       *repository.CategoryRepository
	tags             *repository.TagRepository
	markdown         *markdownSvc.Service
	settings         *settingSvc.Service
	generatorFactory func(aiSvc.Config) aiSvc.MetadataGenerator
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

func New(articles *repository.ArticleRepository, categories *repository.CategoryRepository, tags *repository.TagRepository, markdown *markdownSvc.Service, settings *settingSvc.Service) *Service {
	return &Service{
		articles:         articles,
		categories:       categories,
		tags:             tags,
		markdown:         markdown,
		settings:         settings,
		generatorFactory: aiSvc.NewGenerator,
	}
}

func (s *Service) Save(ctx context.Context, input SaveInput) (*model.Article, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var article *model.Article
	var affectedCategoryIDs []uint
	var affectedTagIDs []uint
	if input.ID > 0 {
		existing, err := s.articles.FindByID(input.ID)
		if err != nil {
			return nil, err
		}
		article = existing
		affectedCategoryIDs = append(affectedCategoryIDs, valueOrZero(existing.CategoryID))
		affectedTagIDs = append(affectedTagIDs, tagIDsFromModels(existing.Tags)...)
	} else {
		article = &model.Article{}
	}

	rendered, err := s.markdown.Render(content)
	if err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		if input.ID > 0 {
			status = article.Status // 编辑时保留原状态，避免意外下线
		} else {
			status = model.ArticleStatusDraft
		}
	}

	article.Title = title
	article.Slug = slugify(input.Slug, title)
	article.Excerpt = strings.TrimSpace(input.Excerpt)
	article.Content = content
	article.HTMLContent = rendered.HTML
	article.CoverImage = strings.TrimSpace(input.CoverImage)
	article.CategoryID = input.CategoryID
	article.Category = nil
	article.Tags = nil
	article.SEODescription = strings.TrimSpace(input.SEODescription)
	article.SEOKeywords = strings.TrimSpace(input.SEOKeywords)
	s.fillMissingMetadata(ctx, title, content, &article.Excerpt, &article.SEOKeywords, &article.SEODescription)
	if article.Excerpt == "" {
		article.Excerpt = excerptFrom(content)
	}
	article.Status = status
	article.PublishedAt = repository.PublishedAtForStatus(status, article.PublishedAt)
	affectedCategoryIDs = append(affectedCategoryIDs, valueOrZero(input.CategoryID))
	affectedTagIDs = append(affectedTagIDs, input.TagIDs...)

	if err := s.articles.Save(article); err != nil {
		return nil, err
	}

	tags, tagErr := s.tags.FindByIDs(input.TagIDs)
	if tagErr != nil {
		return nil, tagErr
	}
	if err := s.articles.SetTags(article, tags); err != nil {
		return nil, err
	}

	if err := s.articles.RefreshCategoryCounts(affectedCategoryIDs); err != nil {
		return nil, err
	}
	if err := s.articles.RefreshTagCounts(affectedTagIDs); err != nil {
		return nil, err
	}
	return s.articles.FindByID(article.ID)
}

func (s *Service) fillMissingMetadata(ctx context.Context, title, content string, excerpt, seoKeywords, seoDescription *string) {
	if s.settings == nil || (strings.TrimSpace(*excerpt) != "" && strings.TrimSpace(*seoKeywords) != "" && strings.TrimSpace(*seoDescription) != "") {
		return
	}

	resolved, err := s.settings.Resolve()
	if err != nil {
		log.Printf("article ai metadata skipped: resolve settings failed: %v", err)
		return
	}

	generatorFactory := s.generatorFactory
	if generatorFactory == nil {
		generatorFactory = aiSvc.NewGenerator
	}

	result, err := generatorFactory(aiSvc.Config{
		Enabled:        resolved.AI.Enabled,
		Provider:       aiSvc.ProviderType(resolved.AI.Provider),
		Model:          resolved.AI.Model,
		APIKey:         resolved.AI.APIKey,
		BaseURL:        resolved.AI.BaseURL,
		TimeoutSeconds: resolved.AI.TimeoutSeconds,
	}).GenerateArticleMetadata(ctx, aiSvc.GenerateInput{
		Title:   title,
		Content: content,
	})
	if err != nil {
		log.Printf("article ai metadata skipped: generate failed: %v", err)
		return
	}
	result = aiSvc.SanitizeResult(result)
	if result == nil {
		return
	}

	if strings.TrimSpace(*excerpt) == "" {
		*excerpt = strings.TrimSpace(result.Excerpt)
	}
	if strings.TrimSpace(*seoKeywords) == "" {
		*seoKeywords = strings.TrimSpace(result.SEOKeywords)
	}
	if strings.TrimSpace(*seoDescription) == "" {
		*seoDescription = strings.TrimSpace(result.SEODescription)
	}
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
	tagIDs := tagIDsFromModels(article.Tags)
	categoryID := valueOrZero(article.CategoryID)
	article.Status = status
	article.PublishedAt = repository.PublishedAtForStatus(status, article.PublishedAt)
	article.Category = nil
	article.Tags = nil
	if err := s.articles.Save(article); err != nil {
		return nil, err
	}
	if err := s.articles.RefreshCategoryCounts([]uint{categoryID}); err != nil {
		return nil, err
	}
	if err := s.articles.RefreshTagCounts(tagIDs); err != nil {
		return nil, err
	}
	return s.articles.FindByID(id)
}

func (s *Service) Delete(id uint) error {
	article, err := s.articles.FindByID(id)
	if err != nil {
		return err
	}

	affectedCategoryIDs := []uint{valueOrZero(article.CategoryID)}
	affectedTagIDs := tagIDsFromModels(article.Tags)
	if err := s.articles.Delete(id); err != nil {
		return err
	}
	if err := s.articles.RefreshCategoryCounts(affectedCategoryIDs); err != nil {
		return err
	}
	return s.articles.RefreshTagCounts(affectedTagIDs)
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

func valueOrZero(id *uint) uint {
	if id == nil {
		return 0
	}
	return *id
}

func tagIDsFromModels(tags []model.Tag) []uint {
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}
