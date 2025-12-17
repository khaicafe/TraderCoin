package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"tradercoin/backend/config"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetBittrexSymbols - Lấy danh sách symbols từ Bittrex API
func GetBittrexSymbols(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		apiURL := cfg.Exchanges.Bittrex.APIURL + "/markets"

		// Call Bittrex API to get markets
		resp, err := http.Get(apiURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Bittrex data"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Bittrex API error"})
			return
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Parse JSON
		var markets []struct {
			Symbol        string `json:"symbol"`
			BaseCurrency  string `json:"baseCurrencySymbol"`
			QuoteCurrency string `json:"quoteCurrencySymbol"`
			Status        string `json:"status"`
		}

		if err := json.Unmarshal(body, &markets); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		// Filter only ONLINE status with USDT quote
		var activeSymbols []map[string]interface{}
		for _, market := range markets {
			if market.Status == "ONLINE" && market.QuoteCurrency == "USDT" {
				activeSymbols = append(activeSymbols, map[string]interface{}{
					"symbol":      market.Symbol,
					"base_asset":  market.BaseCurrency,
					"quote_asset": market.QuoteCurrency,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(activeSymbols),
			"symbols": activeSymbols,
		})
	}
}
