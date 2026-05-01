package api

import (
	"net/http"

	"go-cms/internal/cms"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service *cms.MediaService
}

func NewMediaHandler(s *cms.MediaService) *MediaHandler {
	return &MediaHandler{service: s}
}

func (h *MediaHandler) List(c *gin.Context) {
	media, err := h.service.ListMedia()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": media})
}

func (h *MediaHandler) Upload(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "media uploaded - TODO"})
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	c.Status(http.StatusNoContent)
	_ = id
}
