package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
)

type Article struct {
	BaseModel
	Title          string        `json:"title" gorm:"size:255;not null"`
	Slug           string        `json:"slug" gorm:"size:255;uniqueIndex;not null"`
	Excerpt        string        `json:"excerpt" gorm:"size:500"`
	Content        string        `json:"content" gorm:"type:longtext;not null"`
	HTMLContent    string        `json:"htmlContent" gorm:"type:longtext;not null"`
	CoverImage     string        `json:"coverImage" gorm:"size:500"`
	Status         ArticleStatus `json:"status" gorm:"size:20;index;not null"`
	CategoryID     *uint         `json:"categoryId" gorm:"index"`
	Category       *Category     `json:"category,omitempty"`
	Tags           []Tag         `json:"tags,omitempty" gorm:"many2many:article_tags;constraint:OnDelete:CASCADE"`
	SEODescription string        `json:"seoDescription" gorm:"size:500"`
	SEOKeywords    string        `json:"seoKeywords" gorm:"size:255"`
	PublishedAt    *time.Time    `json:"publishedAt" gorm:"index"`
}

type Category struct {
	BaseModel
	Name         string `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Slug         string `json:"slug" gorm:"size:120;not null;uniqueIndex"`
	Description  string `json:"description" gorm:"size:500"`
	Sort         int    `json:"sort" gorm:"default:0"`
	ArticleCount int64  `json:"articleCount" gorm:"default:0"`
}

type Tag struct {
	BaseModel
	Name         string `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Slug         string `json:"slug" gorm:"size:120;not null;uniqueIndex"`
	Description  string `json:"description" gorm:"size:500"`
	ArticleCount int64  `json:"articleCount" gorm:"default:0"`
}

type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey"`
	TagID     uint `gorm:"primaryKey"`
}

type SiteSetting struct {
	BaseModel
	SiteName          string `json:"siteName" gorm:"size:100;not null"`
	SiteSubtitle      string `json:"siteSubtitle" gorm:"size:255"`
	SiteDescription   string `json:"siteDescription" gorm:"size:500"`
	SiteKeywords      string `json:"siteKeywords" gorm:"size:255"`
	AboutContent      string `json:"aboutContent" gorm:"type:longtext"`
	HeroTitle         string `json:"heroTitle" gorm:"size:100"`
	HeroDescription   string `json:"heroDescription" gorm:"size:500"`
	ThemeDefault      string `json:"themeDefault" gorm:"size:20"`
	FooterText        string `json:"footerText" gorm:"size:255"`
	ICP               string `json:"icp" gorm:"size:120"`
	GithubURL         string `json:"githubUrl" gorm:"size:255"`
	Logo              string `json:"logo" gorm:"size:500"`
	DefaultOGImage    string `json:"defaultOgImage" gorm:"size:500"`
	SearchPlaceholder string `json:"searchPlaceholder" gorm:"size:120"`
	BaseURL           string `json:"baseUrl" gorm:"size:255"`
	PreviewSecret     string `json:"previewSecret" gorm:"size:255"`
	MaxUploadSize     int64  `json:"maxUploadSize"`
	StorageDriver     string `json:"storageDriver" gorm:"size:20"`
	StorageLocalPath  string `json:"storageLocalPath" gorm:"size:255"`
	StorageLocalBaseURL string `json:"storageLocalBaseUrl" gorm:"size:255"`
	StorageS3Endpoint string `json:"storageS3Endpoint" gorm:"size:255"`
	StorageS3Region   string `json:"storageS3Region" gorm:"size:120"`
	StorageS3Bucket   string `json:"storageS3Bucket" gorm:"size:120"`
	StorageS3AccessKey string `json:"storageS3AccessKey" gorm:"size:255"`
	StorageS3SecretKey string `json:"storageS3SecretKey" gorm:"size:255"`
	StorageS3UseSSL   bool   `json:"storageS3UseSsl"`
	StorageS3PublicURL string `json:"storageS3PublicUrl" gorm:"size:255"`
	StoragePublicURL  string `json:"storagePublicUrl" gorm:"size:255"`
}

type AdminUser struct {
	BaseModel
	Username      string     `json:"username" gorm:"size:60;not null;uniqueIndex"`
	PasswordHash  string     `json:"-" gorm:"size:255;not null"`
	LastLoginAt   *time.Time `json:"lastLoginAt"`
	LastLoginIP   string     `json:"lastLoginIp" gorm:"size:64"`
	PasswordReset bool       `json:"passwordReset" gorm:"default:false"`
}

type Media struct {
	BaseModel
	Filename    string `json:"filename" gorm:"size:255;not null"`
	Original    string `json:"original" gorm:"size:255;not null"`
	MIMEType    string `json:"mimeType" gorm:"size:120"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	StorageType string `json:"storageType" gorm:"size:20;index"`
	ObjectKey   string `json:"objectKey" gorm:"size:500;not null"`
	URL         string `json:"url" gorm:"size:500;not null"`
}
