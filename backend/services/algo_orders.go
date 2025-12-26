package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"tradercoin/backend/models"
)

// PlaceAlgoStopLoss places a stop loss via standard Binance /fapi/v1/order endpoint
func (ts *TradingService) PlaceAlgoStopLoss(
	config *models.TradingConfig,
	symbol string,
	stopPrice float64,
	side string, // closing side: SELL để đóng LONG, BUY để đóng SHORT
	positionSide string, // LONG / SHORT (bắt buộc nếu Hedge Mode) HOẶC "BOTH" nếu One-way Mode
) OrderResult {

	// ====== INIT ======
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/algoOrder" // Endpoint đúng

	// ====== BUILD PARAMS ======
	params := url.Values{}
	params.Set("algoType", "CONDITIONAL") // Bắt buộc
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))                  // SELL cho đóng LONG
	params.Set("type", "STOP_MARKET")                          // type=STOP_MARKET
	params.Set("triggerPrice", fmt.Sprintf("%.2f", stopPrice)) // triggerPrice + precision phù hợp ETHUSDT (tickSize 0.01)

	params.Set("closePosition", "true") // Đóng toàn bộ vị thế khi trigger
	// ⭐ positionSide: Chỉ gửi nếu tài khoản đang ở Hedge Mode
	// Nếu tài khoản ở One-way Mode, KHÔNG gửi param này (hoặc gửi "BOTH")
	// Vì lỗi -4061 xảy ra khi gửi positionSide=LONG/SHORT trong khi tài khoản ở One-way Mode
	// Giải pháp: Kiểm tra mode trước, hoặc tạm thời thử KHÔNG gửi (cho One-way) hoặc gửi "BOTH"

	// Tạm thời fix bằng cách KHÔNG gửi positionSide (tương đương BOTH, phù hợp One-way Mode)
	// Nếu bạn dùng Hedge Mode, uncomment dòng dưới và đảm bảo positionSide đúng (LONG/SHORT)
	// params.Set("positionSide", strings.ToUpper(positionSide))

	params.Set("workingType", "MARK_PRICE") // Trigger theo Mark Price (an toàn hơn)
	params.Set("priceProtect", "TRUE")      // Uppercase như docs

	// ====== AUTH ======
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	query := params.Encode()
	signature := ts.sign(query)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	// ====== DEBUG REQUEST ======
	fmt.Println("\n========== BINANCE STOP LOSS REQUEST ==========")
	fmt.Println("URL:", baseURL+endpoint)
	fmt.Println("METHOD: POST")
	fmt.Println("QUERY:", params.Encode())
	fmt.Println("PARAMS:")
	for k, v := range params {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Println("==============================================")

	// ====== SEND REQUEST ======
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// ====== DEBUG RESPONSE ======
	fmt.Println("\n========== BINANCE STOP LOSS RESPONSE =========")
	fmt.Println("HTTP STATUS:", resp.StatusCode)
	fmt.Println("RAW BODY:", string(body))
	fmt.Println("==============================================")

	// ====== HANDLE ERROR ======
	if resp.StatusCode != http.StatusOK {
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Stop loss order failed (HTTP %d): %s", resp.StatusCode, string(body)),
		}
	}

	// ====== PARSE SUCCESS RESPONSE ======
	var orderResp struct {
		AlgoID       int64  `json:"algoId"`
		Symbol       string `json:"symbol"`
		TriggerPrice string `json:"triggerPrice"`
		AlgoStatus   string `json:"algoStatus"`
		OrderType    string `json:"orderType"`
	}

	if err := json.Unmarshal(body, &orderResp); err != nil {
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse response: %v - Body: %s", err, string(body)),
		}
	}

	// ====== SUCCESS ======
	return OrderResult{
		Success: true,
		OrderID: strconv.FormatInt(orderResp.AlgoID, 10),
		Symbol:  orderResp.Symbol,
	}
}

