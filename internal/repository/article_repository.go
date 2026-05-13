package repository

import (
    "net/url"
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
    if strings.TrimSpace(filter.Query) != "" {
        q := "%" + strings.TrimSpace(filter.Query) + "%"
        tx = tx.Where("title LIKE ? OR excerpt LIKE ? OR content LIKE ?", q, q, q)
    }

    var total int64
    if err := tx.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var articles []model.Article
    err := tx.Order("published_at DESC").Order("created_at DESC").
        Offset((filter.Page - 1) * filter.PageSize).
        Limit(filter.PageSize).
        Find(&articles).Error
    return articles, total, err
}

func (r *ArticleRepository) SetTags(article *model.Article, tags []model.Tag) error {
    return r.db.Model(article).Association("Tags").Replace(tags)
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
