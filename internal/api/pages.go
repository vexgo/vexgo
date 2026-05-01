package api

import (
	"net/http"
	"strconv"

	"go-cms/internal/cms"

	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	service *cms.ContentService
}

func NewPageHandler(s *cms.ContentService) *PageHandler {
	return &PageHandler{service: s}
}

func (h *PageHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pages list - TODO"})
}

func (h *PageHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	c.JSON(http.StatusOK, gin.H{"slug": slug, "message": "page details - TODO"})
}

func (h *PageHandler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "page created - TODO"})
}

func (h *PageHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "page updated - TODO"})
}

func (h *PageHandler) Delete(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
