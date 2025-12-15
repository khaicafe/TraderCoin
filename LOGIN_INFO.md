# 🎯 Hướng Dẫn Đăng Nhập TraderCoin

## ✅ Đã Hoàn Tất

Hệ thống TraderCoin đã được cấu hình xong với:

- ✅ Backend API (Golang + SQLite)
- ✅ Frontend User Portal (Next.js)
- ✅ Backoffice Admin Portal (Next.js)
- ✅ Tài khoản mẫu đã được tạo sẵn
- ✅ Database tự động seed

---

## 🔐 Tài Khoản Đăng Nhập

### 👤 **USER PORTAL** (Frontend)

**URL:** http://localhost:3000

**Tài khoản:**

```
Email: user@example.com
Password: user123
```

**Chức năng:**

- Xem dashboard với thống kê trading
- Quản lý Exchange API Keys (Binance, Bittrex)
- Cấu hình stop-loss và take-profit
- Xem lịch sử giao dịch
- Quản lý profile

---

### 🔐 **ADMIN PORTAL** (Backoffice)

**URL:** http://localhost:3001

**Tài khoản:**

```
Email: admin@tradercoin.com
Password: admin123
```

**Chức năng:**

- Dashboard admin với thống kê tổng quan
- Quản lý users (suspend/activate)
- Quản lý subscriptions
- Xem transactions
- Analytics và reports

---

## 🚀 Cách Chạy

### 1. Start Backend (Terminal 1)

```bash
cd Backend
go run main.go
```

✅ Backend chạy tại: http://localhost:8080

### 2. Start Frontend (Terminal 2)

```bash
cd frontend
npm run dev
```

✅ Frontend chạy tại: http://localhost:3000

### 3. Start Backoffice (Terminal 3 - Optional)

```bash
cd backoffice
npm run dev
```

✅ Backoffice chạy tại: http://localhost:3001

---

## 📋 Quy Trình Sử Dụng

### User (Trader)

1. **Đăng nhập** tại http://localhost:3000

   - Email: `user@example.com`
   - Password: `user123`

2. **Thêm Exchange API Key**

   - Vào "Exchange Keys"
   - Nhấn "+ Add New Key"
   - Chọn Binance hoặc Bittrex
   - Nhập API Key và Secret

3. **Cấu hình Trading**

   - Vào "Trading Config"
   - Tạo config mới với:
     - Symbol (ví dụ: BTCUSDT)
     - Stop Loss % (ví dụ: -5%)
     - Take Profit % (ví dụ: +10%)

4. **Theo dõi Orders**
   - Vào "Orders" để xem lịch sử giao dịch
   - Hệ thống tự động thực hiện stop-loss/take-profit

---

### Admin (Quản trị)

1. **Đăng nhập Admin** tại http://localhost:3001

   - Email: `admin@tradercoin.com`
   - Password: `admin123`

2. **Quản lý Users**

   - Xem danh sách users
   - Suspend/Activate tài khoản
   - Xem thông tin chi tiết

3. **Quản lý Subscriptions**

   - Xem các gói đăng ký
   - Gia hạn hoặc hủy subscription
   - Xem doanh thu

4. **Xem Transactions**
   - Theo dõi các giao dịch
   - Export reports
   - Analytics

---

## 🔧 Troubleshooting

### Frontend không redirect đúng?

- Xóa localStorage: Mở DevTools (F12) → Application → Local Storage → Clear All
- Refresh trang

### Backend không chạy?

```bash
# Kiểm tra port 8080 có bị chiếm không
lsof -i :8080

# Kill process nếu cần
kill -9 <PID>
```

### Database bị lỗi?

```bash
# Xóa và tạo lại database
cd Backend
rm tradercoin.db
go run main.go
```

---

## 📱 Screenshots

### User Dashboard

![Dashboard](docs/screenshots/user-dashboard.png)

- Tổng quan portfolio
- Trading stats
- Quick actions

### Exchange Keys

![Keys](docs/screenshots/exchange-keys.png)

- Quản lý API keys
- Encrypted storage
- Multiple exchanges

### Admin Portal

![Admin](docs/screenshots/admin-dashboard.png)

- User management
- Revenue stats
- System analytics

---

## 🎉 Sẵn Sàng Trading!

Hệ thống của bạn đã sẵn sàng để:

- ✅ Kết nối với Binance/Bittrex
- ✅ Tự động stop-loss/take-profit
- ✅ Theo dõi portfolio realtime
- ✅ Quản lý multi-user

**Happy Trading! 📈🚀**

---

## 📞 Hỗ Trợ

Nếu có vấn đề, kiểm tra:

1. Backend logs (Terminal 1)
2. Frontend console (DevTools F12)
3. File `.env` và `.env.local`
4. Port conflicts (8080, 3000, 3001)

**Email:** support@tradercoin.com
