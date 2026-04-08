package market

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HandleSearch serves GET /api/market/search — global search (⌘K).
// Query params: q (search query), limit (max results, default 10).
func (h *Handlers) HandleSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	if h.SearchService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search service not initialized"})
		return
	}

	results, err := h.SearchService.Search(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results, "total": len(results)})
}
