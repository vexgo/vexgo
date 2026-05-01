package cms

import (
	"go-cms/internal/db"
	"gorm.io/gorm"
)

type MediaService struct {
	db *gorm.DB
}

func NewMediaService(database *gorm.DB) *MediaService {
	return &MediaService{db: database}
}

func (s *MediaService) CreateMedia(filename, url, mimeType string, size int64) (*db.Media, error) {
	media := &db.Media{
		Filename: filename,
		URL:      url,
		MimeType: mimeType,
		Size:     size,
	}
	if err := s.db.Create(media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (s *MediaService) ListMedia() ([]db.Media, error) {
	var media []db.Media
	if err := s.db.Order("created_at DESC").Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (s *MediaService) DeleteMedia(id uint) error {
	return s.db.Delete(&db.Media{}, id).Error
}
