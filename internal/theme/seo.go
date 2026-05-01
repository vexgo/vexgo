package theme

import (
	"fmt"
	"html/template"
	"strings"
)

func MetaTags(title, description, ogImage, canonicalURL string) template.HTML {
	var tags []string
	if title != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="title" content="%s">`, escapeHTML(title)))
		tags = append(tags, fmt.Sprintf(`<meta property="og:title" content="%s">`, escapeHTML(title)))
	}
	if description != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="description" content="%s">`, escapeHTML(description)))
		tags = append(tags, fmt.Sprintf(`<meta property="og:description" content="%s">`, escapeHTML(description)))
	}
	if ogImage != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:image" content="%s">`, ogImage))
	}
	if canonicalURL != "" {
		tags = append(tags, fmt.Sprintf(`<link rel="canonical" href="%s">`, canonicalURL))
	}
	return template.HTML(strings.Join(tags, "\n    "))
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
