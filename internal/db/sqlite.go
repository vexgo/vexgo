package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewConnection(dsn string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Post{}, &Page{}, &User{}, &Category{},
		&Media{}, &Setting{}, &Plugin{}, &Theme{},
	)
}
