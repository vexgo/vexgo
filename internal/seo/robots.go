package seo

import "fmt"

// GenerateRobots generates a robots.txt content for the given baseURL.
func GenerateRobots(baseURL string) string {
	return fmt.Sprintf(`User-agent: *
Allow: /

Sitemap: %s/sitemap.xml
`, baseURL)
}
