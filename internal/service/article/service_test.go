package article

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cybernote/md-blog/internal/config"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	aiSvc "github.com/cybernote/md-blog/internal/service/ai"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaveRefreshesPublishedCounts(t *testing.T) {
	db := openArticleTestDB(t)
	service, categoryRepo, tagRepo, _ := newArticleTestService(t, db)

	categoryOne := createCategory(t, categoryRepo, "Go")
	categoryTwo := createCategory(t, categoryRepo, "Vue")
	tagOne := createTag(t, tagRepo, "backend")
	tagTwo := createTag(t, tagRepo, "frontend")

	article, err := service.Save(context.Background(), SaveInput{
		Title:      "Draft Article",
		Content:    "# draft",
		CategoryID: &categoryOne.ID,
		TagIDs:     []uint{tagOne.ID, tagTwo.ID},
		Status:     model.ArticleStatusDraft,
	})
	if err != nil {
		t.Fatalf("save draft article: %v", err)
	}

	assertCategoryCount(t, categoryRepo, categoryOne.ID, 0)
	assertTagCount(t, tagRepo, tagOne.ID, 0)
	assertTagCount(t, tagRepo, tagTwo.ID, 0)

	if _, err := service.Save(context.Background(), SaveInput{
		ID:         article.ID,
		Title:      "Published Article",
		Content:    "# published",
		CategoryID: &categoryTwo.ID,
		TagIDs:     []uint{tagTwo.ID},
		Status:     model.ArticleStatusPublished,
	}); err != nil {
		t.Fatalf("publish article: %v", err)
	}

	assertCategoryCount(t, categoryRepo, categoryOne.ID, 0)
	assertCategoryCount(t, categoryRepo, categoryTwo.ID, 1)
	assertTagCount(t, tagRepo, tagOne.ID, 0)
	assertTagCount(t, tagRepo, tagTwo.ID, 1)
}

func TestDeleteRemovesTagRelationsAndRefreshesCounts(t *testing.T) {
	db := openArticleTestDB(t)
	service, categoryRepo, tagRepo, _ := newArticleTestService(t, db)

	category := createCategory(t, categoryRepo, "Ops")
	tag := createTag(t, tagRepo, "deploy")

	article, err := service.Save(context.Background(), SaveInput{
		Title:      "Published Article",
		Content:    "# hello",
		CategoryID: &category.ID,
		TagIDs:     []uint{tag.ID},
		Status:     model.ArticleStatusPublished,
	})
	if err != nil {
		t.Fatalf("save article: %v", err)
	}

	assertCategoryCount(t, categoryRepo, category.ID, 1)
	assertTagCount(t, tagRepo, tag.ID, 1)

	if err := service.Delete(article.ID); err != nil {
		t.Fatalf("delete article: %v", err)
	}

	assertCategoryCount(t, categoryRepo, category.ID, 0)
	assertTagCount(t, tagRepo, tag.ID, 0)

	var relationCount int64
	if err := db.Model(&model.ArticleTag{}).Where("article_id = ?", article.ID).Count(&relationCount).Error; err != nil {
		t.Fatalf("count article tags: %v", err)
	}
	if relationCount != 0 {
		t.Fatalf("expected article tag relations to be removed, got %d", relationCount)
	}
}

func TestSaveFillsMissingMetadataWithAI(t *testing.T) {
	db := openArticleTestDB(t)
	service, _, _, settingsService := newArticleTestService(t, db)
	enableAIForTest(t, settingsService)
	service.generatorFactory = func(aiSvc.Config) aiSvc.MetadataGenerator {
		return stubGenerator{
			result: &aiSvc.GenerateResult{
				Excerpt:        "AI 摘要",
				SEOKeywords:    "golang,ai,seo",
				SEODescription: "AI 生成的 SEO 描述",
			},
		}
	}

	article, err := service.Save(context.Background(), SaveInput{
		Title:   "AI Article",
		Content: "# hello ai",
	})
	if err != nil {
		t.Fatalf("save article with ai metadata: %v", err)
	}
	if article.Excerpt != "AI 摘要" {
		t.Fatalf("expected ai excerpt, got %q", article.Excerpt)
	}
	if article.SEOKeywords != "golang,ai,seo" {
		t.Fatalf("expected ai seo keywords, got %q", article.SEOKeywords)
	}
	if article.SEODescription != "AI 生成的 SEO 描述" {
		t.Fatalf("expected ai seo description, got %q", article.SEODescription)
	}
}

func TestSaveKeepsManualMetadataWhenAIEnabled(t *testing.T) {
	db := openArticleTestDB(t)
	service, _, _, settingsService := newArticleTestService(t, db)
	enableAIForTest(t, settingsService)
	service.generatorFactory = func(aiSvc.Config) aiSvc.MetadataGenerator {
		return stubGenerator{
			result: &aiSvc.GenerateResult{
				Excerpt:        "AI 摘要",
				SEOKeywords:    "ai,generated",
				SEODescription: "AI 生成描述",
			},
		}
	}

	article, err := service.Save(context.Background(), SaveInput{
		Title:          "Manual Metadata",
		Content:        "# content",
		Excerpt:        "手写摘要",
		SEOKeywords:    "manual,keywords",
		SEODescription: "手写 SEO 描述",
	})
	if err != nil {
		t.Fatalf("save article with manual metadata: %v", err)
	}
	if article.Excerpt != "手写摘要" {
		t.Fatalf("expected manual excerpt to be kept, got %q", article.Excerpt)
	}
	if article.SEOKeywords != "manual,keywords" {
		t.Fatalf("expected manual keywords to be kept, got %q", article.SEOKeywords)
	}
	if article.SEODescription != "手写 SEO 描述" {
		t.Fatalf("expected manual description to be kept, got %q", article.SEODescription)
	}
}

