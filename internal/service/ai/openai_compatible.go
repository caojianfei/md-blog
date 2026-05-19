package ai

import (
	"context"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAICompatibleGenerator struct {
	config Config
}

func (g OpenAICompatibleGenerator) GenerateArticleMetadata(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.TimeoutSeconds)*time.Second)
	defer cancel()

	clientConfig := openai.DefaultConfig(g.config.APIKey)
	clientConfig.BaseURL = g.config.BaseURL
	client := openai.NewClientWithConfig(clientConfig)

	resp, err := client.CreateChatCompletion(reqCtx, openai.ChatCompletionRequest{
		Model: g.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: buildPrompt(input),
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return &GenerateResult{}, nil
	}
	return parseResponse(resp.Choices[0].Message.Content)
}
