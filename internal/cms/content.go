package cms

import (
	"strings"
	"time"
	"go-cms/internal/db"
	"gorm.io/gorm"
)

type ContentService struct {
	db *gorm.DB
}

func NewContentService(database *gorm.DB) *ContentService {
	return &ContentService{db: database}
}

type CreatePostInput struct {
	Title      string
	Content    string
	Status     string
	MetaTitle  string
	MetaDesc   string
	OGImage    string
	Tags       string
	CategoryID uint
}

func (s *ContentService) CreatePost(input CreatePostInput) (*db.Post, error) {
	slug := generateSlug(input.Title)
	now := time.Now()
	post := &db.Post{
		Title:      input.Title,
		Slug:       slug,
		Content:    input.Content,
		Status:     input.Status,
		MetaTitle:  input.MetaTitle,
		MetaDesc:   input.MetaDesc,
		OGImage:    input.OGImage,
		Tags:       input.Tags,
		CategoryID: input.CategoryID,
		PublishedAt: &now,
	}
	if err := s.db.Create(post).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (s *ContentService) GetPostBySlug(slug string) (*db.Post, error) {
	var post db.Post
	if err := s.db.Where("slug = ? AND status = ?", slug, "published").First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *ContentService) ListPosts(page, pageSize int) ([]db.Post, int64, error) {
	var posts []db.Post
	var total int64
	s.db.Model(&db.Post{}).Where("status = ?", "published").Count(&total)
	offset := (page - 1) * pageSize
	if err := s.db.Where("status = ?", "published").Offset(offset).Limit(pageSize).Order("published_at DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (s *ContentService) UpdatePost(id uint, input CreatePostInput) (*db.Post, error) {
	var post db.Post
	if err := s.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	post.Title = input.Title
	post.Content = input.Content
	post.Status = input.Status
	post.MetaTitle = input.MetaTitle
	post.MetaDesc = input.MetaDesc
	post.OGImage = input.OGImage
	post.Tags = input.Tags
	post.CategoryID = input.CategoryID
	if err := s.db.Save(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *ContentService) DeletePost(id uint) error {
	return s.db.Delete(&db.Post{}, id).Error
}

func (s *ContentService) GetPageBySlug(slug string) (*db.Page, error) {
	var page db.Page
	if err := s.db.Where("slug = ? AND status = ?", slug, "published").First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *ContentService) ListPages(page, pageSize int) ([]db.Page, int64, error) {
	var pages []db.Page
	var total int64
	s.db.Model(&db.Page{}).Where("status = ?", "published").Count(&total)
	offset := (page - 1) * pageSize
	if err := s.db.Where("status = ?", "published").Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&pages).Error; err != nil {
		return nil, 0, err
	}
	return pages, total, nil
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}
