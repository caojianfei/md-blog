package setting

import (
	"testing"

	"github.com/cybernote/md-blog/internal/config"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetProvidesAIDefaults(t *testing.T) {
	service := newSettingTestService(t)

	setting, err := service.Get()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if setting.AIEnabled {
		t.Fatalf("expected ai to be disabled by default")
	}
	if setting.AIProvider != "openai_compatible" {
		t.Fatalf("expected default ai provider openai_compatible, got %q", setting.AIProvider)
	}
	if setting.AIBaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected default ai base url, got %q", setting.AIBaseURL)
	}
	if setting.AITimeoutSeconds != 15 {
		t.Fatalf("expected default ai timeout 15, got %d", setting.AITimeoutSeconds)
	}
}

func TestValidateAllowsEmptyAIFieldsWhenDisabled(t *testing.T) {
	service := newSettingTestService(t)

	err := service.Validate(&model.SiteSetting{
		SiteName:            "Test",
		BaseURL:             "https://example.com",
		PreviewSecret:       "preview-secret",
		MaxUploadSize:       1,
		StorageDriver:       "local",
		StorageLocalPath:    "uploads",
		StorageLocalBaseURL: "/uploads",
		AIEnabled:           false,
	})
	if err != nil {
		t.Fatalf("expected validation success when ai disabled, got %v", err)
	}
}

func TestValidateRequiresBaseURLForOpenAICompatible(t *testing.T) {
	service := newSettingTestService(t)

	err := service.Validate(&model.SiteSetting{
		SiteName:            "Test",
		BaseURL:             "https://example.com",
		PreviewSecret:       "preview-secret",
		MaxUploadSize:       1,
		StorageDriver:       "local",
		StorageLocalPath:    "uploads",
		StorageLocalBaseURL: "/uploads",
		AIEnabled:           true,
		AIProvider:          "openai_compatible",
		AIModel:             "gpt-4.1-mini",
		AIAPIKey:            "key",
		AITimeoutSeconds:    15,
	})
	if err == nil {
		t.Fatalf("expected validation error when openai compatible base url is empty")
	}
}

func TestValidateRequiresModelAndAPIKeyWhenAIEnabled(t *testing.T) {
	service := newSettingTestService(t)

	err := service.Validate(&model.SiteSetting{
		SiteName:            "Test",
		BaseURL:             "https://example.com",
		PreviewSecret:       "preview-secret",
		MaxUploadSize:       1,
		StorageDriver:       "local",
		StorageLocalPath:    "uploads",
		StorageLocalBaseURL: "/uploads",
		AIEnabled:           true,
		AIProvider:          "anthropic",
		AITimeoutSeconds:    15,
	})
	if err == nil {
		t.Fatalf("expected validation error when ai model and key are missing")
	}
}

func TestValidateRejectsUnknownProvider(t *testing.T) {
	service := newSettingTestService(t)

	err := service.Validate(&model.SiteSetting{
		SiteName:            "Test",
		BaseURL:             "https://example.com",
		PreviewSecret:       "preview-secret",
		MaxUploadSize:       1,
		StorageDriver:       "local",
		StorageLocalPath:    "uploads",
		StorageLocalBaseURL: "/uploads",
		AIEnabled:           true,
		AIProvider:          "unknown",
		AIModel:             "model",
		AIAPIKey:            "key",
		AITimeoutSeconds:    15,
	})
	if err == nil {
		t.Fatalf("expected validation error for unknown provider")
	}
}

func newSettingTestService(t *testing.T) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.SiteSetting{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return New(config.Config{
		App: config.AppConfig{
			DataDir: t.TempDir(),
		},
	}, repository.NewSettingRepository(db))
}
