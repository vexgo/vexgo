package seo

import (
	"encoding/xml"
	"fmt"
	"time"

	"go-cms/internal/db"
)

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// GenerateSitemap creates a sitemap XML string from posts and pages.
// baseURL should not have a trailing slash.
func GenerateSitemap(posts []db.Post, pages []db.Page, baseURL string) string {
	sitemap := Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []URL{},
	}

	// Add homepage
	sitemap.URLs = append(sitemap.URLs, URL{
		Loc:        baseURL + "/",
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	// Add posts
	for _, post := range posts {
		if post.Status != "published" {
			continue
		}
		lastMod := post.UpdatedAt.Format(time.RFC3339)
		sitemap.URLs = append(sitemap.URLs, URL{
			Loc:        fmt.Sprintf("%s/posts/%s", baseURL, post.Slug),
			LastMod:    lastMod,
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	// Add pages
	for _, page := range pages {
		if page.Status != "published" {
			continue
		}
		lastMod := page.UpdatedAt.Format(time.RFC3339)
		sitemap.URLs = append(sitemap.URLs, URL{
			Loc:        fmt.Sprintf("%s/pages/%s", baseURL, page.Slug),
			LastMod:    lastMod,
			ChangeFreq: "monthly",
			Priority:   "0.7",
		})
	}

	output, _ := xml.MarshalIndent(sitemap, "", "  ")
	return xml.Header + string(output)
}
