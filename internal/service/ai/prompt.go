package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	maxExcerptLength        = 500
	maxSEODescriptionLength = 500
	maxSEOKeywordsLength    = 255
)

type promptResponse struct {
	Excerpt        string `json:"excerpt"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`
}

func buildPrompt(input GenerateInput) string {
	return fmt.Sprintf(`你是一个专业的中文技术博客编辑助手。请根据提供的文章标题和 Markdown 正文，生成以下 JSON：
{
  "excerpt": "文章摘要",
  "seo_keywords": "关键词1,关键词2,关键词3",
  "seo_description": "SEO描述"
}

要求：
1. 只能依据标题和正文内容生成，不要捏造正文中没有的信息。
2. 输出必须是 JSON 对象，不要输出 Markdown 代码块，不要附加任何解释。
3. excerpt 用于文章列表展示，1-2 句，简洁自然。
4. seo_description 用于搜索引擎描述，通顺自然，不要堆砌关键词。
5. seo_keywords 返回 3-8 个简洁短语，使用英文逗号分隔。
6. 所有内容都使用简体中文。

标题：
%s

正文：
%s`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content))
}

func parseResponse(raw string) (*GenerateResult, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var payload promptResponse
	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		return nil, err
	}

	return sanitizeResult(&GenerateResult{
		Excerpt:        payload.Excerpt,
		SEOKeywords:    payload.SEOKeywords,
		SEODescription: payload.SEODescription,
	}), nil
}

func SanitizeResult(result *GenerateResult) *GenerateResult {
	return sanitizeResult(result)
}

func sanitizeResult(result *GenerateResult) *GenerateResult {
	if result == nil {
		return &GenerateResult{}
	}

	spaceRe := regexp.MustCompile(`\s+`)
	normalize := func(value string, maxLen int) string {
		value = strings.TrimSpace(value)
		value = spaceRe.ReplaceAllString(value, " ")
		if maxLen <= 0 {
			return value
		}
		runes := []rune(value)
		if len(runes) > maxLen {
			value = string(runes[:maxLen])
		}
		return strings.TrimSpace(value)
	}

	keywords := normalize(result.SEOKeywords, 0)
	keywords = strings.ReplaceAll(keywords, "，", ",")
	keywords = strings.Trim(strings.Join(filterEmpty(strings.Split(keywords, ",")), ","), ",")
	if len([]rune(keywords)) > maxSEOKeywordsLength {
		keywords = string([]rune(keywords)[:maxSEOKeywordsLength])
		keywords = strings.TrimRight(keywords, ",")
	}

	return &GenerateResult{
		Excerpt:        normalize(result.Excerpt, maxExcerptLength),
		SEOKeywords:    keywords,
		SEODescription: normalize(result.SEODescription, maxSEODescriptionLength),
	}
}

func filterEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
