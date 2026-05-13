package repository

import (
    "github.com/cybernote/md-blog/internal/model"
    "gorm.io/gorm"
)

type CategoryRepository struct{ db *gorm.DB }
type TagRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }
func NewTagRepository(db *gorm.DB) *TagRepository { return &TagRepository{db: db} }

func (r *CategoryRepository) List() ([]model.Category, error) {
    var items []model.Category
    err := r.db.Order("sort ASC, created_at DESC").Find(&items).Error
    return items, err
}

func (r *CategoryRepository) Save(item *model.Category) error { return r.db.Save(item).Error }
func (r *CategoryRepository) Delete(id uint) error            { return r.db.Delete(&model.Category{}, id).Error }
func (r *CategoryRepository) FindByID(id uint) (*model.Category, error) {
    var item model.Category
    if err := r.db.First(&item, id).Error; err != nil {
        return nil, err
    }
    return &item, nil
}
func (r *CategoryRepository) FindBySlug(slug string) (*model.Category, error) {
    var item model.Category
    if err := r.db.Where("slug = ?", slug).First(&item).Error; err != nil {
        return nil, err
    }
    return &item, nil
}

func (r *TagRepository) List() ([]model.Tag, error) {
    var items []model.Tag
    err := r.db.Order("created_at DESC").Find(&items).Error
    return items, err
}

func (r *TagRepository) Save(item *model.Tag) error { return r.db.Save(item).Error }
func (r *TagRepository) Delete(id uint) error       { return r.db.Delete(&model.Tag{}, id).Error }
func (r *TagRepository) FindByIDs(ids []uint) ([]model.Tag, error) {
    if len(ids) == 0 {
        return []model.Tag{}, nil
    }
    var items []model.Tag
    err := r.db.Where("id IN ?", ids).Find(&items).Error
    return items, err
}
func (r *TagRepository) FindBySlug(slug string) (*model.Tag, error) {
    var item model.Tag
    if err := r.db.Where("slug = ?", slug).First(&item).Error; err != nil {
        return nil, err
    }
    return &item, nil
}
