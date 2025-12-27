package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"tradercoin/backend/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type TelegramService struct {
	db *gorm.DB
}

func NewTelegramService(db *gorm.DB) *TelegramService {
	return &TelegramService{db: db}
}

// SendMessage sends a simple text message
func (s *TelegramService) SendMessage(botToken, chatID, message, parseMode string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = parseMode

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendMessageWithButtons sends message with inline keyboard
func (s *TelegramService) SendMessageWithButtons(botToken, chatID, message, parseMode string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = parseMode
	msg.ReplyMarkup = keyboard

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendPhoto sends a photo with caption
func (s *TelegramService) SendPhoto(botToken, chatID, photoURL, caption string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewPhoto(chatIDInt, tgbotapi.FileURL(photoURL))
	msg.Caption = caption
	msg.ParseMode = "HTML"

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}

	return nil
}

// SendDocument sends a file/document
func (s *TelegramService) SendDocument(botToken, chatID, fileURL, caption string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewDocument(chatIDInt, tgbotapi.FileURL(fileURL))
	msg.Caption = caption

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}

	return nil
}

// SendMessageToUser sends message to user by user_id
func (s *TelegramService) SendMessageToUser(userID uint, message, parseMode string) error {
	var config models.TelegramConfig
	if err := s.db.Where("user_id = ? AND is_enabled = ?", userID, true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("telegram configuration not found or disabled for user")
		}
		return fmt.Errorf("failed to fetch telegram config: %w", err)
	}

	return s.SendMessage(config.BotToken, config.ChatID, message, parseMode)
}

// SendBotStatus sends bot status update notification
func (s *TelegramService) SendBotStatus(userID uint, botName, status, details string) error {
	message := "🤖 <b>Bot Status Update</b>\n\n"
	message += fmt.Sprintf("Bot: <b>%s</b>\n", botName)
	message += fmt.Sprintf("Status: <b>%s</b>\n", status)
	if details != "" {
		message += fmt.Sprintf("Details: %s", details)
	}

	return s.SendMessageToUser(userID, message, "HTML")
}

// SendOrderNotification sends order notification
func (s *TelegramService) SendOrderNotification(userID uint, orderInfo map[string]interface{}) error {
	message := "📊 <b>Order Notification</b>\n\n"

	if symbol, ok := orderInfo["symbol"].(string); ok {
		message += fmt.Sprintf("Symbol: <b>%s</b>\n", symbol)
	}

	if side, ok := orderInfo["side"].(string); ok {
		message += fmt.Sprintf("Side: <b>%s</b>\n", side)
	}

	if amount, ok := orderInfo["amount"]; ok {
		message += fmt.Sprintf("Amount: <b>%v</b>\n", amount)
	}

	if status, ok := orderInfo["status"].(string); ok {
		message += fmt.Sprintf("Status: <b>%s</b>\n", status)
	}

	if price, ok := orderInfo["price"]; ok {
		message += fmt.Sprintf("Price: <b>%v</b>\n", price)
	}

	if errorMsg, ok := orderInfo["error"].(string); ok && errorMsg != "" {
		message += fmt.Sprintf("❌ Error: %s", errorMsg)
	}

	return s.SendMessageToUser(userID, message, "HTML")
}

// SendErrorAlert sends error alert notification
func (s *TelegramService) SendErrorAlert(userID uint, botName, errorMessage string) error {
	message := "⚠️ <b>Error Alert</b>\n\n"
	message += fmt.Sprintf("Bot: <b>%s</b>\n", botName)
	message += fmt.Sprintf("Error: <code>%s</code>", errorMessage)

	return s.SendMessageToUser(userID, message, "HTML")
}

// SendBotPausedAlert sends alert when bot is paused
func (s *TelegramService) SendBotPausedAlert(userID uint, botName, reason string) error {
	message := "⏸️ <b>Bot Paused</b>\n\n"
	message += fmt.Sprintf("Bot: <b>%s</b>\n", botName)
	if reason != "" {
		message += fmt.Sprintf("Reason: %s", reason)
	}

	return s.SendMessageToUser(userID, message, "HTML")
}

