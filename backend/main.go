package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/metrics/rate-limits", handleRateLimitMetrics)

	r.GET("/api/market/quote", handleMarketQuote)
	r.GET("/api/market/chart", handleMarketChart)
	r.GET("/api/market/listing", handleMarketListing)
	r.GET("/api/market/screener", handleMarketScreener)

	r.GET("/api/crypto/quote", handleCryptoQuote)
	r.GET("/api/prices/fx", handleFXRate)
	r.GET("/api/news", handleNews)
	r.POST("/api/chat", handleChat)
	r.POST("/api/models", handleModels)

	initLLM()
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
