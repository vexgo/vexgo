package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go-cms/internal/admin"
	"go-cms/internal/api"
	"go-cms/internal/auth"
	"go-cms/internal/cms"
	"go-cms/internal/db"
	"go-cms/internal/middleware"
	"go-cms/internal/seo"
	"go-cms/internal/theme"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupRouter(database *gorm.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Initialize theme loader with embedded default theme
	themeLoader := theme.NewLoader(theme.DefaultThemeFS, "./internal/theme/themes")
	defaultTheme, err := themeLoader.LoadEmbedded()
	if err != nil {
		log.Printf("warning: could not load embedded default theme: %v", err)
		// Fallback to filesystem
		defaultTheme, err = themeLoader.LoadFromFS("default")
		if err != nil {
			log.Printf("warning: could not load default theme from filesystem: %v", err)
		}
	}
	if defaultTheme != nil {
		themeLoader.SetActive(defaultTheme)
		// Serve theme static files from embedded FS
		if defaultTheme.StaticFS != nil {
			r.StaticFS("/themes/default/static", http.FS(defaultTheme.StaticFS))
		}
	}

	// Visitor frontend routes (SEO-friendly SSR)
	cs := cms.NewContentService(database)

	// Home page
	r.GET("/", func(c *gin.Context) {
		posts, _, err := cs.ListPosts(1, 10)
		if err != nil {
			c.Status(500)
			return
		}
		if err := themeLoader.Render(c.Writer, "home", theme.RenderData{
			SiteName:    "Go CMS",
			MetaTitle:   "Home",
			Content:     posts,
			ThemeName:   "default",
		}); err != nil {
			c.Status(404)
			return
		}
	})

	// Single post
	r.GET("/posts/:slug", func(c *gin.Context) {
		post, err := cs.GetPostBySlug(c.Param("slug"))
		if err != nil {
			c.Status(404)
			return
		}
		if err := themeLoader.Render(c.Writer, "post", theme.RenderData{
			SiteName:        "Go CMS",
			MetaTitle:       post.Title,
			MetaDescription: post.Excerpt,
			CanonicalURL:    c.Request.URL.String(),
			Content:         post,
			ThemeName:       "default",
		}); err != nil {
			c.Status(404)
			return
		}
	})

	// Single page
	r.GET("/pages/:slug", func(c *gin.Context) {
		page, err := cs.GetPageBySlug(c.Param("slug"))
		if err != nil {
			c.Status(404)
			return
		}
		if err := themeLoader.Render(c.Writer, "page", theme.RenderData{
			SiteName:        "Go CMS",
			MetaTitle:       page.Title,
			MetaDescription: page.MetaDesc,
			CanonicalURL:    c.Request.URL.String(),
			Content:         page,
			ThemeName:       "default",
		}); err != nil {
			c.Status(404)
			return
		}
	})

	// Archive page
	r.GET("/archive", func(c *gin.Context) {
		posts, _, _ := cs.ListPosts(1, 10000)
		pages, _, _ := cs.ListPages(1, 10000)
		if err := themeLoader.Render(c.Writer, "archive", theme.RenderData{
			SiteName:  "Go CMS",
			MetaTitle: "Archive",
			Content:   gin.H{"Posts": posts, "Pages": pages},
			ThemeName: "default",
		}); err != nil {
			c.Status(404)
			return
		}
	})

	// Auth
	authHandler := api.NewAuthHandler(database)
	r.POST("/api/login", authHandler.Login)

	// Public API
	postHandler := api.NewPostHandler(cs)
	pageHandler := api.NewPageHandler(cs)
	public := r.Group("/api")
	{
		public.GET("/posts", postHandler.List)
		public.GET("/posts/:slug", postHandler.GetBySlug)
		public.GET("/pages/:slug", pageHandler.GetBySlug)
	}

	// Protected API
	auth.SetSecret(jwtSecret)
	protected := r.Group("/api")
	protected.Use(auth.Middleware())
	{
		protected.POST("/posts", postHandler.Create)
		protected.PUT("/posts/:id", postHandler.Update)
		protected.DELETE("/posts/:id", postHandler.Delete)
		protected.POST("/pages", pageHandler.Create)
		protected.PUT("/pages/:id", pageHandler.Update)
		protected.DELETE("/pages/:id", pageHandler.Delete)

		mediaHandler := api.NewMediaHandler(cms.NewMediaService(database))
		protected.GET("/media", mediaHandler.List)
		protected.POST("/media/upload", mediaHandler.Upload)
		protected.DELETE("/media/:id", mediaHandler.Delete)

		settingsHandler := api.NewSettingsHandler(cms.NewSettingsService(database))
		protected.GET("/settings", settingsHandler.GetAll)
		protected.PUT("/settings", settingsHandler.Update)
	}

	// SEO routes
	r.GET("/sitemap.xml", func(c *gin.Context) {
		posts, _, _ := cs.ListPosts(1, 10000)
		pages, _, _ := cs.ListPages(1, 10000)
		baseURL := "http://" + c.Request.Host
		xml := seo.GenerateSitemap(posts, pages, baseURL)
		c.Header("Content-Type", "application/xml")
		c.String(200, xml)
	})
	r.GET("/robots.txt", func(c *gin.Context) {
		baseURL := "http://" + c.Request.Host
		c.String(200, seo.GenerateRobots(baseURL))
	})

	// Strip "dist/" prefix from embedded FS for clean /admin/ serving
	adminFS, err := fs.Sub(admin.AdminDistFS, "dist")
	if err != nil {
		log.Fatalf("failed to sub FS: %v", err)
	}
	// Unified handler for /admin/* - serves static files or SPA fallback
	r.GET("/admin/*any", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("any"), "/")
		// Try to serve static file
		data, err := fs.ReadFile(adminFS, path)
		if err == nil {
			// Determine content type
			contentType := "application/octet-stream"
			switch {
			case strings.HasSuffix(path, ".html"):
				contentType = "text/html; charset=utf-8"
			case strings.HasSuffix(path, ".js"):
				contentType = "text/javascript; charset=utf-8"
			case strings.HasSuffix(path, ".css"):
				contentType = "text/css; charset=utf-8"
			case strings.HasSuffix(path, ".json"):
				contentType = "application/json"
			case strings.HasSuffix(path, ".png"):
				contentType = "image/png"
			case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
				contentType = "image/jpeg"
			case strings.HasSuffix(path, ".svg"):
				contentType = "image/svg+xml"
			case strings.HasSuffix(path, ".ico"):
				contentType = "image/x-icon"
			case strings.HasSuffix(path, ".woff"):
				contentType = "font/woff"
			case strings.HasSuffix(path, ".woff2"):
				contentType = "font/woff2"
			}
			c.Data(200, contentType, data)
			return
		}
		// SPA fallback - serve index.html for client-side routes
		data, err = fs.ReadFile(adminFS, "index.html")
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	return r
}

func main() {
	dsn := os.Getenv("CMS_DSN")
	if dsn == "" {
		dsn = "./data/cms.db"
	}
	database, err := db.NewConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
	// Create default admin user if none exists
	var count int64
	database.Model(&db.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		database.Create(&db.User{Username: "admin", Password: string(hash), Email: "admin@example.com", Role: "admin"})
		log.Println("created default admin user: admin / admin123")
	}

	port := os.Getenv("CMS_PORT")
	if port == "" {
		port = "8080"
	}
	jwtSecret := os.Getenv("CMS_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	router := setupRouter(database, jwtSecret)
	log.Printf("starting server on :%s", port)
	router.Run(":" + port)
}
