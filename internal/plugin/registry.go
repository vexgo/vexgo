package plugin

import (
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Registry struct {
	loader *Loader
	db     *gorm.DB
	mu     sync.RWMutex
}

func NewRegistry(loader *Loader, database *gorm.DB) *Registry {
	return &Registry{
		loader: loader,
		db:     database,
	}
}

func (r *Registry) EnablePlugin(name string) error {
	p, ok := r.loader.GetPlugin(name)
	if !ok {
		return errors.New("plugin not found: " + name)
	}
	api := &pluginAPI{
		db:     r.db,
		settings: make(map[string]string),
	}
	if err := p.Init(api); err != nil {
		return err
	}
	return nil
}

type pluginAPI struct {
	db       *gorm.DB
	settings map[string]string
}

func (a *pluginAPI) RegisterRoute(method, path string, handler gin.HandlerFunc) {
	// Will be used by Gin route registration
}

func (a *pluginAPI) GetSetting(key string) string {
	return a.settings[key]
}

func (a *pluginAPI) SetSetting(key, value string) error {
	a.settings[key] = value
	return nil
}

func (a *pluginAPI) DB() interface{} {
	return a.db
}

func (a *pluginAPI) Log(message string) {
	// Simple log - in production use proper logger
}
