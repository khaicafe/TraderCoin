# 🚀 TraderCoin - Quick Start Guide

## 📋 Tài Khoản Mẫu

Sau khi chạy Backend lần đầu, hệ thống sẽ tự động tạo 2 tài khoản mẫu:

### 👤 User Account (Frontend)

- **Email:** user@example.com
- **Password:** user123
- **URL:** http://localhost:3000

### 🔐 Admin Account (Backoffice)

- **Email:** admin@tradercoin.com
- **Password:** admin123
- **URL:** http://localhost:3001

---

## 🛠️ Cài Đặt & Chạy Project

### 1️⃣ Backend (Golang)

```bash
cd Backend

# Cài đặt dependencies
go mod download

# Chạy server (sẽ tự động tạo DB và seed data)
go run main.go
```

✅ Backend chạy tại: **http://localhost:8080**

---

### 2️⃣ Frontend (Next.js - User Portal)

```bash
cd frontend

# Cài đặt dependencies
npm install

# Chạy development server
npm run dev
```

✅ Frontend chạy tại: **http://localhost:3000**

**Đăng nhập với:** user@example.com / user123

---

### 3️⃣ Backoffice (Next.js - Admin Portal)

```bash
cd backoffice

# Cài đặt dependencies
npm install

# Chạy development server
npm run dev
```

✅ Backoffice chạy tại: **http://localhost:3001**

**Đăng nhập với:** admin@tradercoin.com / admin123

---

## 📦 Yêu Cầu Hệ Thống

- **Go:** 1.21 trở lên
- **Node.js:** 20.9.0 trở lên
- **Redis:** (Optional - có thể chạy mà không cần Redis)

---

## 🔧 Cấu Hình

### Backend `.env`

File đã được tạo sẵn tại `/Backend/.env`:

```env
DB_PATH=./tradercoin.db
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
PORT=8080
```

### Frontend `.env.local`

File đã được tạo sẵn tại `/frontend/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Backoffice `.env.local`

File đã được tạo sẵn tại `/backoffice/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 🎯 Tính Năng Đã Triển Khai

### Frontend (User)

- ✅ Login/Register
- ✅ Dashboard với thống kê
- ✅ Exchange API Keys Management (Binance, Bittrex)
- 🔄 Trading Config Setup (Coming soon)
- 🔄 Order History (Coming soon)

### Backoffice (Admin)

- ✅ Admin Login
- ✅ Dashboard với stats
- ✅ User Management (suspend/activate users)
- 🔄 Subscriptions Management (Coming soon)
- 🔄 Transactions History (Coming soon)

### Backend (API)

- ✅ Database setup (SQLite)
- ✅ Auto migrations
- ✅ Sample data seeding
- 🔄 JWT Authentication (In progress)
- 🔄 Exchange API Integration (In progress)
- 🔄 Trading Engine (Coming soon)

---

## 🐛 Troubleshooting

### Backend không chạy?

```bash
# Kiểm tra Go version
go version

# Xóa DB cũ và chạy lại
rm tradercoin.db
go run main.go
```

### Frontend/Backoffice lỗi?

```bash
# Xóa node_modules và cài lại
rm -rf node_modules package-lock.json
npm install

# Kiểm tra Node version
node -v  # Should be >= 20.9.0
```

### Không kết nối được API?

- Kiểm tra Backend có chạy tại http://localhost:8080
- Kiểm tra file `.env.local` trong frontend/backoffice
- Kiểm tra CORS settings

---

## 📝 Development Notes

- Database file: `Backend/tradercoin.db`
- Sample accounts tự động tạo khi chạy Backend lần đầu
- Frontend tự động redirect: `/` → `/login` hoặc `/dashboard`
- Admin portal độc lập tại port 3001

---

## 🎉 Bắt Đầu

```bash
# Terminal 1 - Backend
cd Backend && go run main.go

# Terminal 2 - Frontend
cd frontend && npm run dev

# Terminal 3 - Backoffice (optional)
cd backoffice && npm run dev
```

Sau đó truy cập:

- User Portal: http://localhost:3000
- Admin Portal: http://localhost:3001

---

**Happy Trading! 🚀📈**
