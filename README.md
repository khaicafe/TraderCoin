# TraderCoin - Automated Crypto Trading Platform

Hệ thống giao dịch tiền điện tử tự động với quản lý stop-loss và take-profit.

## 📁 Cấu trúc dự án

```
TraderCoin/
├── Backend/          # Golang Backend (Gin + SQLite)
├── frontend/         # Next.js User Frontend
├── backoffice/       # Next.js Admin Backoffice
└── README.md
```

## 🚀 Tính năng

### Frontend (User Portal)

- ✅ Đăng ký/Đăng nhập tài khoản
- ✅ Quản lý API Key các sàn (Binance, Bittrex)
- ✅ Cấu hình Stop Loss / Take Profit
- ✅ Theo dõi danh sách coin theo sàn
- ✅ Xem lịch sử giao dịch
- ✅ Quản lý đăng ký/gia hạn

### Backoffice (Admin Portal)

- ✅ Quản lý user (khóa/mở khóa)
- ✅ Quản lý subscription (gia hạn)
- ✅ Xem lịch sử giao dịch coin
- ✅ Quản lý giao dịch nạp tiền
- ✅ Dashboard thống kê

### Backend (API Server)

- ✅ RESTful API với Gin framework
- ✅ SQLite database
- ✅ Redis caching
- ✅ JWT Authentication
- ✅ Tích hợp API Binance, Bittrex
- ✅ WebSocket real-time updates
- ✅ Automated trading engine

## 🛠️ Tech Stack

### Backend

- **Language**: Golang 1.21+
- **Framework**: Gin
- **Database**: SQLite
- **Cache**: Redis
- **Authentication**: JWT

### Frontend & Backoffice

- **Framework**: Next.js 14
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React Context/Hooks
- **HTTP Client**: Axios

## 📦 Installation

### Prerequisites

- Go 1.21+
- Node.js 18+
- Redis Server
- SQLite3

### Backend Setup

```bash
cd Backend

# Install dependencies
go mod download

# Create .env file
cp .env.example .env

# Edit .env with your config
nano .env

# Run migrations and start server
go run main.go
```

Server sẽ chạy tại: `http://localhost:8080`

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Create .env.local
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# Run development server
npm run dev
```

Frontend sẽ chạy tại: `http://localhost:3000`

### Backoffice Setup

```bash
cd backoffice

# Install dependencies
npm install

# Create .env.local
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# Run development server
npm run dev
```

Backoffice sẽ chạy tại: `http://localhost:3001`

## 📋 API Documentation

### Authentication

- `POST /api/auth/register` - Đăng ký user mới
- `POST /api/auth/login` - Đăng nhập
- `POST /api/auth/refresh` - Refresh token

### User

- `GET /api/user/profile` - Lấy thông tin user
- `PUT /api/user/profile` - Cập nhật profile

### Exchange Keys

- `GET /api/exchange-keys` - Lấy danh sách API keys
- `POST /api/exchange-keys` - Thêm API key mới
- `PUT /api/exchange-keys/:id` - Cập nhật API key
- `DELETE /api/exchange-keys/:id` - Xóa API key

### Trading Config

- `GET /api/trading-configs` - Lấy cấu hình trading
- `POST /api/trading-configs` - Tạo cấu hình mới
- `PUT /api/trading-configs/:id` - Cập nhật cấu hình
- `DELETE /api/trading-configs/:id` - Xóa cấu hình

### Orders

- `GET /api/orders` - Lấy lịch sử orders
- `GET /api/orders/:id` - Chi tiết order

### Admin (Backoffice)

- `GET /api/admin/users` - Danh sách users
- `PUT /api/admin/users/:id/status` - Cập nhật trạng thái user
- `GET /api/admin/transactions` - Danh sách transactions
- `GET /api/admin/statistics` - Thống kê

## 🔐 Environment Variables

### Backend (.env)

```env
PORT=8080
DB_PATH=./tradercoin.db

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_SECRET=your_secret_key_here
JWT_EXPIRATION=24h

BINANCE_API_URL=https://api.binance.com
BITTREX_API_URL=https://api.bittrex.com
```

### Frontend (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

### Backoffice (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=TraderCoin Admin
```

## 📊 Database Schema

### Users

- id, email, password_hash, full_name, phone
- status (active, suspended, expired)
- subscription_end
- created_at, updated_at

### Exchange Keys

- id, user_id, exchange, api_key, api_secret
- is_active, created_at, updated_at

### Trading Configs

- id, user_id, exchange, symbol
- stop_loss_percent, take_profit_percent
- is_active, created_at, updated_at

### Orders

- id, user_id, exchange, symbol, order_id
- side, type, quantity, price, status
- created_at, updated_at

### Transactions

- id, user_id, amount, type, status
- description, created_at

### Admins

- id, email, password_hash, full_name
- role, created_at

## 🚦 Running in Production

### Backend

```bash
# Build
go build -o tradercoin-api main.go

# Run
./tradercoin-api
```

### Frontend & Backoffice

```bash
# Build
npm run build

