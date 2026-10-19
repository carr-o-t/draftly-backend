package sanitize

import "github.com/microcosm-cc/bluemonday"

// the stored html is later dropped into a thumbnail with innerHTML, so we
// only keep the tags our editor can actually produce. everything else
// (scripts, event handlers, styles, iframes, forms ...) is stripped.
var policy = buildPolicy()

func buildPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// only http, https and mailto urls, so javascript: or data: links die
	p.AllowStandardURLs()

	p.AllowElements(
		"p", "br", "hr",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"strong", "b", "em", "i", "u", "s", "strike", "del", "mark",
		"code", "pre", "blockquote",
		"ul", "ol", "li",
		"a", "span",
	)

	// links, tiptap writes href/target/rel
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("target", "rel").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
	// class is used for code block languages and inline marks
	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements("code", "pre", "span", "a")

	// force safe links, nofollow and open externally in a new tab
	p.RequireNoFollowOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)

	return p
}

// HTML returns the content with anything not on our allow list removed.
func HTML(content string) string {
	if content == "" {
		return content
	}
	return policy.Sanitize(content)
}
