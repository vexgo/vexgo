package db

import (
	"time"
	"gorm.io/gorm"
)

type Post struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Slug       string         `gorm:"uniqueIndex;size:255;not null" json:"slug"`
	Content    string         `gorm:"type:text" json:"content"`
	Excerpt    string         `gorm:"type:text" json:"excerpt"`
	Status     string         `gorm:"size:20;default:published" json:"status"`
	MetaTitle  string         `gorm:"size:255" json:"meta_title"`
	MetaDesc   string         `gorm:"type:text" json:"meta_description"`
	OGImage    string         `gorm:"size:500" json:"og_image"`
	Tags       string         `gorm:"type:text" json:"tags"`
	CategoryID uint           `json:"category_id"`
	AuthorID   uint           `json:"author_id"`
	PublishedAt *time.Time    `json:"published_at"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type Page struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Slug       string         `gorm:"uniqueIndex;size:255;not null" json:"slug"`
	Content    string         `gorm:"type:text" json:"content"`
	Status     string         `gorm:"size:20;default:published" json:"status"`
	MetaTitle  string         `gorm:"size:255" json:"meta_title"`
	MetaDesc   string         `gorm:"type:text" json:"meta_description"`
	OGImage    string         `gorm:"size:500" json:"og_image"`
	Template   string         `gorm:"size:100" json:"template"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Email     string    `gorm:"size:255" json:"email"`
	Role      string    `gorm:"size:20;default:editor" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:100;not null" json:"name"`
	Slug string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
}

type Media struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Filename  string    `gorm:"size:255;not null" json:"filename"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	MimeType  string    `gorm:"size:100" json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Key   string `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

type Plugin struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"size:100;not null" json:"name"`
	File    string `gorm:"size:255;not null" json:"file"`
	Enabled bool   `gorm:"default:false" json:"enabled"`
	Config  string `gorm:"type:text" json:"config"`
}

type Theme struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"size:100;not null" json:"name"`
	Active bool   `gorm:"default:false" json:"active"`
}