func TestSaveFallsBackWhenAIFails(t *testing.T) {
	db := openArticleTestDB(t)
	service, _, _, settingsService := newArticleTestService(t, db)
	enableAIForTest(t, settingsService)
	service.generatorFactory = func(aiSvc.Config) aiSvc.MetadataGenerator {
		return stubGenerator{err: fmt.Errorf("boom")}
	}

	content := "这是一段用于生成摘要的正文内容"
	article, err := service.Save(context.Background(), SaveInput{
		Title:   "AI Failure",
		Content: content,
	})
	if err != nil {
		t.Fatalf("save article when ai fails: %v", err)
	}
	if article.Excerpt != excerptFrom(content) {
		t.Fatalf("expected fallback excerpt, got %q", article.Excerpt)
	}
	if article.SEOKeywords != "" {
		t.Fatalf("expected empty seo keywords when ai fails, got %q", article.SEOKeywords)
	}
	if article.SEODescription != "" {
		t.Fatalf("expected empty seo description when ai fails, got %q", article.SEODescription)
	}
}

func TestSaveTrimsAndClampsAIMetadata(t *testing.T) {
	db := openArticleTestDB(t)
	service, _, _, settingsService := newArticleTestService(t, db)
	enableAIForTest(t, settingsService)
	service.generatorFactory = func(aiSvc.Config) aiSvc.MetadataGenerator {
		return stubGenerator{
			result: &aiSvc.GenerateResult{
				Excerpt:        strings.Repeat("摘", 600),
				SEODescription: strings.Repeat("描", 600),
				SEOKeywords:    strings.Repeat("关", 300),
			},
		}
	}

	article, err := service.Save(context.Background(), SaveInput{
		Title:   "Clamp Metadata",
		Content: "# content",
	})
	if err != nil {
		t.Fatalf("save article with long ai metadata: %v", err)
	}
	if len([]rune(article.Excerpt)) != 500 {
		t.Fatalf("expected excerpt length 500, got %d", len([]rune(article.Excerpt)))
	}
	if len([]rune(article.SEODescription)) != 500 {
		t.Fatalf("expected seo description length 500, got %d", len([]rune(article.SEODescription)))
	}
	if len([]rune(article.SEOKeywords)) != 255 {
		t.Fatalf("expected seo keywords length 255, got %d", len([]rune(article.SEOKeywords)))
	}
}

func openArticleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&model.Category{}, &model.Tag{}, &model.Article{}, &model.ArticleTag{}, &model.SiteSetting{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newArticleTestService(t *testing.T, db *gorm.DB) (*Service, *repository.CategoryRepository, *repository.TagRepository, *settingSvc.Service) {
	t.Helper()

	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	settingsService := settingSvc.New(config.Config{
		App: config.AppConfig{DataDir: t.TempDir()},
	}, repository.NewSettingRepository(db))
	return New(articleRepo, categoryRepo, tagRepo, markdownSvc.New(), settingsService), categoryRepo, tagRepo, settingsService
}

func enableAIForTest(t *testing.T, settingsService *settingSvc.Service) {
	t.Helper()

	setting, err := settingsService.Get()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	setting.AIEnabled = true
	setting.AIProvider = "openai_compatible"
	setting.AIModel = "test-model"
	setting.AIAPIKey = "test-key"
	setting.AIBaseURL = "https://example.com/v1"
	setting.AITimeoutSeconds = 15
	if _, err := settingsService.Save(setting); err != nil {
		t.Fatalf("save ai settings: %v", err)
	}
}

func createCategory(t *testing.T, repo *repository.CategoryRepository, name string) model.Category {
	t.Helper()

	item := model.Category{Name: name, Slug: slugify(name, name)}
	if err := repo.Save(&item); err != nil {
		t.Fatalf("save category %s: %v", name, err)
	}
	return item
}

func createTag(t *testing.T, repo *repository.TagRepository, name string) model.Tag {
	t.Helper()

	item := model.Tag{Name: name, Slug: slugify(name, name)}
	if err := repo.Save(&item); err != nil {
		t.Fatalf("save tag %s: %v", name, err)
	}
	return item
}

func assertCategoryCount(t *testing.T, repo *repository.CategoryRepository, id uint, expected int64) {
	t.Helper()

	item, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("find category %d: %v", id, err)
	}
	if item.ArticleCount != expected {
		t.Fatalf("expected category %d articleCount=%d, got %d", id, expected, item.ArticleCount)
	}
}

func assertTagCount(t *testing.T, repo *repository.TagRepository, id uint, expected int64) {
	t.Helper()

	items, err := repo.FindByIDs([]uint{id})
	if err != nil {
		t.Fatalf("find tag %d: %v", id, err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(items))
	}
	item := items[0]
	if item.ArticleCount != expected {
		t.Fatalf("expected tag %d articleCount=%d, got %d", id, expected, item.ArticleCount)
	}
}

type stubGenerator struct {
	result *aiSvc.GenerateResult
	err    error
}

func (g stubGenerator) GenerateArticleMetadata(context.Context, aiSvc.GenerateInput) (*aiSvc.GenerateResult, error) {
	if g.err != nil {
		return nil, g.err
	}
	if g.result == nil {
		return &aiSvc.GenerateResult{}, nil
	}
	return g.result, nil
}
