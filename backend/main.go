package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"vexgo/backend/cmd"
	"vexgo/backend/config"
	"vexgo/backend/handler"
	"vexgo/backend/middleware"
	"vexgo/backend/public"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Parse command line arguments
	cfg := cmd.ParseFlags()

	// 2. Initialize configuration (load JWT secret, etc., support config files and environment variables)
	config.Init(cfg.JWTSecret)

	// 2.1 Load SSO configuration from config file (overrides environment variables)
	config.LoadFromConfig(cfg)

	// Set data directory (for file uploads, only used if S3 is not enabled)
	handler.DataDir = cfg.DataDir

	// 3. Initialize S3 storage if enabled
	if cfg.S3Enabled {
		s3Cfg := &config.S3Config{
			Enabled:                  cfg.S3Enabled,
			Endpoint:                 cfg.S3Endpoint,
			Region:                   cfg.S3Region,
			Bucket:                   cfg.S3Bucket,
			AccessKey:                cfg.S3AccessKey,
			SecretKey:                cfg.S3SecretKey,
			ForcePath:                cfg.S3ForcePath,
			CustomDomain:             cfg.S3CustomDomain,
			DisableBucketInCustomURL: cfg.S3DisableBucketInCustomURL,
		}
		fmt.Printf("S3 Config Loaded: Enabled=%v, Endpoint=%s, Region=%s, Bucket=%s, CustomDomain=%s, DisableBucketInCustomURL=%v\n", s3Cfg.Enabled, s3Cfg.Endpoint, s3Cfg.Region, s3Cfg.Bucket, s3Cfg.CustomDomain, s3Cfg.DisableBucketInCustomURL)
		if err := handler.InitS3(s3Cfg); err != nil {
			panic(err)
		}
		fmt.Println("S3 storage initialized")
	} else {
		fmt.Println("Using local file storage")
	}

	// 4. Initialize database connection (ensure database driver and connection string are configured correctly)
	handler.InitDB(cfg, cfg.DataDir)
	// Set database connection to authentication middleware
	middleware.SetDB(handler.DB())

	// 4. Create Gin engine instance (includes Logger and Recovery middleware by default)
	r := gin.Default()

	// Configure trusted proxies based on environment/configuration
	// If BEHIND_REVERSE_PROXY=true, use TRUSTED_PROXIES list or common defaults
	// If BEHIND_REVERSE_PROXY=false, disable proxy trust (no warning)
	if cfg.BehindReverseProxy {
		if len(cfg.TrustedProxies) > 0 {
			// Use explicitly configured trusted proxies
			r.SetTrustedProxies(cfg.TrustedProxies)
			fmt.Printf("Trusted proxies configured: %v\n", cfg.TrustedProxies)
		} else {
			// Use common defaults: trust all private IP ranges and localhost
			// This is a reasonable default for self-hosted behind reverse proxy
			defaultProxies := []string{
				"127.0.0.1",
				"::1",
				"192.168.0.0/16",
				"10.0.0.0/8",
				"172.16.0.0/12",
			}
			r.SetTrustedProxies(defaultProxies)
			fmt.Printf("Trusted proxies set to common private networks (behind reverse proxy)\n")
		}
	} else {
		// Not behind a reverse proxy, disable trust
		r.SetTrustedProxies(nil)
	}

	// ===================== Core API routing group (all endpoints under /api) =====================
	api := r.Group("/api")
	api.Use(middleware.OptionalJWTAuth())
	{
		// -------------------- Public API (no JWT authentication required) --------------------
		api.GET("/posts", handler.GetPosts)
		api.GET("/posts/:id", handler.GetPost)

		api.GET("/verify-email", handler.VerifyEmail)

		api.GET("/captcha", handler.GenerateCaptcha)
		api.POST("/captcha/verify", handler.VerifyCaptcha)

		api.GET("/categories", handler.GetCategories)
		api.GET("/tags", handler.GetTags)

		api.GET("/stats", handler.GetStats)
		api.GET("/stats/popular-posts", handler.GetPopularPosts)
		api.GET("/stats/latest-posts", handler.GetLatestPosts)

		api.GET("/comments/post/:id", handler.GetComments)
		api.GET("/likes/:postId", handler.GetLikeStatus)
		api.GET("/posts/user/:id", handler.GetUserPosts)

		// -------------------- Authentication related API --------------------
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.GET("/me", middleware.JWTAuth(), handler.GetCurrentUser)
			auth.GET("/user", middleware.JWTAuth(), handler.GetCurrentUser)
			auth.PUT("/profile", middleware.JWTAuth(), handler.UpdateProfile)
			auth.PUT("/password", middleware.JWTAuth(), handler.ChangePassword)
			auth.PUT("/email", middleware.JWTAuth(), handler.UpdateEmail)
			auth.PUT("/settings", middleware.JWTAuth(), handler.UpdateSettings)
			auth.POST("/request-password-reset", handler.RequestPasswordReset)
			auth.POST("/reset-password", handler.ResetPassword)
			auth.GET("/verification-status", middleware.JWTAuth(), handler.GetVerificationStatus)
			auth.POST("/resend-verification", middleware.JWTAuth(), handler.ResendVerificationEmail)
		}

		// -------------------- SSO --------------------
		sso := api.Group("/sso")
		{
			// Public: returns enabled providers, used by frontend to show/hide buttons
			// GET /api/sso/providers
			sso.GET("/providers", handler.SSOProviders)

			// Step 1: open in popup → redirects to provider
			// GET /api/sso/:provider/login?method=sso_get_token|get_sso_id
			sso.GET("/:provider/login", handler.SSOLoginRedirect)

			// Step 2: provider redirects back, popup sends postMessage → closes
			// GET /api/sso/:provider/callback?method=...&code=...&state=...
			sso.GET("/:provider/callback", func(c *gin.Context) {
				handler.SSOCallback(c, handler.DB())
			})
		}

		// -------------------- Business API requiring JWT authentication --------------------
		api.POST("/posts", middleware.JWTAuth(), handler.CreatePost)
		api.GET("/posts/user/my-posts", middleware.JWTAuth(), handler.GetMyPosts)
		api.GET("/posts/drafts", middleware.JWTAuth(), handler.GetDraftPosts)
		api.PUT("/posts/:id", middleware.JWTAuth(), handler.UpdatePost)
		api.DELETE("/posts/:id", middleware.JWTAuth(), handler.DeletePost)

		api.POST("/categories", middleware.JWTAuth(), handler.CreateCategory)
		api.POST("/tags", middleware.JWTAuth(), handler.CreateTag)

		api.POST("/comments", middleware.JWTAuth(), handler.CreateComment)
		api.DELETE("/comments/:id", middleware.JWTAuth(), handler.DeleteComment)

		api.GET("/moderation/comments/pending", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetPendingComments)
		api.GET("/moderation/comments/approved", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetApprovedComments)
		api.GET("/moderation/comments/rejected", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetRejectedComments)
		api.PUT("/moderation/comments/approve/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.ApproveComment)
		api.PUT("/moderation/comments/reject/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.RejectComment)
		api.GET("/moderation/comments/config", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetCommentModerationConfig)
		api.PUT("/moderation/comments/config", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.UpdateCommentModerationConfig)

		api.POST("/likes/:postId", middleware.JWTAuth(), handler.ToggleLike)

		api.POST("/upload/file", middleware.JWTAuth(), handler.UploadFile)
		api.POST("/upload/files", middleware.JWTAuth(), handler.UploadFiles)
		api.GET("/upload/my-files", middleware.JWTAuth(), handler.GetMyFiles)
		api.DELETE("/upload/:id", middleware.JWTAuth(), handler.DeleteFile)

		api.GET("/moderation/pending", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetPendingPosts)
		api.GET("/moderation/approved", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetApprovedPosts)
		api.GET("/moderation/rejected", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetRejectedPosts)
		api.PUT("/moderation/approve/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.ApprovePost)
		api.PUT("/moderation/reject/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.RejectPost)
		api.PUT("/moderation/resubmit/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.ResubmitPost)

		api.GET("/users", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetUserList)
		api.PUT("/users/:id/role", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.UpdateUserRole)
		api.DELETE("/users/:id", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.DeleteUser)

		api.GET("/config/smtp", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetSMTPConfig)
		api.PUT("/config/smtp", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.UpdateSMTPConfig)
		api.POST("/config/smtp/test", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.TestSMTP)

		api.GET("/config/ai", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetAIConfig)
		api.PUT("/config/ai", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.UpdateAIConfig)
		api.POST("/config/ai/test", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.TestAI)
		api.GET("/config/ai/models", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.GetAIModels)

		api.GET("/config/general", handler.GetGeneralSettings)
		api.PUT("/config/general", middleware.JWTAuth(), middleware.PermissionMiddleware("admin", "super_admin"), handler.UpdateGeneralSettings)
	}

	// ===================== Static file hosting =====================
	// Only serve local uploads if S3 is not enabled
	if !cfg.S3Enabled {
		mediaDir := filepath.Join(cfg.DataDir, "media")
		r.Static("/uploads", mediaDir)
	}

	r.GET("/assets/*filepath", func(c *gin.Context) {
		file := strings.TrimPrefix(c.Param("filepath"), "/")
		content, err := public.ReadAsset("assets/" + file)
		if err != nil {
			c.Status(404)
			return
		}
		ext := filepath.Ext(file)
		switch ext {
		case ".js":
			c.Data(200, "application/javascript", content)
		case ".css":
			c.Data(200, "text/css", content)
		case ".html":
			c.Data(200, "text/html", content)
		case ".json":
			c.Data(200, "application/json", content)
		case ".png":
			c.Data(200, "image/png", content)
		case ".jpg", ".jpeg":
			c.Data(200, "image/jpeg", content)
		case ".gif":
			c.Data(200, "image/gif", content)
		case ".svg":
			c.Data(200, "image/svg+xml", content)
		case ".ico":
			c.Data(200, "image/x-icon", content)
		case ".woff":
			c.Data(200, "font/woff", content)
		case ".woff2":
			c.Data(200, "font/woff2", content)
		default:
			c.Data(200, "application/octet-stream", content)
		}
	})

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", public.GetIndexHTML())
	})

	// ===================== Frontend SPA route compatibility =====================
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Data(200, "text/html; charset=utf-8", public.GetIndexHTML())
			return
		}
		c.JSON(404, gin.H{"error": "Not Found"})
	})

	r.Run(cfg.GetListenAddr())
}
