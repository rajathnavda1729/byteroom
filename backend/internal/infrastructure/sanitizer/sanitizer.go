package sanitizer

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var ugcPolicy = bluemonday.UGCPolicy()

// SanitizeMarkdown removes unsafe HTML from user-provided markdown content,
// preventing XSS while allowing safe formatting tags.
//
// Fenced code blocks (``` ... ```) are left untouched: bluemonday would
// escape characters like > in Mermaid arrows (-->) as HTML entities and
// break diagrams and other code.
func SanitizeMarkdown(content string) string {
	var b strings.Builder
	b.Grow(len(content) + len(content)/8)
	rest := content
	for {
		i := strings.Index(rest, "```")
		if i == -1 {
			b.WriteString(ugcPolicy.Sanitize(rest))
			return b.String()
		}
		b.WriteString(ugcPolicy.Sanitize(rest[:i]))
		rest = rest[i:]
		// rest starts with ```
		nl := strings.IndexByte(rest[3:], '\n')
		if nl == -1 {
			b.WriteString(ugcPolicy.Sanitize(rest))
			return b.String()
		}
		nl += 3 // index of newline ending ```lang line
		closeRel := strings.Index(rest[nl+1:], "\n```")
		if closeRel == -1 {
			b.WriteString(ugcPolicy.Sanitize(rest))
			return b.String()
		}
		closeAbs := nl + 1 + closeRel
		end := closeAbs + 4 // include "\n```"
		b.WriteString(rest[:end])
		rest = rest[end:]
	}
}