// PlaceAlgoTakeProfit places a take profit via standard Binance /fapi/v1/order endpoint
func (ts *TradingService) PlaceAlgoTakeProfit(
	config *models.TradingConfig,
	symbol string,
	takeProfitPrice float64,
	side string, // closing side: SELL để đóng LONG, BUY để đóng SHORT
	positionSide string, // LONG / SHORT (bắt buộc nếu Hedge Mode) HOẶC "BOTH" nếu One-way Mode
) OrderResult {

	// ====== INIT ======
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/algoOrder" // Endpoint đúng - giống Stop Loss

	// ====== BUILD PARAMS ======
	params := url.Values{}
	params.Set("algoType", "CONDITIONAL") // Bắt buộc
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))                        // SELL cho đóng LONG, BUY cho đóng SHORT
	params.Set("type", "TAKE_PROFIT_MARKET")                         // type=TAKE_PROFIT_MARKET
	params.Set("triggerPrice", fmt.Sprintf("%.2f", takeProfitPrice)) // triggerPrice với precision phù hợp

	params.Set("closePosition", "true") // Đóng toàn bộ vị thế khi trigger
	// ⭐ positionSide: Chỉ gửi nếu tài khoản đang ở Hedge Mode
	// Nếu tài khoản ở One-way Mode, KHÔNG gửi param này (hoặc gửi "BOTH")
	// Tạm thời fix bằng cách KHÔNG gửi positionSide (tương đương BOTH, phù hợp One-way Mode)
	// Nếu bạn dùng Hedge Mode, uncomment dòng dưới và đảm bảo positionSide đúng (LONG/SHORT)
	// params.Set("positionSide", strings.ToUpper(positionSide))

	params.Set("workingType", "MARK_PRICE") // Trigger theo Mark Price (an toàn hơn)
	params.Set("priceProtect", "TRUE")      // Uppercase như docs

	// ====== AUTH ======
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	query := params.Encode()
	signature := ts.sign(query)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	// ====== DEBUG REQUEST ======
	fmt.Println("\n========== BINANCE TAKE PROFIT REQUEST ==========")
	fmt.Println("URL:", baseURL+endpoint)
	fmt.Println("METHOD: POST")
	fmt.Println("QUERY:", params.Encode())
	fmt.Println("PARAMS:")
	for k, v := range params {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Println("==============================================")

	// ====== SEND REQUEST ======
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// ====== DEBUG RESPONSE ======
	fmt.Println("\n========== BINANCE TAKE PROFIT RESPONSE =========")
	fmt.Println("HTTP STATUS:", resp.StatusCode)
	fmt.Println("RAW BODY:", string(body))
	fmt.Println("==============================================")

	// ====== HANDLE ERROR ======
	if resp.StatusCode != http.StatusOK {
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Take profit order failed (HTTP %d): %s", resp.StatusCode, string(body)),
		}
	}

	// ====== PARSE SUCCESS RESPONSE ======
	var orderResp struct {
		AlgoID       int64  `json:"algoId"`
		Symbol       string `json:"symbol"`
		TriggerPrice string `json:"triggerPrice"`
		AlgoStatus   string `json:"algoStatus"`
		OrderType    string `json:"orderType"`
	}

	if err := json.Unmarshal(body, &orderResp); err != nil {
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse response: %v - Body: %s", err, string(body)),
		}
	}

	// ====== SUCCESS ======
	fmt.Printf("✅ TAKE PROFIT ORDER PLACED:\n")
	fmt.Printf("   AlgoID: %d\n", orderResp.AlgoID)
	fmt.Printf("   Symbol: %s\n", orderResp.Symbol)
	fmt.Printf("   TriggerPrice: %s\n", orderResp.TriggerPrice)
	fmt.Printf("   AlgoStatus: %s\n", orderResp.AlgoStatus)
	fmt.Printf("   OrderType: %s\n\n", orderResp.OrderType)

	return OrderResult{
		Success: true,
		OrderID: strconv.FormatInt(orderResp.AlgoID, 10),
		Symbol:  orderResp.Symbol,
	}
}
