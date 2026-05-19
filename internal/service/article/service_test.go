package article

import (
	"fmt"
	"testing"

	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaveRefreshesPublishedCounts(t *testing.T) {
	db := openArticleTestDB(t)
	service, categoryRepo, tagRepo := newArticleTestService(db)

	categoryOne := createCategory(t, categoryRepo, "Go")
	categoryTwo := createCategory(t, categoryRepo, "Vue")
	tagOne := createTag(t, tagRepo, "backend")
	tagTwo := createTag(t, tagRepo, "frontend")

	article, err := service.Save(SaveInput{
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

	if _, err := service.Save(SaveInput{
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
	service, categoryRepo, tagRepo := newArticleTestService(db)

	category := createCategory(t, categoryRepo, "Ops")
	tag := createTag(t, tagRepo, "deploy")

	article, err := service.Save(SaveInput{
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

func openArticleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&model.Category{}, &model.Tag{}, &model.Article{}, &model.ArticleTag{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newArticleTestService(db *gorm.DB) (*Service, *repository.CategoryRepository, *repository.TagRepository) {
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	return New(articleRepo, categoryRepo, tagRepo, markdownSvc.New()), categoryRepo, tagRepo
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