// SendTradeNotification sends trade execution notification
func (s *TelegramService) SendTradeNotification(userID uint, tradeInfo map[string]interface{}) error {
	message := "💰 <b>Trade Executed</b>\n\n"

	if symbol, ok := tradeInfo["symbol"].(string); ok {
		message += fmt.Sprintf("Symbol: <b>%s</b>\n", symbol)
	}

	if side, ok := tradeInfo["side"].(string); ok {
		message += fmt.Sprintf("Side: <b>%s</b>\n", strings.ToUpper(side))
	}

	if quantity, ok := tradeInfo["quantity"]; ok {
		message += fmt.Sprintf("Quantity: <b>%v</b>\n", quantity)
	}

	if price, ok := tradeInfo["price"]; ok {
		message += fmt.Sprintf("Price: <b>%v</b>\n", price)
	}

	if total, ok := tradeInfo["total"]; ok {
		message += fmt.Sprintf("Total: <b>%v</b>\n", total)
	}

	if fee, ok := tradeInfo["fee"]; ok {
		message += fmt.Sprintf("Fee: <b>%v</b>", fee)
	}

	return s.SendMessageToUser(userID, message, "HTML")
}

// TestConnection tests the Telegram bot connection with comprehensive examples
func (s *TelegramService) TestConnection(botToken, chatID string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("❌ Failed to create bot: %v", err)
		return fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		log.Printf("❌ Invalid chat ID: %v", err)
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	log.Printf("📤 Sending comprehensive test message to Telegram")
	log.Printf("   Bot Username: @%s", bot.Self.UserName)
	log.Printf("   Chat ID: %s", chatID)

	// 🎨 COMPREHENSIVE TEST MESSAGE WITH ALL FORMATTING STYLES
	message := "✅ <b>Telegram Bot Connected Successfully!</b>\n\n"

	// Text Formatting Examples
	// message += "📝 <b>Text Formatting Examples:</b>\n"
	// message += "• <b>Bold Text</b>\n"
	// message += "• <i>Italic Text</i>\n"
	// message += "• <u>Underlined Text</u>\n"
	// message += "• <s>Strikethrough Text</s>\n"
	// message += "• <code>Inline Code</code>\n"
	// message += "• <pre>Preformatted Code Block</pre>\n"
	// message += "• <a href='https://tradercoin.com'>Hyperlink</a>\n\n"

	// // Emoji Examples
	// message += "🎯 <b>Emoji Examples:</b>\n"
	// message += "💰 📊 📈 📉 🚀 ⚡ 🔥 💎 ⭐ ✅ ❌ ⚠️ 🔔 🎉 🤖 💸\n\n"

	// // Trading Notification Example
	// message += "📊 <b>Sample Trade Notification:</b>\n"
	// message += "━━━━━━━━━━━━━━━━\n"
	// message += "Symbol: <b>BTC/USDT</b>\n"
	// message += "Side: <b>🟢 LONG</b>\n"
	// message += "Entry: <code>$45,000.00</code>\n"
	// message += "Amount: <b>0.5 BTC</b>\n"
	// message += "Stop Loss: <code>$44,000.00</code> (-2.22%)\n"
	// message += "Take Profit: <code>$47,000.00</code> (+4.44%)\n"
	// message += "Status: <b>✅ FILLED</b>\n"
	// message += "━━━━━━━━━━━━━━━━\n\n"

	// // PnL Example
	// message += "💰 <b>P&amp;L Update:</b>\n"
	// message += "Current Price: <code>$45,500.00</code>\n"
	// message += "Unrealized P&amp;L: <b>+$250.00</b> 🟢 (+1.11%)\n\n"

	// // Alert Example
	// message += "⚠️ <b>Sample Alert:</b>\n"
	// message += "Stop Loss triggered at $44,100.00\n"
	// message += "Loss: <code>-$450.00</code> ❌ (-2.00%)\n\n"

	// // Bot Status Example
	// message += "🤖 <b>Bot Status:</b>\n"
	// message += "Status: <b>🟢 ACTIVE</b>\n"
	// message += "Runtime: <code>5h 23m</code>\n"
	// message += "Total Trades: <b>12</b>\n"
	// message += "Win Rate: <b>75%</b> (9W / 3L)\n"
	// message += "Total P&amp;L: <b>+$1,250.00</b> 🚀\n\n"

	// message += "🎉 <i>Your bot is ready to send notifications!</i>\n"
	// message += "Click the buttons below to explore 👇"

	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = "HTML"

	// 🎮 COMPREHENSIVE INLINE KEYBOARD WITH MULTIPLE EXAMPLES
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// // Row 1: Documentation & GitHub
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonURL("📚 Documentation", "https://github.com/khaicafe/TraderCoin"),
		// 	tgbotapi.NewInlineKeyboardButtonURL("⭐ Star on GitHub", "https://github.com/khaicafe/TraderCoin/stargazers"),
		// ),
		// // Row 2: Trading Links
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonURL("📊 Binance", "https://www.binance.com"),
		// 	tgbotapi.NewInlineKeyboardButtonURL("💹 TradingView", "https://www.tradingview.com"),
		// ),
		// // Row 3: Crypto News
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonURL("📰 CoinDesk", "https://www.coindesk.com"),
		// 	tgbotapi.NewInlineKeyboardButtonURL("🔍 CoinGecko", "https://www.coingecko.com"),
		// ),
		// Row 4: Quick Trade Buttons (với callback_data)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 ETH/USDT BUY", "trade_BUY_ETHUSDT"),
			tgbotapi.NewInlineKeyboardButtonData("🔴 ETH/USDT SELL", "trade_SELL_ETHUSDT"),
		),
		// Row 5: More Trade Buttons
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonData("🟢 BTCUSDT BUY", "trade_BUY_BTCUSDT"),
		// 	tgbotapi.NewInlineKeyboardButtonData("🔴 BTCUSDT SELL", "trade_SELL_BTCUSDT"),
		// ),
	)
	msg.ReplyMarkup = keyboard

	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("❌ Failed to send test message: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("✅ Telegram connection test successful!")
	log.Printf("   Message ID: %d", sentMsg.MessageID)
	log.Printf("   Sent at: %v", sentMsg.Date)

	return nil
}

