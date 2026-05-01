package plugin

import (
	"github.com/gin-gonic/gin"
)

type Plugin interface {
	Name() string
	Init(api PluginAPI) error
	Handler() gin.HandlerFunc
	Middleare() gin.HandlerFunc
	Cleanup() error
}

type PluginAPI interface {
	RegisterRoute(method, path string, handler gin.HandlerFunc)
	GetSetting(key string) string
	SetSetting(key, value string) error
	DB() interface{}
	Log(message string)
}
