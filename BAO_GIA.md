# BÁO GIÁ DỰ ÁN TRADERCOIN

## Thông Tin Dự Án
**Tên dự án:** TraderCoin - Nền Tảng Giao Dịch Cryptocurrency Tự Động  
**Ngày báo giá:** 16/12/2025  
**Thời gian thực hiện:** 8-12 tuần  
**Bảo hành:** 3 tháng miễn phí

---

## Tổng Quan Hệ Thống

TraderCoin là một nền tảng giao dịch cryptocurrency tự động hoàn chỉnh bao gồm:
- **Frontend:** Ứng dụng người dùng (Next.js + TypeScript + Tailwind CSS)
- **Backend:** API và Trading Engine (Golang + PostgreSQL + Redis)
- **Backoffice:** Hệ thống quản trị (Next.js + TypeScript + Tailwind CSS)

---

## CHI TIẾT TÍNH NĂNG VÀ GIÁ

### 1. FRONTEND - ỨNG DỤNG NGƯỜI DÙNG

#### 1.1 Authentication & User Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Đăng ký tài khoản | Form đăng ký với validation | 2,000,000 |
| Đăng nhập | JWT authentication | 1,500,000 |
| Quên mật khẩu | Reset password qua email | 1,500,000 |
| Quản lý profile | Cập nhật thông tin cá nhân | 1,000,000 |
| **Tổng phụ** | | **6,000,000** |

#### 1.2 Dashboard
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Tổng quan thống kê | 4 cards thống kê chính | 2,500,000 |
| Biểu đồ doanh thu | Charts với real-time data | 3,000,000 |
| Lịch sử giao dịch | Timeline activities | 2,000,000 |
| **Tổng phụ** | | **7,500,000** |

#### 1.3 Exchange Keys Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Kết nối Binance API | Integration với Binance | 3,000,000 |
| Kết nối Bittrex API | Integration với Bittrex | 3,000,000 |
| Quản lý API keys | CRUD operations | 2,000,000 |
| Test connection | Kiểm tra kết nối API | 1,500,000 |
| **Tổng phụ** | | **9,500,000** |

#### 1.4 Bot Configuration
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Tạo bot config | Form với 9 fields validation | 3,000,000 |
| Danh sách bot configs | Table với filter & search | 2,500,000 |
| Chỉnh sửa/Xóa config | CRUD operations | 2,000,000 |
| Import/Export configs | JSON format | 1,500,000 |
| **Tổng phụ** | | **9,000,000** |

#### 1.5 Trading (Đặt Lệnh)
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Đặt lệnh Market | Giao dịch giá thị trường | 3,000,000 |
| Đặt lệnh Limit | Giao dịch giá cố định | 3,000,000 |
| Stop Loss / Take Profit | Tự động cắt lỗ/chốt lời | 4,000,000 |
| Symbol search | Tìm kiếm 40+ symbols | 2,000,000 |
| Warning alerts | Cảnh báo rủi ro | 1,500,000 |
| **Tổng phụ** | | **13,500,000** |

#### 1.6 Orders Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Danh sách lệnh | Table với pagination | 2,500,000 |
| Lọc & tìm kiếm | Filter by status, time | 2,000,000 |
| Hủy lệnh | Cancel pending orders | 1,500,000 |
| Chi tiết lệnh | Order details modal | 1,500,000 |
| **Tổng phụ** | | **7,500,000** |

#### 1.7 Monitoring
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Real-time bot status | WebSocket updates | 4,000,000 |
| Vị thế đang mở | Active positions tracking | 3,000,000 |
| Performance charts | Biểu đồ hiệu suất | 3,000,000 |
| Recent activity logs | Timeline activities | 2,000,000 |
| **Tổng phụ** | | **12,000,000** |

#### 1.8 Logs & Error Tracking
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| System logs | Log viewer với colors | 2,500,000 |
| Lọc theo loại log | Success/Error/Warning/Info | 1,500,000 |
| Tìm kiếm logs | Search functionality | 1,500,000 |
| Export logs | CSV/JSON export | 1,500,000 |
| **Tổng phụ** | | **7,000,000** |

#### 1.9 Portfolio
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Tổng quan tài sản | Holdings overview | 2,500,000 |
| Lịch sử giao dịch | Transaction history | 2,000,000 |
| P&L tracking | Profit/Loss calculation | 3,000,000 |
| **Tổng phụ** | | **7,500,000** |

#### 1.10 Settings
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Profile settings | Cập nhật thông tin | 1,500,000 |
| Trading settings | Default SL/TP, leverage | 2,000,000 |
| Notifications | Email/Push notifications | 2,000,000 |
| **Tổng phụ** | | **5,500,000** |

