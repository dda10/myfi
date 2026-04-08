package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleMacroIndicators serves GET /api/market/macro/indicators — full macro snapshot.
func (h *Handlers) HandleMacroIndicators(c *gin.Context) {
	if h.MacroService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "macro service not initialized"})
		return
	}

	data, err := h.MacroService.GetMacroIndicators(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleInterbankRates serves GET /api/market/macro/interbank — interbank lending rates.
func (h *Handlers) HandleInterbankRates(c *gin.Context) {
	if h.MacroService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "macro service not initialized"})
		return
	}

	rates, err := h.MacroService.GetInterbankRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// HandleBondYields serves GET /api/market/macro/bonds — government bond yields.
func (h *Handlers) HandleBondYields(c *gin.Context) {
	if h.MacroService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "macro service not initialized"})
		return
	}

	yields, err := h.MacroService.GetBondYields(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": yields})
}

// HandleFXRates serves GET /api/market/macro/fx — foreign exchange rates.
func (h *Handlers) HandleFXRates(c *gin.Context) {
	if h.MacroService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "macro service not initialized"})
		return
	}

	rates, err := h.MacroService.GetFXRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}
