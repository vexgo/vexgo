package seo

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"go-cms/internal/db"
)

func TestGenerateSitemap(t *testing.T) {
	posts := []db.Post{
		{
			Slug:      "hello-world",
			Status:    "published",
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			Slug:      "draft-post",
			Status:    "draft",
			UpdatedAt: time.Now(),
		},
	}
	pages := []db.Page{
		{
			Slug:      "about",
			Status:    "published",
			UpdatedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	baseURL := "https://example.com"
	result := GenerateSitemap(posts, pages, baseURL)

	// Check XML header
	if !strings.Contains(result, xml.Header) {
		t.Error("missing XML header")
	}

	// Check xmlns
	if !strings.Contains(result, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Error("missing xmlns")
	}

	// Check homepage
	if !strings.Contains(result, "<loc>https://example.com/</loc>") {
		t.Error("missing homepage URL")
	}

	// Check published post is included
	if !strings.Contains(result, "<loc>https://example.com/posts/hello-world</loc>") {
		t.Error("missing published post URL")
	}

	// Check draft post is excluded
	if strings.Contains(result, "<loc>https://example.com/posts/draft-post</loc>") {
		t.Error("draft post should not be in sitemap")
	}

	// Check published page is included
	if !strings.Contains(result, "<loc>https://example.com/pages/about</loc>") {
		t.Error("missing published page URL")
	}

	// Check lastmod format
	if !strings.Contains(result, "2024-01-15T") {
		t.Error("missing lastmod for post")
	}
}

func TestGenerateRobots(t *testing.T) {
	baseURL := "https://example.com"
	result := GenerateRobots(baseURL)

	if !strings.Contains(result, "User-agent: *") {
		t.Error("missing User-agent")
	}
	if !strings.Contains(result, "Allow: /") {
		t.Error("missing Allow")
	}
	if !strings.Contains(result, "Sitemap: https://example.com/sitemap.xml") {
		t.Error("missing Sitemap directive")
	}
}