**TỔNG FRONTEND: 85,000,000 VNĐ**

---

### 2. BACKEND - API & TRADING ENGINE

#### 2.1 Core API
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| RESTful API | Golang + Gin framework | 10,000,000 |
| JWT Authentication | Secure authentication | 3,000,000 |
| Database design | PostgreSQL schema | 5,000,000 |
| Redis caching | Performance optimization | 3,000,000 |
| API documentation | Swagger/OpenAPI | 2,000,000 |
| **Tổng phụ** | | **23,000,000** |

#### 2.2 Exchange Integration
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Binance API integration | Trading, market data | 8,000,000 |
| Bittrex API integration | Trading, market data | 8,000,000 |
| WebSocket real-time | Price updates | 5,000,000 |
| Error handling | Retry logic, fallback | 3,000,000 |
| **Tổng phụ** | | **24,000,000** |

#### 2.3 Trading Engine
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Automated trading logic | Bot execution engine | 15,000,000 |
| Stop Loss execution | Tự động cắt lỗ | 5,000,000 |
| Take Profit execution | Tự động chốt lời | 5,000,000 |
| Order monitoring | Theo dõi lệnh real-time | 6,000,000 |
| Risk management | Position sizing, limits | 5,000,000 |
| **Tổng phụ** | | **36,000,000** |

#### 2.4 Data Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| User management | CRUD operations | 3,000,000 |
| Bot configs | CRUD + validation | 3,000,000 |
| Orders history | Storage & retrieval | 3,000,000 |
| Transactions tracking | P&L calculation | 4,000,000 |
| Logs storage | System logs database | 2,000,000 |
| **Tổng phụ** | | **15,000,000** |

#### 2.5 WebSocket Server
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Real-time updates | Price, orders, positions | 6,000,000 |
| Connection management | Handle 1000+ clients | 4,000,000 |
| **Tổng phụ** | | **10,000,000** |

**TỔNG BACKEND: 108,000,000 VNĐ**

---

### 3. BACKOFFICE - HỆ THỐNG QUẢN TRỊ

#### 3.1 Admin Authentication
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Admin login | Secure admin access | 2,000,000 |
| Role-based access | Admin/Super Admin roles | 3,000,000 |
| **Tổng phụ** | | **5,000,000** |

#### 3.2 User Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Danh sách users | Table với pagination | 3,000,000 |
| Suspend/Activate users | User control | 2,000,000 |
| View user details | Full information | 2,000,000 |
| User statistics | Charts & reports | 3,000,000 |
| **Tổng phụ** | | **10,000,000** |

#### 3.3 Subscription Management
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Gói subscription | Plan management | 3,000,000 |
| Billing history | Payment tracking | 3,000,000 |
| Revenue reports | Analytics | 4,000,000 |
| **Tổng phụ** | | **10,000,000** |

#### 3.4 System Monitoring
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| System health | Server status | 3,000,000 |
| Trading statistics | Overall metrics | 3,000,000 |
| Error monitoring | System errors tracking | 3,000,000 |
| **Tổng phụ** | | **9,000,000** |

#### 3.5 Reports & Analytics
| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Trading reports | Daily/Weekly/Monthly | 4,000,000 |
| User activity | Usage analytics | 3,000,000 |
| Export reports | PDF/Excel export | 2,000,000 |
| **Tổng phụ** | | **9,000,000** |

**TỔNG BACKOFFICE: 43,000,000 VNĐ**

---

### 4. DEPLOYMENT & INFRASTRUCTURE

| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Docker containers | Containerization | 3,000,000 |
| Docker Compose setup | Development environment | 2,000,000 |
| Database migration | Schema versioning | 2,000,000 |
| Environment config | .env setup cho prod/dev | 1,000,000 |
| CI/CD setup | GitHub Actions | 3,000,000 |
| Server deployment | AWS/GCP/Azure setup | 4,000,000 |
| SSL certificate | HTTPS setup | 1,000,000 |
| Domain configuration | DNS setup | 500,000 |
| **Tổng** | | **16,500,000** |

---

### 5. TESTING & DOCUMENTATION

| Tính năng | Mô tả | Giá (VNĐ) |
|-----------|-------|-----------|
| Unit testing | Backend tests | 5,000,000 |
| Integration testing | API tests | 4,000,000 |
| E2E testing | Frontend tests | 4,000,000 |
| User manual | Hướng dẫn sử dụng | 3,000,000 |
| Technical docs | Developer documentation | 3,000,000 |
| API documentation | Swagger/Postman | 2,000,000 |
| **Tổng** | | **21,000,000** |

---

### 6. BẢO HÀNH & HỖ TRỢ

