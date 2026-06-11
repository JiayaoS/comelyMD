package render

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	html_renderer "github.com/yuin/goldmark/renderer/html"
)

var (
	displayMathRe = regexp.MustCompile(`(?s)\$\$[^\$]+\$\$`)
	inlineMathRe  = regexp.MustCompile(`\$[^\$\n]+\$`)
)

// MarkdownToHTML renders Markdown and sanitizes the resulting HTML for display.
func MarkdownToHTML(source []byte) (string, error) {
	// Preserve math expressions so Markdown emphasis does not alter their syntax.
	mathBlocks := make(map[string][]byte)
	counter := 0

	source = displayMathRe.ReplaceAllFunc(source, func(match []byte) []byte {
		placeholder := fmt.Sprintf("COMELYMATHDISPLAY%dENDMATH", counter)
		mathBlocks[placeholder] = match
		counter++
		return []byte(placeholder)
	})

	source = inlineMathRe.ReplaceAllFunc(source, func(match []byte) []byte {
		placeholder := fmt.Sprintf("COMELYMATHINLINE%dENDMATH", counter)
		mathBlocks[placeholder] = match
		counter++
		return []byte(placeholder)
	})

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html_renderer.WithHardWraps(),
			html_renderer.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", err
	}

	p := bluemonday.UGCPolicy()

	p.AllowAttrs("class").OnElements("code", "pre", "span")
	p.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6")

	safeHTML := string(p.SanitizeBytes(buf.Bytes()))

	// Restore escaped math expressions after sanitization.
	for placeholder, original := range mathBlocks {
		escapedMath := html.EscapeString(string(original))
		safeHTML = strings.Replace(safeHTML, placeholder, escapedMath, 1)
	}

	return safeHTML, nil
}
