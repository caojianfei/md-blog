package ai

import "context"

type NoopGenerator struct{}

func (NoopGenerator) GenerateArticleMetadata(context.Context, GenerateInput) (*GenerateResult, error) {
	return &GenerateResult{}, nil
}
