package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"tradercoin/backend/models"
	"tradercoin/backend/utils"

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

func (s *TelegramService) SendMessageToUser(userID uint, message, parseMode string) error {
	var config models.TelegramConfig
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, "active").First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("telegram configuration not found or disabled for user")
		}
		return fmt.Errorf("failed to fetch telegram config: %w", err)
	}

	return s.SendMessage(config.BotToken, config.ChatID, message, parseMode)
}

// SendMessageToUser sends message to user by chatID
func (s *TelegramService) SendMessageToUserSignal(botToken, chatID, symbol, side string) error {
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

	baseLabel, callback := BuildTradeLabels(symbol, side)
	// 🎨 COMPREHENSIVE TEST MESSAGE WITH ALL FORMATTING STYLES
	message := fmt.Sprintf("\n************************************\n")
	message += "<b>🚨 New Trading Signal</b>\n\n"
	message += fmt.Sprintf("Symbol: <b>%s</b>\n", baseLabel)
	message += fmt.Sprintf("Side: <b>%s</b>\n", strings.ToUpper(side))
	message += fmt.Sprintf("\n")

	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = "HTML"

	// Xác định emoji dựa trên side
	emoji := "⚪️" // Emoji mặc định
	if strings.ToUpper(side) == "BUY" {
		emoji = "🟢"
	} else if strings.ToUpper(side) == "SELL" {
		emoji = "🔴"
	}
	// 🎮 COMPREHENSIVE INLINE KEYBOARD WITH MULTIPLE EXAMPLES
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// Row 4: Quick Trade Buttons (với callback_data)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(emoji+" "+baseLabel, callback),
		),
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

// BroadcastTestConnectionToAllUsers sends the test connection message
// to every user that has a non-empty chat_id in the users table.
func (s *TelegramService) BroadcastTestConnectionToAllUsers(botToken, symbol, side string) error {
	var users []models.User

	// Lọc user có chat_id khác rỗng và status = active (tùy DB schema của bạn)
	if err := s.db.
		Where("chat_id <> '' AND chat_id IS NOT NULL").
		Find(&users).Error; err != nil {
		return fmt.Errorf("failed to query users with chat_id: %w", err)
	}

	if len(users) == 0 {
		log.Printf("⚠️ No users with chat_id found to send test connection")
		return nil
	}

	log.Printf("📤 Broadcasting Telegram test connection to %d users", len(users))

	var firstErr error

	for _, u := range users {
		chatID := u.ChatID // đảm bảo field này tồn tại trong models.User
		if chatID == "" {
			continue
		}

		// Gửi test message giống TestConnection hiện tại
		if err := s.SendMessageToUserSignal(botToken, chatID, symbol, side); err != nil {
			log.Printf("❌ Failed to send test connection to userID=%d chatID=%s: %v", u.ID, chatID, err)
			if firstErr == nil {
				firstErr = err
			}
			// tiếp tục gửi cho user khác
			continue
		}

		log.Printf("✅ Test connection sent to userID=%d chatID=%s", u.ID, chatID)
	}

	return firstErr
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
			tgbotapi.NewInlineKeyboardButtonData("🟢 DOGE/USDT BUY", "trade_BUY_DOGEUSDT"),
			tgbotapi.NewInlineKeyboardButtonData("🔴 DOGE/USDT SELL", "trade_SELL_DOGEUSDT"),
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
				side := strings.ToLower(parts[1]) // buy hoặc sell
				symbol := parts[2]                // ETHUSDT, BTCUSDT, etc.

				// Lấy user ID từ Telegram chat ID
				userID, err := s.getUserIDFromChatID(callback.From.ID)
				if err != nil {
					responseText := fmt.Sprintf("❌ Không tìm thấy user. Lỗi: %v", err)
					msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
					msg.ParseMode = "HTML"
					bot.Send(msg)

					callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ User not found")
					bot.Request(callbackConfig)
					continue
				}

				// Đặt lệnh qua hàm PlaceOrderFromTelegram
				orderResult, err := s.PlaceOrderFromTelegram(userID, symbol, side, "market", 0, 0)
				if err != nil {
					responseText := fmt.Sprintf("❌ Lỗi đặt lệnh %s %s:\n<code>%v</code>", side, symbol, err)
					msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
					msg.ParseMode = "HTML"
					bot.Send(msg)

					callbackConfig := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("❌ Lỗi: %s", side))
					bot.Request(callbackConfig)
					continue
				}

				// Tạo response message với thông tin order
				responseText := "✅ <b>Đặt lệnh thành công!</b>\n\n"
				responseText += fmt.Sprintf("Symbol: <b>%s</b>\n", orderResult.Symbol)
				responseText += fmt.Sprintf("Side: <b>%s</b>\n", strings.ToUpper(orderResult.Side))
				responseText += fmt.Sprintf("Type: <b>%s</b>\n", strings.ToUpper(orderResult.Type))
				responseText += fmt.Sprintf("Quantity: <b>%v</b>\n", orderResult.Quantity)
				responseText += fmt.Sprintf("Price: <b>%v</b>\n", orderResult.FilledPrice)
				responseText += fmt.Sprintf("Status: <b>%s</b>\n", orderResult.Status)
				responseText += fmt.Sprintf("Order ID: <code>%s</code>", orderResult.OrderID)

				// Send confirmation message
				msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
				msg.ParseMode = "HTML"
				bot.Send(msg)

				// Answer callback query (tắt loading indicator)
				callbackConfig := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("✅ %s %s", side, symbol))
				bot.Request(callbackConfig)

				log.Printf("✅ Order placed: %s %s - OrderID: %s", side, symbol, orderResult.OrderID)
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

