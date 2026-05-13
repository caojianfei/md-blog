package repository

import (
    "github.com/cybernote/md-blog/internal/model"
    "gorm.io/gorm"
)

type MediaRepository struct{ db *gorm.DB }

func NewMediaRepository(db *gorm.DB) *MediaRepository { return &MediaRepository{db: db} }

func (r *MediaRepository) List() ([]model.Media, error) {
    var items []model.Media
    err := r.db.Order("created_at DESC").Find(&items).Error
    return items, err
}

func (r *MediaRepository) Save(item *model.Media) error { return r.db.Save(item).Error }
func (r *MediaRepository) Delete(id uint) error         { return r.db.Delete(&model.Media{}, id).Error }
func (r *MediaRepository) FindByID(id uint) (*model.Media, error) {
    var item model.Media
    if err := r.db.First(&item, id).Error; err != nil {
        return nil, err
    }
    return &item, nil
}