// HandleUpdates listens for callback queries from Telegram buttons
func (s *TelegramService) HandleUpdates(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = true
	log.Printf("🤖 Bot authorized on account: @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("🔔 Listening for callback queries...")

	for update := range updates {
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			log.Printf("📥 Received callback: %s from user %d", callback.Data, callback.From.ID)

			// Parse callback data: trade_BUY_ETHUSDT hoặc trade_SELL_BTCUSDT
			parts := strings.Split(callback.Data, "_")
			if len(parts) == 3 && parts[0] == "trade" {
				side := parts[1]   // BUY hoặc SELL
				symbol := parts[2] // ETHUSDT, BTCUSDT, etc.

				// Tạo response message
				responseText := fmt.Sprintf("✅ Đã nhận lệnh %s %s!\n\n", side, symbol)
				responseText += "⚠️ <i>Đây là test mode, chưa thực hiện giao dịch thật.</i>\n\n"
				responseText += "Để thực hiện giao dịch thật, bạn cần:\n"
				responseText += "1. Cấu hình Exchange API Key\n"
				responseText += "2. Có đủ số dư trong tài khoản\n"
				responseText += "3. Bật chế độ live trading"

				// Send confirmation message
				msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
				msg.ParseMode = "HTML"
				bot.Send(msg)

				// Answer callback query (tắt loading indicator)
				callbackConfig := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("✅ %s %s", side, symbol))
				bot.Request(callbackConfig)

				log.Printf("✅ Processed: %s %s", side, symbol)
			} else {
				// Unknown callback data
				callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ Unknown command")
				bot.Request(callbackConfig)
				log.Printf("❌ Unknown callback: %s", callback.Data)
			}
		}
	}

	return nil
}
