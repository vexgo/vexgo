package cms

import (
	"testing"

	"go-cms/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	dbConn, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	dbConn.AutoMigrate(&db.Post{}, &db.Page{}, &db.Setting{})
	return dbConn
}

func TestCreatePost(t *testing.T) {
	database := setupTestDB()
	service := NewContentService(database)

	post, err := service.CreatePost(CreatePostInput{
		Title:   "Hello World",
		Content: "First post",
		Status:  "published",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Title != "Hello World" {
		t.Errorf("expected title Hello World, got %s", post.Title)
	}
	if post.Slug == "" {
		t.Error("expected slug to be generated")
	}
}

func TestGetPostBySlug(t *testing.T) {
	database := setupTestDB()
	service := NewContentService(database)
	service.CreatePost(CreatePostInput{
		Title:   "Test Post",
		Content: "Content here",
		Status:  "published",
	})

	post, err := service.GetPostBySlug("test-post")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Title != "Test Post" {
		t.Errorf("expected Test Post, got %s", post.Title)
	}
}

func TestListPosts(t *testing.T) {
	database := setupTestDB()
	service := NewContentService(database)
	service.CreatePost(CreatePostInput{Title: "Post 1", Content: "C1", Status: "published"})
	service.CreatePost(CreatePostInput{Title: "Post 2", Content: "C2", Status: "published"})

	posts, total, err := service.ListPosts(1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
}

func TestUpdatePost(t *testing.T) {
	database := setupTestDB()
	service := NewContentService(database)
	post, _ := service.CreatePost(CreatePostInput{Title: "Old", Content: "Old", Status: "published"})

	updated, err := service.UpdatePost(post.ID, CreatePostInput{
		Title:  "New Title",
		Content: "New Content",
		Status:  "published",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected New Title, got %s", updated.Title)
	}
}

func TestDeletePost(t *testing.T) {
	database := setupTestDB()
	service := NewContentService(database)
	post, _ := service.CreatePost(CreatePostInput{Title: "ToDelete", Content: "X", Status: "published"})

	err := service.DeletePost(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetPostBySlug("todelete")
	if err == nil {
		t.Error("expected error after delete")
	}
}
