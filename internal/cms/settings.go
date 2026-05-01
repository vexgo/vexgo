package cms

import (
	"go-cms/internal/db"
	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

func NewSettingsService(database *gorm.DB) *SettingsService {
	return &SettingsService{db: database}
}

func (s *SettingsService) Get(key string) (string, error) {
	var setting db.Setting
	if err := s.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingsService) Set(key, value string) error {
	var setting db.Setting
	result := s.db.Where("key = ?", key).First(&setting)
	if result.Error == gorm.ErrRecordNotFound {
		setting = db.Setting{Key: key, Value: value}
		return s.db.Create(&setting).Error
	}
	setting.Value = value
	return s.db.Save(&setting).Error
}

func (s *SettingsService) GetAll() (map[string]string, error) {
	var settings []db.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	return m, nil
}