| Dịch vụ | Mô tả | Giá (VNĐ) |
|---------|-------|-----------|
| Bảo hành 3 tháng | Bug fixes miễn phí | Miễn phí |
| Hỗ trợ kỹ thuật | Email/Chat support | Miễn phí |
| Training | Đào tạo sử dụng hệ thống | 5,000,000 |
| Handover | Bàn giao source code & docs | 3,000,000 |
| **Tổng** | | **8,000,000** |

---

## TỔNG KẾT GIÁ

| Hạng mục | Giá (VNĐ) | USD (est.) |
|----------|-----------|------------|
| Frontend | 85,000,000 | $3,400 |
| Backend | 108,000,000 | $4,320 |
| Backoffice | 43,000,000 | $1,720 |
| Deployment & Infrastructure | 16,500,000 | $660 |
| Testing & Documentation | 21,000,000 | $840 |
| Bảo hành & Hỗ trợ | 8,000,000 | $320 |
| **TỔNG CỘNG** | **281,500,000** | **$11,260** |

### CHIẾT KHẤU (nếu thanh toán full)
- Giảm 10%: **253,350,000 VNĐ** (~$10,134 USD)

---

## PHƯƠNG THỨC THANH TOÁN

### Gói 1: Thanh toán theo tiến độ
- **Đợt 1 (30%):** Ký hợp đồng - 84,450,000 VNĐ
- **Đợt 2 (30%):** Hoàn thành UI/UX + Backend API - 84,450,000 VNĐ
- **Đợt 3 (30%):** Hoàn thành Trading Engine - 84,450,000 VNĐ
- **Đợt 4 (10%):** Deploy & Bàn giao - 28,150,000 VNĐ

### Gói 2: Thanh toán full (có chiết khấu)
- **1 lần:** 253,350,000 VNĐ (giảm 10%)

---

## TIMELINE THỰC HIỆN

| Giai đoạn | Thời gian | Công việc |
|-----------|-----------|-----------|
| **Tuần 1-2** | 2 tuần | Setup project, Database design, UI/UX wireframe |
| **Tuần 3-5** | 3 tuần | Frontend development (Dashboard, Trading, Orders) |
| **Tuần 6-8** | 3 tuần | Backend API & Exchange integration |
| **Tuần 9-10** | 2 tuần | Trading Engine & WebSocket |
| **Tuần 11** | 1 tuần | Backoffice development |
| **Tuần 12** | 1 tuần | Testing, Bug fixes, Deployment |

**Tổng: 12 tuần (3 tháng)**

---

## CÔNG NGHỆ SỬ DỤNG

### Frontend
- **Framework:** Next.js 14
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Icons:** Heroicons
- **State Management:** React Hooks

### Backend
- **Language:** Golang
- **Framework:** Gin
- **Database:** PostgreSQL
- **Cache:** Redis
- **Authentication:** JWT

### DevOps
- **Containerization:** Docker
- **Orchestration:** Docker Compose
- **CI/CD:** GitHub Actions
- **Cloud:** AWS/GCP/Azure

---

## YÊU CẦU HỆ THỐNG

### Server Requirements (Minimum)
- **CPU:** 4 cores
- **RAM:** 8GB
- **Storage:** 100GB SSD
- **Bandwidth:** 100Mbps

### Server Requirements (Recommended)
- **CPU:** 8 cores
- **RAM:** 16GB
- **Storage:** 200GB SSD
- **Bandwidth:** 1Gbps

---

## ĐIỀU KHOẢN & ĐIỀU KIỆN

1. **Source Code:** Bàn giao full source code sau khi thanh toán 100%
2. **Bảo hành:** 3 tháng kể từ ngày bàn giao
3. **Hỗ trợ:** Email/Chat support trong giờ hành chính
4. **Thay đổi:** Mọi thay đổi phạm vi sẽ được báo giá riêng
5. **Hủy dự án:** Khách hàng phải thanh toán phần công việc đã hoàn thành

---

## BẢO MẬT & BẢN QUYỀN

- Source code thuộc quyền sở hữu của khách hàng sau khi thanh toán full
- Ký NDA (Non-Disclosure Agreement) trước khi bắt đầu
- Mọi dữ liệu khách hàng được bảo mật tuyệt đối
- API keys và secrets không được lưu trong source code

---

## LIÊN HỆ

**Email:** contact@tradercoin.com  
**Phone:** +84 xxx xxx xxx  
**Website:** https://tradercoin.com

---

*Báo giá có hiệu lực trong 30 ngày kể từ ngày phát hành*  
*Giá chưa bao gồm VAT (10%)*  
*Chi phí server/hosting tính riêng theo thực tế*
