package main

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handleNews(c *gin.Context) {
	url := "https://cafef.vn/tin-tuc.rss"
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news data: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "CafeF RSS error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response data: " + err.Error()})
		return
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse XML: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rss.Channel.Items})
}
