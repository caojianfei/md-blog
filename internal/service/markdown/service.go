package markdown

import (
    "bytes"
    "strings"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark-highlighting/v2"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/parser"
    htmlRenderer "github.com/yuin/goldmark/renderer/html"
    "github.com/yuin/goldmark/text"
)

type Heading struct {
    Level int    `json:"level"`
    Text  string `json:"text"`
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
            highlighting.NewHighlighting(),
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
        headings = append(headings, Heading{Level: heading.Level, Text: strings.TrimSpace(string(heading.Text([]byte(source))))})
        return ast.WalkContinue, nil
    })

    return &RenderResult{HTML: buf.String(), Headings: headings}, nil
}
