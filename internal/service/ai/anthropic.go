package ai

import (
	"context"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicGenerator struct {
	config Config
}

func (g AnthropicGenerator) GenerateArticleMetadata(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.TimeoutSeconds)*time.Second)
	defer cancel()

	opts := []option.RequestOption{option.WithAPIKey(g.config.APIKey)}
	if g.config.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(g.config.BaseURL))
	}

	client := anthropic.NewClient(opts...)
	resp, err := client.Messages.New(reqCtx, anthropic.MessageNewParams{
		Model:     g.config.Model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(buildPrompt(input))),
		},
	})
	if err != nil {
		return nil, err
	}

	raw := extractAnthropicText(resp)
	if raw == "" {
		return &GenerateResult{}, nil
	}
	return parseResponse(raw)
}

func extractAnthropicText(message *anthropic.Message) string {
	if message == nil || len(message.Content) == 0 {
		return ""
	}

	parts := make([]string, 0, len(message.Content))
	for _, block := range message.Content {
		if text := strings.TrimSpace(block.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