// getUserIDFromChatID lấy UserID từ Telegram Chat ID
func (s *TelegramService) getUserIDFromChatID(telegramUserID int64) (uint, error) {
	var config models.User
	// ChatID được lưu dạng string trong database
	chatIDStr := fmt.Sprintf("%d", telegramUserID)

	err := s.db.Where("chat_id = ? AND status = ?", chatIDStr, "active").First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("không tìm thấy cấu hình Telegram cho chat ID: %d", telegramUserID)
		}
		return 0, fmt.Errorf("lỗi truy vấn database: %w", err)
	}

	return config.ID, nil
}

// PlaceOrderFromTelegram đặt lệnh từ Telegram bot
func (s *TelegramService) PlaceOrderFromTelegram(userID uint, symbol, side, orderType string, amount, price float64) (*OrderResult, error) {
	// Lấy bot config đầu tiên của user (hoặc có thể lấy theo default)
	var config models.TradingConfig
	err := s.db.Where("user_id = ? AND is_active = ? AND symbol = ? AND is_default = ?", userID, true, symbol, true).
		First(&config).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("lỗi truy vấn bot config (chưa set bot config default)")
	}

	// Kiểm tra API credentials
	if config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("bot config thiếu API credentials")
	}

	// Sử dụng amount từ config nếu không được cung cấp
	if amount <= 0 {
		amount = config.Amount
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount phải lớn hơn 0")
	}

	// Giải mã API credentials
	apiKey, err := utils.DecryptString(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("lỗi giải mã API key: %w", err)
	}

	apiSecret, err := utils.DecryptString(config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("lỗi giải mã API secret: %w", err)
	}

	// Tạo trading service và đặt lệnh
	tradingService := NewTradingService(apiKey, apiSecret, config.Exchange, s.db, userID)
	orderResult := tradingService.PlaceOrder(&config, side, orderType, symbol, amount, price)

	if !orderResult.Success {
		errorMsg := orderResult.Error
		if errorMsg == "" {
			errorMsg = "Đặt lệnh thất bại"
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// Tính toán SL/TP prices
	var stopLoss, takeProfit float64
	filledPrice := orderResult.FilledPrice
	if filledPrice > 0 {
		if config.StopLossPercent > 0 {
			if side == "buy" {
				stopLoss = filledPrice * (1 - config.StopLossPercent/100)
			} else {
				stopLoss = filledPrice * (1 + config.StopLossPercent/100)
			}
		}

		if config.TakeProfitPercent > 0 {
			if side == "buy" {
				takeProfit = filledPrice * (1 + config.TakeProfitPercent/100)
			} else {
				takeProfit = filledPrice * (1 - config.TakeProfitPercent/100)
			}
		}
	}

	// Lưu order vào database
	order := models.Order{
		UserID:           userID,
		BotConfigID:      config.ID,
		Exchange:         config.Exchange,
		Symbol:           orderResult.Symbol,
		OrderID:          orderResult.OrderID,
		Side:             orderResult.Side,
		Type:             orderResult.Type,
		Quantity:         orderResult.Quantity,
		Price:            orderResult.Price,
		FilledPrice:      orderResult.FilledPrice,
		Status:           orderResult.Status,
		TradingMode:      config.TradingMode,
		Leverage:         config.Leverage,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		AlgoIDStopLoss:   orderResult.AlgoIDStopLoss,
		AlgoIDTakeProfit: orderResult.AlgoIDTakeProfit,
		PnL:              0,
		PnLPercent:       0,
	}

	if err := s.db.Create(&order).Error; err != nil {
		log.Printf("⚠️ Lỗi lưu order vào database: %v", err)
		// Không return error vì order đã được đặt thành công trên exchange
	}

	log.Printf("✅ Order từ Telegram đã được đặt: OrderID=%s, Symbol=%s, Side=%s, Amount=%f",
		orderResult.OrderID, orderResult.Symbol, orderResult.Side, orderResult.Quantity)

	return &orderResult, nil
}

// BuildTradeLabels nhận symbol (vd: DOGEUSDT) và side (vd: BUY)
// trả về:
//   - prettySymbol: "DOGE/USDT BUY"
//   - callbackData: "trade_BUY_DOGEUSDT"
func BuildTradeLabels(symbol, side string) (prettySymbol, callbackData string) {
	// Danh sách các quote phổ biến, ưu tiên length dài hơn trước
	quotes := []string{
		"USDT", "BUSD", "FDUSD", "TUSD",
		"BTC", "ETH", "BNB",
		"USDC", "DAI",
	}

	base := symbol
	quote := ""

	// Tìm quote bằng cách check hậu tố của symbol
	for _, q := range quotes {
		if strings.HasSuffix(strings.ToUpper(symbol), q) && len(symbol) > len(q) {
			base = symbol[:len(symbol)-len(q)]
			quote = q
			break
		}
	}

	// Nếu không tìm được quote, giữ nguyên symbol
	if quote == "" {
		prettySymbol = fmt.Sprintf("%s %s", strings.ToUpper(symbol), strings.ToUpper(side))
	} else {
		prettySymbol = fmt.Sprintf("%s/%s %s", strings.ToUpper(base), strings.ToUpper(quote), strings.ToUpper(side))
	}

	callbackData = fmt.Sprintf("trade_%s_%s", strings.ToUpper(side), strings.ToUpper(symbol))
	return
}
