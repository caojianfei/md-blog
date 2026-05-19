package repository

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cybernote/md-blog/internal/model"
	"gorm.io/gorm"
)

type ArticleFilter struct {
	Query      string
	Status     string
	CategoryID uint
	TagID      uint
	Year       int
	Month      int
	Page       int
	PageSize   int
	OnlyPublic bool
}

type ArchiveItem struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

type DashboardStats struct {
	Articles      int64           `json:"articles"`
	Published     int64           `json:"published"`
	Drafts        int64           `json:"drafts"`
	Categories    int64           `json:"categories"`
	Tags          int64           `json:"tags"`
	RecentArticle []model.Article `json:"recentArticles"`
}

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Save(article *model.Article) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(article).Error
}

func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", id).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Article{}, id).Error
	})
}

func (r *ArticleRepository) FindByID(id uint) (*model.Article, error) {
	var article model.Article
	if err := r.db.Preload("Category").Preload("Tags").First(&article, id).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) FindBySlug(slug string) (*model.Article, error) {
	variants := map[string]struct{}{}
	slug = strings.TrimSpace(slug)
	if slug != "" {
		variants[slug] = struct{}{}
	}
	if escaped := url.PathEscape(slug); escaped != "" {
		variants[escaped] = struct{}{}
	}
	if unescaped, err := url.PathUnescape(slug); err == nil && unescaped != "" {
		variants[unescaped] = struct{}{}
	}

	candidates := make([]string, 0, len(variants))
	for value := range variants {
		candidates = append(candidates, value)
	}

	var article model.Article
	if err := r.db.Preload("Category").Preload("Tags").Where("slug IN ?", candidates).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) List(filter ArticleFilter) ([]model.Article, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	tx := r.db.Model(&model.Article{}).Preload("Category").Preload("Tags")
	if filter.OnlyPublic {
		tx = tx.Where("status = ?", model.ArticleStatusPublished)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}
	if filter.CategoryID > 0 {
		tx = tx.Where("category_id = ?", filter.CategoryID)
	}
	if filter.TagID > 0 {
		tx = tx.Joins("JOIN article_tags at ON at.article_id = articles.id").Where("at.tag_id = ?", filter.TagID)
	}
	if filter.Year > 0 && filter.Month > 0 {
		// Use sqlite-compatible or generic approach.
		// Actually gorm abstract this, but depending on db we might need string formatting.
		// Since it's sqlite, we can use strftime('%Y-%m', published_at) = 'YYYY-MM'
		// Or we can use range query: published_at >= startOfMonth AND published_at < startOfNextMonth
		startOfMonth := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.Local)
		startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
		tx = tx.Where("published_at >= ? AND published_at < ?", startOfMonth, startOfNextMonth)
	}
	if strings.TrimSpace(filter.Query) != "" {
		q := "%" + strings.TrimSpace(filter.Query) + "%"
		tx = tx.Where("title LIKE ? OR excerpt LIKE ? OR content LIKE ?", q, q, q)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []model.Article
	err := tx.Order("created_at DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepository) SetTags(article *model.Article, tags []model.Tag) error {
	return r.db.Model(article).Association("Tags").Replace(tags)
}

func (r *ArticleRepository) CountByCategory(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Article{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (r *ArticleRepository) CountByTag(tagID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Article{}).
		Joins("JOIN article_tags at ON at.article_id = articles.id").
		Where("at.tag_id = ?", tagID).
		Count(&count).Error
	return count, err
}

func (r *ArticleRepository) ClearCategory(categoryID uint) error {
	return r.db.Model(&model.Article{}).
		Where("category_id = ?", categoryID).
		Update("category_id", nil).Error
}

func (r *ArticleRepository) DetachTag(tagID uint) error {
	return r.db.Where("tag_id = ?", tagID).Delete(&model.ArticleTag{}).Error
}

func (r *ArticleRepository) RefreshCategoryCounts(ids []uint) error {
	return r.refreshCategoryCounts(r.db, ids)
}

func (r *ArticleRepository) RefreshTagCounts(ids []uint) error {
	return r.refreshTagCounts(r.db, ids)
}

func (r *ArticleRepository) DashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}
	if err := r.db.Model(&model.Article{}).Count(&stats.Articles).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&model.Article{}).Where("status = ?", model.ArticleStatusPublished).Count(&stats.Published).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&model.Article{}).Where("status = ?", model.ArticleStatusDraft).Count(&stats.Drafts).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&model.Category{}).Count(&stats.Categories).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&model.Tag{}).Count(&stats.Tags).Error; err != nil {
		return nil, err
	}
	if err := r.db.Preload("Category").Order("updated_at DESC").Limit(8).Find(&stats.RecentArticle).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *ArticleRepository) Archives() ([]ArchiveItem, error) {
	var articles []model.Article
	if err := r.db.Select("published_at").Where("status = ? AND published_at IS NOT NULL", model.ArticleStatusPublished).Order("published_at DESC").Find(&articles).Error; err != nil {
		return nil, err
	}
	grouped := make(map[string]*ArchiveItem)
	order := make([]string, 0)
	for _, article := range articles {
		if article.PublishedAt == nil {
			continue
		}
		t := *article.PublishedAt
		key := t.Format("2006-01")
		item, ok := grouped[key]
		if !ok {
			item = &ArchiveItem{Year: t.Year(), Month: int(t.Month()), Count: 0}
			grouped[key] = item
			order = append(order, key)
		}
		item.Count++
	}
	result := make([]ArchiveItem, 0, len(order))
	for _, key := range order {
		result = append(result, *grouped[key])
	}
	return result, nil
}

