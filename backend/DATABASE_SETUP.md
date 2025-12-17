# Hướng dẫn cấu hình Database

Backend TraderCoin sử dụng **GORM ORM** và hỗ trợ 2 loại database: **SQLite** và **PostgreSQL**

## GORM ORM

Backend đã được upgrade để sử dụng GORM ORM thay vì raw SQL:

- ✅ Type-safe database operations
- ✅ Automatic migrations
- ✅ Relationship management
- ✅ Soft deletes
- ✅ Query builder dễ sử dụng
- ✅ Hỗ trợ cả SQLite và PostgreSQL

Chi tiết migration: xem file `GORM_MIGRATION.md`

## Cấu hình SQLite (Mặc định)

SQLite phù hợp cho development và testing, không cần cài đặt database server.

### Bước 1: Cấu hình .env

```bash
DB_TYPE=sqlite
DB_PATH=./tradercoin.db
```

### Bước 2: Chạy backend

```bash
cd backend
go run main.go
```

Database file `tradercoin.db` sẽ được tạo tự động.

## Cấu hình PostgreSQL

PostgreSQL phù hợp cho production với hiệu năng cao và khả năng mở rộng tốt.

### Bước 1: Cài đặt PostgreSQL

**macOS (Homebrew):**

```bash
brew install postgresql@15
brew services start postgresql@15
```

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Docker:**

```bash
docker run -d \
  --name tradercoin-postgres \
  -e POSTGRES_USER=tradercoin \
  -e POSTGRES_PASSWORD=tradercoin123 \
  -e POSTGRES_DB=tradercoin_db \
  -p 5432:5432 \
  postgres:15-alpine
```

### Bước 2: Tạo Database và User

```bash
# Kết nối PostgreSQL
psql -U postgres

# Trong PostgreSQL shell
CREATE USER tradercoin WITH PASSWORD 'tradercoin123';
CREATE DATABASE tradercoin_db OWNER tradercoin;
GRANT ALL PRIVILEGES ON DATABASE tradercoin_db TO tradercoin;
\q
```

### Bước 3: Cấu hình .env

```bash
DB_TYPE=postgresql
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=tradercoin
POSTGRES_PASSWORD=tradercoin123
POSTGRES_DB=tradercoin_db
POSTGRES_SSLMODE=disable
```

### Bước 4: Chạy backend

```bash
cd backend
go run main.go
```

Backend sẽ tự động tạo tables và migrate schema.

## Chuyển đổi giữa SQLite và PostgreSQL

### SQLite → PostgreSQL

1. Export data từ SQLite (nếu cần)
2. Thay đổi `DB_TYPE=postgresql` trong `.env`
3. Cấu hình PostgreSQL connection
4. Restart backend - tables sẽ được tạo tự động

### PostgreSQL → SQLite

1. Export data từ PostgreSQL (nếu cần)
2. Thay đổi `DB_TYPE=sqlite` trong `.env`
3. Cấu hình `DB_PATH`
4. Restart backend - tables sẽ được tạo tự động

## Kiểm tra kết nối

Khi start backend, bạn sẽ thấy log:

**SQLite:**

```
📦 Using SQLite database
✅ Database connected successfully
✅ Database migrations completed
```

**PostgreSQL:**

```
🐘 Using PostgreSQL database
✅ Database connected successfully
✅ Database migrations completed
```

## Lưu ý

### SQLite

- ✅ Đơn giản, không cần setup
- ✅ Tốt cho development/testing
- ❌ Không phù hợp cho production với nhiều concurrent users
- ❌ Giới hạn về performance

### PostgreSQL

- ✅ Hiệu năng cao
- ✅ Hỗ trợ nhiều concurrent connections
- ✅ Phù hợp cho production
- ✅ Nhiều tính năng nâng cao (indexing, partitioning, replication)
- ❌ Cần cài đặt và cấu hình database server

## Troubleshooting

### Lỗi: "unsupported database type"

- Kiểm tra `DB_TYPE` trong `.env` phải là `sqlite` hoặc `postgresql`

### Lỗi kết nối PostgreSQL

```bash
# Kiểm tra PostgreSQL đang chạy
pg_isready -h localhost -p 5432

# Kiểm tra user và password
psql -U tradercoin -d tradercoin_db -h localhost
```

### Lỗi: "could not import github.com/lib/pq"

```bash
cd backend
go mod tidy
go get github.com/lib/pq
```

### Reset database

**SQLite:**

```bash
rm ./tradercoin.db
go run main.go  # Sẽ tạo lại database
```

**PostgreSQL:**

```sql
DROP DATABASE tradercoin_db;
CREATE DATABASE tradercoin_db OWNER tradercoin;
```

## Docker Compose cho PostgreSQL

Tạo file `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: tradercoin-postgres
    environment:
      POSTGRES_USER: tradercoin
      POSTGRES_PASSWORD: tradercoin123
      POSTGRES_DB: tradercoin_db
    ports:
      - '5432:5432'
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U tradercoin']
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

Chạy:

```bash
docker-compose up -d
```
