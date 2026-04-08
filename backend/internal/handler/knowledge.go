package handler

import (
	"myfi-backend/internal/domain/knowledge"

	"github.com/gin-gonic/gin"
)

// HandleGetObservations delegates to the knowledge domain handler.
// GET /api/knowledge/observations
func (h *Handlers) HandleGetObservations(c *gin.Context) {
	kh := &knowledge.Handlers{KnowledgeBase: h.KnowledgeBase}
	kh.HandleGetObservations(c)
}

// HandleGetAccuracy delegates to the knowledge domain handler.
// GET /api/knowledge/accuracy/:patternType
func (h *Handlers) HandleGetAccuracy(c *gin.Context) {
	kh := &knowledge.Handlers{KnowledgeBase: h.KnowledgeBase}
	kh.HandleGetAccuracy(c)
}