func (r *ArticleRepository) PrevNext(article *model.Article) (*model.Article, *model.Article, error) {
	var prev, next model.Article
	var prevPtr, nextPtr *model.Article
	if article.PublishedAt == nil {
		return nil, nil, nil
	}
	if err := r.db.Where("status = ? AND published_at < ?", model.ArticleStatusPublished, article.PublishedAt).Order("published_at DESC").First(&prev).Error; err == nil {
		prevPtr = &prev
	}
	if err := r.db.Where("status = ? AND published_at > ?", model.ArticleStatusPublished, article.PublishedAt).Order("published_at ASC").First(&next).Error; err == nil {
		nextPtr = &next
	}
	return prevPtr, nextPtr, nil
}

func PublishedAtForStatus(status model.ArticleStatus, previous *time.Time) *time.Time {
	if status == model.ArticleStatusPublished {
		if previous != nil {
			return previous
		}
		now := time.Now()
		return &now
	}
	return nil
}

func (r *ArticleRepository) refreshCategoryCounts(db *gorm.DB, ids []uint) error {
	targetIDs := uniqueUintIDs(ids)
	return db.Transaction(func(tx *gorm.DB) error {
		reset := tx.Model(&model.Category{})
		if len(targetIDs) > 0 {
			reset = reset.Where("id IN ?", targetIDs)
		} else {
			reset = reset.Where("id > 0")
		}
		if err := reset.Update("article_count", 0).Error; err != nil {
			return err
		}

		type categoryCountRow struct {
			CategoryID uint
			Count      int64
		}

		query := tx.Model(&model.Article{}).
			Select("category_id AS category_id, COUNT(*) AS count").
			Where("status = ? AND category_id IS NOT NULL", model.ArticleStatusPublished)
		if len(targetIDs) > 0 {
			query = query.Where("category_id IN ?", targetIDs)
		}

		var rows []categoryCountRow
		if err := query.Group("category_id").Scan(&rows).Error; err != nil {
			return err
		}

		for _, row := range rows {
			if err := tx.Model(&model.Category{}).
				Where("id = ?", row.CategoryID).
				Update("article_count", row.Count).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ArticleRepository) refreshTagCounts(db *gorm.DB, ids []uint) error {
	targetIDs := uniqueUintIDs(ids)
	return db.Transaction(func(tx *gorm.DB) error {
		reset := tx.Model(&model.Tag{})
		if len(targetIDs) > 0 {
			reset = reset.Where("id IN ?", targetIDs)
		} else {
			reset = reset.Where("id > 0")
		}
		if err := reset.Update("article_count", 0).Error; err != nil {
			return err
		}

		type tagCountRow struct {
			TagID uint
			Count int64
		}

		query := tx.Table("article_tags AS at").
			Select("at.tag_id AS tag_id, COUNT(*) AS count").
			Joins("JOIN articles ON articles.id = at.article_id").
			Where("articles.status = ? AND articles.deleted_at IS NULL", model.ArticleStatusPublished)
		if len(targetIDs) > 0 {
			query = query.Where("at.tag_id IN ?", targetIDs)
		}

		var rows []tagCountRow
		if err := query.Group("at.tag_id").Scan(&rows).Error; err != nil {
			return err
		}

		for _, row := range rows {
			if err := tx.Model(&model.Tag{}).
				Where("id = ?", row.TagID).
				Update("article_count", row.Count).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func uniqueUintIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
