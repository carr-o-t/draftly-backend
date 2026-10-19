package sanitize

import (
	"strings"
	"testing"
)

func TestHTMLStripsDangerousContent(t *testing.T) {
	cases := map[string]string{
		"script tag":     `<p>hi</p><script>alert(1)</script>`,
		"inline handler": `<p onclick="steal()">hi</p>`,
		"js url":         `<a href="javascript:alert(1)">x</a>`,
		"iframe":         `<iframe src="https://evil.test"></iframe>`,
		"img onerror":    `<img src=x onerror=alert(1)>`,
		"style attr":     `<p style="position:fixed">hi</p>`,
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			out := HTML(input)
			lower := strings.ToLower(out)
			for _, bad := range []string{"<script", "onclick", "javascript:", "<iframe", "onerror", "style="} {
				if strings.Contains(lower, bad) {
					t.Fatalf("expected %q to be removed, got %q", bad, out)
				}
			}
		})
	}
}

func TestHTMLKeepsEditorMarkup(t *testing.T) {
	input := `<h1>Title</h1><p>some <strong>bold</strong> and <em>italic</em> <u>text</u></p><ul><li>one</li></ul><pre><code class="language-go">fmt.Println()</code></pre>`
	out := HTML(input)

	for _, want := range []string{"<h1>", "<strong>", "<em>", "<u>", "<ul>", "<li>", "<pre>", "<code"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q to be kept, got %q", want, out)
		}
	}
}

func TestHTMLEmptyString(t *testing.T) {
	if got := HTML(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
