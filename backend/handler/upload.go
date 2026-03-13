package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"vexgo/backend/config"
	"vexgo/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var DataDir string

// Get file extension
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	return ext
}

// generateFilename generates a unique filename with extension
func generateFilename(originalName string) string {
	ext := getFileExtension(originalName)
	uuid := uuid.New().String()
	if ext != "" {
		return fmt.Sprintf("%s%s", uuid, ext)
	}
	return uuid
}

// Upload file (requires login) and record in database
func UploadFile(c *gin.Context) {
	var userID uint = 0
	if uid, ok := c.Get("userID"); ok {
		if id, ok2 := uid.(uint); ok2 {
			userID = id
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	// Generate unique filename
	filename := generateFilename(file.Filename)

	var fileURL string
	var fileSize int64 = file.Size

	// Upload based on storage configuration
	if UseS3Storage {
		// Open the file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer src.Close()

		// Upload to S3
		url, err := UploadFileToS3(src, filename, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to S3: %v", err)})
			return
		}
		fileURL = url
	} else {
		// Local storage
		uploadDir := filepath.Join(DataDir, "media")
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		fullPath := filepath.Join(uploadDir, filename)

		// Save file
		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fileURL = fmt.Sprintf("/uploads/%s", filename)
	}

	media := model.MediaFile{
		URL:    fileURL,
		Size:   fileSize,
		Type:   "unknown",
		UserID: userID,
	}
	db.Create(&media)

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    media,
	})
}

// Upload multiple files (requires login) and record to database
func UploadFiles(c *gin.Context) {
	var userID uint = 0
	if uid, ok := c.Get("userID"); ok {
		if id, ok2 := uid.(uint); ok2 {
			userID = id
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	files := form.File["files"]
	var uploadedFiles []model.MediaFile

	for _, file := range files {
		filename := generateFilename(file.Filename)
		var fileURL string
		var fileSize int64 = file.Size

		if UseS3Storage {
			// Upload to S3
			src, err := file.Open()
			if err != nil {
				continue
			}
			defer src.Close()

			url, err := UploadFileToS3(src, filename, "")
			if err != nil {
				continue
			}
			fileURL = url
		} else {
			// Local storage
			uploadDir := filepath.Join(DataDir, "media")
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				os.MkdirAll(uploadDir, os.ModePerm)
			}

			fullPath := filepath.Join(uploadDir, filename)

			// Save file
			if err := c.SaveUploadedFile(file, fullPath); err != nil {
				continue
			}
			fileURL = fmt.Sprintf("/uploads/%s", filename)
		}

		media := model.MediaFile{
			URL:    fileURL,
			Size:   fileSize,
			Type:   "unknown",
			UserID: userID,
		}
		db.Create(&media)
		uploadedFiles = append(uploadedFiles, media)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File upload completed",
		"files":   uploadedFiles,
	})
}

// Create tag
func CreateTag(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag := model.Tag{
		Name: req.Name,
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

// Get current user's uploaded files list
func GetMyFiles(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint)
	var files []model.MediaFile
	db.Where("user_id = ?", userID).Find(&files)
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// Delete file (must be uploader or admin)
func DeleteFile(c *gin.Context) {
	id := c.Param("id")
	var media model.MediaFile
	if err := db.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File does not exist"})
		return
	}

	uid, _ := c.Get("userID")
	userID := uid.(uint)
	var user model.User
	if err := db.First(&user, userID).Error; err == nil {
		if user.Role != "admin" && media.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this file"})
			return
		}
	}

	// Delete file based on storage configuration
	if UseS3Storage && S3Cfg != nil {
		// Extract S3 key from URL
		key := ExtractS3Key(media.URL, S3Cfg)
		if key != "" {
			if err := DeleteFileFromS3(key); err != nil {
				// Log error but continue to delete database record
				fmt.Printf("Failed to delete S3 file: %v\n", err)
			}
		}
	} else {
		// Delete local file
		// media.URL format is "/uploads/filename", need to convert to actual path
		filename := filepath.Base(media.URL)
		path := filepath.Join(DataDir, "media", filename)
		os.Remove(path)
	}

	db.Delete(&media)
	c.JSON(http.StatusOK, gin.H{"message": "File deleted"})
}

// ExtractS3Key extracts the S3 object key from a URL
// This function is used by both upload and auth handlers
func ExtractS3Key(url string, cfg *config.S3Config) string {
	// URL format examples:
	// S3: https://bucket.s3.region.amazonaws.com/path/to/file.jpg
	// Custom domain: https://cdn.example.com/path/to/file.jpg
	// Path style: https://s3.amazonaws.com/bucket/path/to/file.jpg

	// Remove protocol
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	// Split by "/"
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}

	// If using custom domain, everything after the domain is the key
	if cfg.CustomDomain != "" {
		if len(parts) > 1 {
			return strings.Join(parts[1:], "/")
		}
		return ""
	}

	// For path-style URLs (ForcePath = true)
	if cfg.ForcePath {
		// Format: endpoint/bucket/key
		// e.g., minio.example.com/bucket-name/uuid.jpg
		if len(parts) >= 3 {
			// Skip endpoint and bucket
			return strings.Join(parts[2:], "/")
		}
		return ""
	}

	// For virtual-hosted style (default AWS S3)
	// Format: bucket.s3.region.amazonaws.com/key
	// or bucket.endpoint.com/key
	if len(parts) >= 2 {
		// Skip the first part (bucket.s3... or bucket)
		return strings.Join(parts[1:], "/")
	}

	return ""
}
