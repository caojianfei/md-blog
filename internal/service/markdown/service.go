package markdown

import (
	"bytes"
	"html"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	htmlRenderer "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id"`
}

type RenderResult struct {
	HTML     string    `json:"html"`
	Headings []Heading `json:"headings"`
}

type Service struct {
	engine goldmark.Markdown
}

func New() *Service {
	engine := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.TaskList,
			extension.Strikethrough,
			extension.Linkify,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(true),
				),
				highlighting.WithWrapperRenderer(func(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
					language := "text"
					if raw, ok := c.Language(); ok && len(raw) > 0 {
						language = strings.ToLower(string(raw))
					}
					escapedLanguage := html.EscapeString(language)
					if entering {
						_, _ = w.WriteString(`<div class="code-block" data-language="` + escapedLanguage + `">`)
						_, _ = w.WriteString(`<div class="code-block__header"><span class="code-block__label">` + escapedLanguage + `</span><button type="button" class="code-copy-button" data-copy-code>复制</button></div>`)
						_, _ = w.WriteString(`<div class="code-block__body">`)
						return
					}
					_, _ = w.WriteString(`</div></div>`)
				}),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(htmlRenderer.WithHardWraps()),
	)
	return &Service{engine: engine}
}

func (s *Service) Render(source string) (*RenderResult, error) {
	var buf bytes.Buffer
	if err := s.engine.Convert([]byte(source), &buf); err != nil {
		return nil, err
	}

	headings := make([]Heading, 0)
	reader := text.NewReader([]byte(source))
	doc := s.engine.Parser().Parse(reader)
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node.Kind() != ast.KindHeading {
			return ast.WalkContinue, nil
		}
		heading := node.(*ast.Heading)
		anchorID := ""
		if rawID, ok := heading.AttributeString("id"); ok {
			if value, ok := rawID.([]byte); ok {
				anchorID = string(value)
			}
		}
		headings = append(headings, Heading{
			Level: heading.Level,
			Text:  strings.TrimSpace(extractNodeText([]byte(source), heading)),
			ID:    anchorID,
		})
		return ast.WalkContinue, nil
	})

	return &RenderResult{HTML: buf.String(), Headings: headings}, nil
}

func extractNodeText(source []byte, node ast.Node) string {
	var builder strings.Builder
	var visit func(ast.Node)
	visit = func(current ast.Node) {
		for child := current.FirstChild(); child != nil; child = child.NextSibling() {
			if textNode, ok := child.(*ast.Text); ok {
				builder.Write(textNode.Segment.Value(source))
				continue
			}
			visit(child)
		}
	}
	visit(node)
	return builder.String()
}
