package repository

import (
    "github.com/cybernote/md-blog/internal/model"
    "gorm.io/gorm"
)

type AdminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) FindByUsername(username string) (*model.AdminUser, error) {
    var admin model.AdminUser
    if err := r.db.Where("username = ?", username).First(&admin).Error; err != nil {
        return nil, err
    }
    return &admin, nil
}

func (r *AdminRepository) FindByID(id uint) (*model.AdminUser, error) {
    var admin model.AdminUser
    if err := r.db.First(&admin, id).Error; err != nil {
        return nil, err
    }
    return &admin, nil
}

func (r *AdminRepository) Save(admin *model.AdminUser) error { return r.db.Save(admin).Error }
