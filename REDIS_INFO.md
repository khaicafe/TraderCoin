# 🔧 Redis Configuration - TraderCoin

## ❓ Redis dùng để làm gì?

Redis được sử dụng trong TraderCoin cho các mục đích:

### 1. **Session Management** (Quản lý phiên đăng nhập)

- Lưu trữ JWT tokens
- Quản lý refresh tokens
- Theo dõi sessions đang hoạt động
- Auto-expire sessions

### 2. **Rate Limiting** (Giới hạn tốc độ)

- Ngăn chặn spam requests
- Bảo vệ API khỏi DDoS
- Giới hạn số lần đăng nhập thất bại
- Throttle API calls

### 3. **Caching** (Bộ nhớ đệm)

- Cache thông tin user profile
- Cache exchange rates/prices
- Cache trading configs
- Giảm tải database

### 4. **Real-time Data** (Dữ liệu thời gian thực)

- WebSocket connections tracking
- Real-time price updates
- Order status notifications
- Live portfolio updates

---

## ⚙️ Trạng Thái Hiện Tại

✅ **Redis là OPTIONAL** - Hệ thống vẫn chạy bình thường nếu không có Redis

Khi bạn chạy Backend, sẽ thấy thông báo:

```
⚠️  Warning: Redis not available: dial tcp [::1]:6379: connect: connection refused
ℹ️  System will run without Redis caching
```

**Điều này là BÌNH THƯỜNG!** Backend vẫn hoạt động đầy đủ chức năng.

---

## 🚀 Cài Đặt Redis (Optional)

### macOS

```bash
# Cài đặt qua Homebrew
brew install redis

# Khởi động Redis
brew services start redis

# Hoặc chạy tạm thời
redis-server
```

### Linux (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

### Docker

```bash
docker run -d -p 6379:6379 --name redis redis:alpine
```

---

## ✅ Kiểm Tra Redis

```bash
# Kiểm tra Redis có chạy không
redis-cli ping
# Kết quả mong đợi: PONG

# Kiểm tra port
lsof -i :6379
```

---

## 🔄 Restart Backend Sau Khi Cài Redis

```bash
cd Backend
go run .
```

Bạn sẽ thấy:

```
✅ Redis connected successfully
```

---

## 📊 Lợi Ích Khi Có Redis

| Tính năng         | Không có Redis | Có Redis                |
| ----------------- | -------------- | ----------------------- |
| **Tốc độ API**    | Bình thường    | Nhanh hơn 10-100x       |
| **Session**       | JWT only       | JWT + Redis cache       |
| **Rate Limiting** | Basic          | Advanced với tracking   |
| **Real-time**     | Polling        | WebSocket + Pub/Sub     |
| **Caching**       | Không có       | Profile, configs cached |

---

## ⚠️ Khi Nào Cần Redis?

### ✅ CẦN Redis khi:

- Production environment
- Nhiều users đồng thời (>100)
- Cần real-time updates
- WebSocket connections
- High-performance caching

### ❌ KHÔNG cần Redis khi:

- Development/Testing
- Ít users (<10)
- Demo/Prototype
- Local development
- **Đang học và thử nghiệm** ← BẠN Ở ĐÂY!

---

## 🎯 Kết Luận

**Hiện tại:** Bạn không cần Redis! Backend đã sửa để chạy tốt mà không cần Redis.

**Sau này:** Khi deploy lên production hoặc cần performance cao, hãy cài Redis.

**Lưu ý quan trọng:**

- ✅ Tất cả API endpoints đã được sửa thành `/api/v1/*`
- ✅ Backend chạy tốt với hoặc không có Redis
- ✅ Login/Register hoạt động bình thường
- ✅ Database SQLite đã có sẵn accounts

---

## 📞 API Endpoints Đã Sửa

### Frontend

- ✅ `/api/auth/login` → `/api/v1/auth/login`
- ✅ `/api/auth/register` → `/api/v1/auth/register`
- ✅ `/api/user/profile` → `/api/v1/user/profile`
- ✅ `/api/exchange-keys` → `/api/v1/keys`
- ✅ `/api/trading-configs` → `/api/v1/trading/configs`

### Backoffice

- ✅ `/api/admin/login` → `/api/v1/admin/login`
- ✅ `/api/admin/users` → `/api/v1/admin/users`
- ✅ `/api/admin/dashboard/stats` → `/api/v1/admin/statistics`

---

**Giờ bạn có thể đăng nhập thành công! 🎉**
