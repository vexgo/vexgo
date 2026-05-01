package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go-cms/internal/cms"
	"go-cms/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	dbConn, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	dbConn.AutoMigrate(&db.Post{}, &db.User{})
	cs := cms.NewContentService(dbConn)
	_ = cms.NewMediaService(dbConn)
	_ = cms.NewSettingsService(dbConn)
	postHandler := NewPostHandler(cs)

	public := r.Group("/api")
	{
		public.GET("/posts", postHandler.List)
		public.GET("/posts/:slug", postHandler.GetBySlug)
	}
	return r
}

func TestListPostsAPI(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/posts?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
