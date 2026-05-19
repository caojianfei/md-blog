package ai

import (
	"context"
	"time"

	"google.golang.org/genai"
)

type GeminiGenerator struct {
	config Config
}

func (g GeminiGenerator) GenerateArticleMetadata(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.TimeoutSeconds)*time.Second)
	defer cancel()

	clientConfig := &genai.ClientConfig{
		APIKey:  g.config.APIKey,
		Backend: genai.BackendGeminiAPI,
	}
	if g.config.BaseURL != "" {
		clientConfig.HTTPOptions.BaseURL = g.config.BaseURL
	}

	client, err := genai.NewClient(reqCtx, clientConfig)
	if err != nil {
		return nil, err
	}

	resp, err := client.Models.GenerateContent(reqCtx, g.config.Model, genai.Text(buildPrompt(input)), &genai.GenerateContentConfig{
		Temperature:      genai.Ptr[float32](0.2),
		MaxOutputTokens:  1024,
		ResponseMIMEType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	raw := resp.Text()
	if raw == "" {
		return &GenerateResult{}, nil
	}
	return parseResponse(raw)
}
