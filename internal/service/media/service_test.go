package media

import (
	"strings"
	"testing"
)

func TestGenerateObjectKeyDoesNotContainOriginalFilename(t *testing.T) {
	key, err := generateObjectKey("My Weird 文件名 ###.PNG", "image/png")
	if err != nil {
		t.Fatalf("generateObjectKey returned error: %v", err)
	}
	if !strings.HasSuffix(key, ".png") {
		t.Fatalf("expected key to keep safe extension, got %q", key)
	}
	if strings.Contains(strings.ToLower(key), "weird") || strings.Contains(strings.ToLower(key), "my") {
		t.Fatalf("expected key to exclude original filename, got %q", key)
	}
}

func TestGenerateObjectKeyFallsBackToMIMEExtension(t *testing.T) {
	key, err := generateObjectKey("no-extension", "image/webp")
	if err != nil {
		t.Fatalf("generateObjectKey returned error: %v", err)
	}
	if !strings.HasSuffix(key, ".webp") {
		t.Fatalf("expected mime-based extension fallback, got %q", key)
	}
}

func TestSafeExtensionStripsUnsafeCharacters(t *testing.T) {
	ext := safeExtension("avatar.jp#g")
	if ext != ".jpg" {
		t.Fatalf("expected sanitized extension .jpg, got %q", ext)
	}
}
