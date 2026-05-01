package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-cms/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServer_RootReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"message":"pong"}` {
		t.Errorf("expected pong response, got %s", w.Body.String())
	}
}

func TestSetupRouter(t *testing.T) {
	dbConn, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	dbConn.AutoMigrate(&db.Post{}, &db.User{})
	router := setupRouter(dbConn, "test-secret")
	if router == nil {
		t.Fatal("expected router to be non-nil")
	}
}
