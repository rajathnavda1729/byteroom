package sanitizer

import "github.com/microcosm-cc/bluemonday"

var ugcPolicy = bluemonday.UGCPolicy()

// SanitizeMarkdown removes unsafe HTML from user-provided markdown content,
// preventing XSS while allowing safe formatting tags.
func SanitizeMarkdown(content string) string {
	return ugcPolicy.Sanitize(content)
}
