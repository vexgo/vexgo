package api

import (
	"net/http"

	"go-cms/internal/cms"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service *cms.SettingsService
}

func NewSettingsHandler(s *cms.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: s}
}

func (h *SettingsHandler) GetAll(c *gin.Context) {
	settings, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range input {
		h.service.Set(k, v)
	}
	c.Status(http.StatusNoContent)
}
