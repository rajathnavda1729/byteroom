package sanitizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeMarkdown_ScriptTag_IsStripped(t *testing.T) {
	input := `Hello <script>alert("xss")</script> world`
	result := SanitizeMarkdown(input)

	assert.NotContains(t, result, "<script>")
	assert.NotContains(t, result, "alert")
	assert.Contains(t, result, "Hello")
}

func TestSanitizeMarkdown_OnClickAttribute_IsStripped(t *testing.T) {
	input := `<p onclick="evil()">Click me</p>`
	result := SanitizeMarkdown(input)

	assert.NotContains(t, result, "onclick")
	assert.Contains(t, result, "Click me")
}

func TestSanitizeMarkdown_SafeBoldTag_IsPreserved(t *testing.T) {
	input := `<strong>bold text</strong>`
	result := SanitizeMarkdown(input)

	assert.Contains(t, result, "bold text")
}

func TestSanitizeMarkdown_IframeTag_IsStripped(t *testing.T) {
	input := `<iframe src="https://evil.com"></iframe>`
	result := SanitizeMarkdown(input)

	assert.NotContains(t, result, "iframe")
}

func TestSanitizeMarkdown_EmptyString_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", SanitizeMarkdown(""))
}

func TestSanitizeMarkdown_PlainText_PassesThrough(t *testing.T) {
	input := "Hello, world! No HTML here."
	result := SanitizeMarkdown(input)

	assert.Equal(t, input, result)
}
