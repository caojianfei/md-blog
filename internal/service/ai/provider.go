package ai

import (
	"context"
	"fmt"
	"strings"
)

type ProviderType string

const (
	ProviderOpenAICompatible ProviderType = "openai_compatible"
	ProviderAnthropic        ProviderType = "anthropic"
	ProviderGemini           ProviderType = "gemini"
)

type Config struct {
	Enabled        bool
	Provider       ProviderType
	Model          string
	APIKey         string
	BaseURL        string
	TimeoutSeconds int
}

type GenerateInput struct {
	Title   string
	Content string
}

type GenerateResult struct {
	Excerpt        string
	SEOKeywords    string
	SEODescription string
}

type MetadataGenerator interface {
	GenerateArticleMetadata(ctx context.Context, input GenerateInput) (*GenerateResult, error)
}

func NewGenerator(cfg Config) MetadataGenerator {
	if !cfg.Enabled {
		return NoopGenerator{}
	}

	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if cfg.TimeoutSeconds <= 0 || cfg.Model == "" || cfg.APIKey == "" {
		return NoopGenerator{}
	}

	switch cfg.Provider {
	case ProviderOpenAICompatible:
		if cfg.BaseURL == "" {
			return NoopGenerator{}
		}
		return OpenAICompatibleGenerator{config: cfg}
	case ProviderAnthropic:
		return AnthropicGenerator{config: cfg}
	case ProviderGemini:
		return GeminiGenerator{config: cfg}
	default:
		return UnsupportedGenerator{provider: string(cfg.Provider)}
	}
}

type UnsupportedGenerator struct {
	provider string
}

func (g UnsupportedGenerator) GenerateArticleMetadata(context.Context, GenerateInput) (*GenerateResult, error) {
	return nil, fmt.Errorf("unsupported ai provider: %s", g.provider)
}
