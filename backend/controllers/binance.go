package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetBinanceFuturesSymbols - Lấy danh sách symbols từ Binance Futures API
func GetBinanceFuturesSymbols(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call Binance Futures API to get exchange info
		resp, err := http.Get("https://fapi.binance.com/fapi/v1/exchangeInfo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Binance data"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Binance API error"})
			return
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Parse JSON
		var exchangeInfo struct {
			Symbols []struct {
				Symbol            string `json:"symbol"`
				Pair              string `json:"pair"`
				ContractType      string `json:"contractType"`
				Status            string `json:"status"`
				BaseAsset         string `json:"baseAsset"`
				QuoteAsset        string `json:"quoteAsset"`
				MarginAsset       string `json:"marginAsset"`
				PricePrecision    int    `json:"pricePrecision"`
				QuantityPrecision int    `json:"quantityPrecision"`
			} `json:"symbols"`
		}

		if err := json.Unmarshal(body, &exchangeInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		// Filter only PERPETUAL contracts with TRADING status and USDT quote
		var activeSymbols []map[string]interface{}
		for _, symbol := range exchangeInfo.Symbols {
			if symbol.Status == "TRADING" &&
				symbol.ContractType == "PERPETUAL" &&
				symbol.QuoteAsset == "USDT" {
				activeSymbols = append(activeSymbols, map[string]interface{}{
					"symbol":             symbol.Symbol,
					"pair":               symbol.Pair,
					"base_asset":         symbol.BaseAsset,
					"quote_asset":        symbol.QuoteAsset,
					"price_precision":    symbol.PricePrecision,
					"quantity_precision": symbol.QuantityPrecision,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(activeSymbols),
			"symbols": activeSymbols,
		})
	}
}

// GetBinanceSpotSymbols - Lấy danh sách symbols từ Binance Spot API
func GetBinanceSpotSymbols(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call Binance Spot API to get exchange info
		resp, err := http.Get("https://api.binance.com/api/v3/exchangeInfo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Binance data"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Binance API error"})
			return
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Parse JSON
		var exchangeInfo struct {
			Symbols []struct {
				Symbol     string `json:"symbol"`
				Status     string `json:"status"`
				BaseAsset  string `json:"baseAsset"`
				QuoteAsset string `json:"quoteAsset"`
			} `json:"symbols"`
		}

		if err := json.Unmarshal(body, &exchangeInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		// Filter only TRADING status with USDT quote
		var activeSymbols []map[string]interface{}
		for _, symbol := range exchangeInfo.Symbols {
			if symbol.Status == "TRADING" && symbol.QuoteAsset == "USDT" {
				activeSymbols = append(activeSymbols, map[string]interface{}{
					"symbol":      symbol.Symbol,
					"base_asset":  symbol.BaseAsset,
					"quote_asset": symbol.QuoteAsset,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(activeSymbols),
			"symbols": activeSymbols,
		})
	}
}
