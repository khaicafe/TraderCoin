package controllers

import (
	"log"
	"tradercoin/backend/models"
	"tradercoin/backend/utils"
)

// GetDecryptedAPICredentials - Helper function to get decrypted API credentials from TradingConfig
func GetDecryptedAPICredentials(config *models.TradingConfig) (apiKey string, apiSecret string, err error) {
	if config.APIKey != "" {
		apiKey, err = utils.DecryptString(config.APIKey)
		if err != nil {
			log.Printf("Failed to decrypt API key for config %d: %v", config.ID, err)
			return "", "", err
		}
	}

	if config.APISecret != "" {
		apiSecret, err = utils.DecryptString(config.APISecret)
		if err != nil {
			log.Printf("Failed to decrypt API secret for config %d: %v", config.ID, err)
			return "", "", err
		}
	}

	return apiKey, apiSecret, nil
}
