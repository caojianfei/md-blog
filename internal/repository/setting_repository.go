package repository

import (
    "github.com/cybernote/md-blog/internal/model"
    "gorm.io/gorm"
)

type SettingRepository struct{ db *gorm.DB }

func NewSettingRepository(db *gorm.DB) *SettingRepository { return &SettingRepository{db: db} }

func (r *SettingRepository) Get() (*model.SiteSetting, error) {
    var setting model.SiteSetting
    if err := r.db.First(&setting).Error; err != nil {
        return nil, err
    }
    return &setting, nil
}

func (r *SettingRepository) Save(setting *model.SiteSetting) error { return r.db.Save(setting).Error }
