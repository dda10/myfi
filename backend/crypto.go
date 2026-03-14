package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleCryptoQuote(c *gin.Context) {
	ids := c.Query("ids")
	if ids == "" {
		ids = "bitcoin,ethereum,binancecoin"
	}

	url := "https://api.coingecko.com/api/v3/simple/price?ids=" + ids + "&vs_currencies=usd,vnd&include_24hr_change=true"
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch crypto data: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "CoinGecko API error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response data: " + err.Error()})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