# Start
npm start
```

## 📝 Development Workflow

1. Backend API development: `Backend/`
2. Frontend development: `frontend/`
3. Backoffice development: `backoffice/`
4. Test API endpoints
5. Integrate with exchange APIs
6. Deploy to production

## 🔧 Troubleshooting

### Backend not starting

- Check Redis is running: `redis-cli ping`
- Check SQLite database permissions
- Verify .env configuration

### Frontend cannot connect

- Verify NEXT_PUBLIC_API_URL is correct
- Check CORS settings in backend
- Ensure backend is running

## 📞 Support

For issues and questions, please create an issue on GitHub.

## 📄 License

MIT License

## 👥 Contributors

- Your Name - Initial work

---

**Note**: This is a trading platform. Always test thoroughly before using with real funds. Use at your own risk.

# Start PostgreSQL

docker run -d -p 5432:5432 \
 -e POSTGRES_USER=tradercoin \
 -e POSTGRES_PASSWORD=tradercoin123 \
 -e POSTGRES_DB=tradercoin_db \
 postgres:15-alpine

# Run backend

cd Backend
DB_TYPE=postgresql go run main.go

📦 Files Mới
seed.go - Script tự động tạo admin và user mẫu
.env - File cấu hình Backend
.env.local - File cấu hình Frontend API URL
.env.local - File cấu hình Backoffice API URL
page.tsx - Auto-redirect login/dashboard
QUICKSTART.md - Hướng dẫn chạy nhanh
LOGIN_INFO.md - Thông tin đăng nhập chi tiết
🔐 Tài Khoản Mẫu
Khi chạy Backend lần đầu, hệ thống tự động tạo:

👤 USER (Frontend - Port 3000)
Email: user@example.com
Password: user123

🔐 ADMIN (Backoffice - Port 3001)
Email: admin@tradercoin.com
Password: admin123

🚀 Cách Chạy
Terminal 1 - Backend:
cd Backend
go run main.go

Terminal 2 - Frontend:
cd frontend
npm run dev

Terminal 3 - Backoffice (Optional):
cd backoffice
npm run dev

🎯 Truy Cập
Frontend User: http://localhost:3000

Login với: user@example.com / user123
Backoffice Admin: http://localhost:3001

Login với: admin@tradercoin.com / admin123
⚡ Tính Năng Auto-Redirect
Frontend (/) giờ sẽ tự động:

❌ Chưa login → Redirect về /login
✅ Đã login → Redirect về /dashboard

apikey
CfJsnKKOqXKzQBXca8Wii6rBW9sCSmSaK9Skn0JGG6ooAdaUSSMgMGbudTa6dnwz
Secret Key
bqQBmHfL0qKjUd8Vj7Y1GpLfA6RVMNq8eoLtHO0Fu6PLwNv4n2X19uzWaJsBbJH9

mwShmmfpqJcXZ3W1TWKWoiIuORmbpF1YCPz523SPTLIJEyyppMgxlWVg0Sy2YdYb
vKUcGXs3VkJlx7UwuUaLlPyWZhYkgE7hVIIpSMv8uoBSndsPb2LnbJMJh63XQa7F

webhook
curl -X POST http://localhost:8080/api/v1/signals/webhook/74c7c7f4ce33 \
 -H "Content-Type: application/json" \
 -d '{
"symbol": "ETHUSDT",
"action": "BUY",
"price": 2250.50,
"stopLoss": 2200.00,
"takeProfit": 2350.00,
"strategy": "Test WebSocket",
"message": "Testing real-time notification"
}'

note: chưa làm
priceProtect":false

Response Body: {"orderId":8389766058155948588,"symbol":"ETHUSDT","status":"FILLED","clientOrderId":"AL0F55jVYngvTWweEpAFb2","price":"0.00","avgPrice":"3057.89000","origQty":"0.007","executedQty":"0.007","cumQuote":"21.40523","timeInForce":"GTC","type":"MARKET","reduceOnly":false,"closePosition":false,"side":"BUY","positionSide":"BOTH","stopPrice":"0.00","workingType":"CONTRACT_PRICE","priceProtect":false,"origType":"MARKET","priceMatch":"NONE","selfTradePreventionMode":"EXPIRE_MAKER","goodTillDate":0,"time":1766412825838,"updateTime":1766412825838}

Response Body: {"orderId":8389766058138516683,"symbol":"ETHUSDT","status":"FILLED","clientOrderId":"fYgXbSXkFkN3j8VP3SG9IF","price":"0.00","avgPrice":"3052.27000","origQty":"0.007","executedQty":"0.007","cumQuote":"21.36589","timeInForce":"GTC","type":"MARKET","reduceOnly":false,"closePosition":false,"side":"BUY","positionSide":"BOTH","stopPrice":"0.00","workingType":"CONTRACT_PRICE","priceProtect":false,"origType":"MARKET","priceMatch":"NONE","selfTradePreventionMode":"EXPIRE_MAKER","goodTillDate":0,"time":1766411274076,"updateTime":1766411274076}
