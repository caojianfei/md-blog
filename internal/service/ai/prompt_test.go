package ai

import (
	"strings"
	"testing"
)

func TestParseResponseParsesJSONCodeFence(t *testing.T) {
	result, err := parseResponse("```json\n{\"excerpt\":\"摘要\",\"seo_keywords\":\"go，ai\",\"seo_description\":\"描述\"}\n```")
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if result.Excerpt != "摘要" {
		t.Fatalf("expected excerpt 摘要, got %q", result.Excerpt)
	}
	if result.SEOKeywords != "go,ai" {
		t.Fatalf("expected normalized keywords go,ai, got %q", result.SEOKeywords)
	}
	if result.SEODescription != "描述" {
		t.Fatalf("expected seo description 描述, got %q", result.SEODescription)
	}
}

func TestParseResponseClampsLongFields(t *testing.T) {
	result, err := parseResponse(`{"excerpt":"` + strings.Repeat("摘", 600) + `","seo_keywords":"` + strings.Repeat("关", 300) + `","seo_description":"` + strings.Repeat("描", 600) + `"}`)
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len([]rune(result.Excerpt)) != maxExcerptLength {
		t.Fatalf("expected excerpt length %d, got %d", maxExcerptLength, len([]rune(result.Excerpt)))
	}
	if len([]rune(result.SEODescription)) != maxSEODescriptionLength {
		t.Fatalf("expected seo description length %d, got %d", maxSEODescriptionLength, len([]rune(result.SEODescription)))
	}
	if len([]rune(result.SEOKeywords)) != maxSEOKeywordsLength {
		t.Fatalf("expected seo keywords length %d, got %d", maxSEOKeywordsLength, len([]rune(result.SEOKeywords)))
	}
}

func TestParseResponseRejectsInvalidJSON(t *testing.T) {
	if _, err := parseResponse("not-json"); err == nil {
		t.Fatalf("expected parse error for invalid json")
	}
}
