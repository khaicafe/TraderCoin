# Migration từ Raw SQL sang GORM

## Tổng quan

Backend đã được cập nhật để sử dụng GORM ORM thay vì raw SQL queries. GORM cung cấp:

- ✅ Type-safe database operations
- ✅ Automatic migrations
- ✅ Hỗ trợ cả SQLite và PostgreSQL
- ✅ Relationship management
- ✅ Soft deletes
- ✅ Query builder dễ sử dụng

## Các thay đổi đã thực hiện

### 1. Models (`backend/models/models.go`)

Models đã được cập nhật với GORM tags:

```go
type User struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    Email           string         `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash    string         `gorm:"not null" json:"-"`
    // ... các fields khác
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete

    // Relationships
    ExchangeKeys   []ExchangeKey   `gorm:"foreignKey:UserID"`
    Orders         []Order         `gorm:"foreignKey:UserID"`
}
```

**Thay đổi quan trọng:**

- `int` → `uint` cho IDs
- Thêm `DeletedAt` cho soft delete
- Thêm relationship declarations
- GORM tags để define constraints

### 2. Database Connection (`backend/database/database.go`)

```go
// Trước (Raw SQL)
func Connect() (*sql.DB, error)

// Sau (GORM)
func Connect() (*gorm.DB, error)
```

### 3. Migrations

GORM tự động tạo và cập nhật schema:

```go
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.ExchangeKey{},
        // ... other models
    )
}
```

### 4. Services (`backend/services/services.go`)

```go
type Services struct {
    DB    *gorm.DB      // Thay đổi từ *sql.DB
    Redis *redis.Client
}
```

## Cập nhật Controllers

### File cần sửa: `backend/controllers/trading.go`

File backup đã được tạo: `trading.go.backup`

### Pattern chuyển đổi phổ biến

#### 1. Query một record

**Trước (Raw SQL):**

```go
var user models.User
err := db.QueryRow("SELECT * FROM users WHERE email = ?", email).
    Scan(&user.ID, &user.Email, &user.PasswordHash, ...)
```

**Sau (GORM):**

```go
var user models.User
err := db.Where("email = ?", email).First(&user).Error
// Hoặc ngắn gọn hơn:
err := db.First(&user, "email = ?", email).Error
```

#### 2. Query nhiều records

**Trước (Raw SQL):**

```go
rows, err := db.Query("SELECT * FROM users WHERE status = ?", "active")
defer rows.Close()

var users []models.User
for rows.Next() {
    var user models.User
    rows.Scan(&user.ID, &user.Email, ...)
    users = append(users, user)
}
```

**Sau (GORM):**

```go
var users []models.User
err := db.Where("status = ?", "active").Find(&users).Error
```

#### 3. Insert

**Trước (Raw SQL):**

```go
result, err := db.Exec(`
    INSERT INTO users (email, password_hash, full_name)
    VALUES (?, ?, ?)
`, email, passwordHash, fullName)
userID, _ := result.LastInsertId()
```

**Sau (GORM):**

```go
user := models.User{
    Email:        email,
    PasswordHash: passwordHash,
    FullName:     fullName,
}
err := db.Create(&user).Error
// user.ID sẽ tự động được set
```

#### 4. Update

**Trước (Raw SQL):**

```go
_, err := db.Exec(`
    UPDATE users SET full_name = ?, phone = ? WHERE id = ?
`, fullName, phone, userID)
```

**Sau (GORM):**

```go
// Cách 1: Update specific fields
err := db.Model(&models.User{}).
    Where("id = ?", userID).
    Updates(map[string]interface{}{
        "full_name": fullName,
        "phone":     phone,
    }).Error

// Cách 2: Update struct
user := models.User{ID: userID}
db.First(&user)
user.FullName = fullName
user.Phone = phone
err := db.Save(&user).Error
```

#### 5. Delete

**Trước (Raw SQL):**

```go
_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
```

**Sau (GORM):**

```go
// Soft delete (recommended)
err := db.Delete(&models.User{}, userID).Error

// Hard delete
err := db.Unscoped().Delete(&models.User{}, userID).Error
```

#### 6. Count

**Trước (Raw SQL):**

```go
var count int
err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
```

**Sau (GORM):**

```go
var count int64
err := db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
```

#### 7. Exists check

**Trước (Raw SQL):**

```go
var exists int
err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
if exists > 0 {
    // User exists
}
```

**Sau (GORM):**

```go
var count int64
db.Model(&models.User{}).Where("email = ?", email).Count(&count)
if count > 0 {
    // User exists
}
```

## Ví dụ cụ thể cho Controllers

### Register Function

```go
func Register(services *services.Services) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Email    string `json:"email" binding:"required,email"`
            Password string `json:"password" binding:"required,min=6"`
            FullName string `json:"full_name" binding:"required"`
            Phone    string `json:"phone"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Check if user exists
        var count int64
        services.DB.Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
        if count > 0 {
            c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
            return
        }

        // Hash password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
            return
        }

        // Create user
        subscriptionEnd := time.Now().AddDate(0, 0, 30)
        user := models.User{
            Email:           input.Email,
            PasswordHash:    string(hashedPassword),
            FullName:        input.FullName,
            Phone:           input.Phone,
            Status:          "active",
            SubscriptionEnd: &subscriptionEnd,
        }

        if err := services.DB.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
            return
        }

        // Generate JWT token
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "user_id": user.ID,
            "email":   user.Email,
            "exp":     time.Now().Add(24 * time.Hour).Unix(),
        })

        tokenString, err := token.SignedString(jwtSecret)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{
            "message": "User registered successfully",
            "token":   tokenString,
            "user": gin.H{
                "id":        user.ID,
                "email":     user.Email,
                "full_name": user.FullName,
            },
        })
    }
}
```

### Login Function

```go
func Login(services *services.Services) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Email    string `json:"email" binding:"required,email"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
            return
        }

        // Find user
        var user models.User
        if err := services.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        // Check password
        if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        // Check user status
        if user.Status != "active" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
            return
        }

        // Generate JWT token
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "user_id": user.ID,
            "email":   user.Email,
            "exp":     time.Now().Add(24 * time.Hour).Unix(),
        })

        tokenString, err := token.SignedString(jwtSecret)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Login successful",
            "token":   tokenString,
            "user": gin.H{
                "id":        user.ID,
                "email":     user.Email,
                "full_name": user.FullName,
            },
        })
    }
}
```

## GORM Query Tips

### Pagination

```go
var users []models.User
page := 1
pageSize := 10
offset := (page - 1) * pageSize

db.Limit(pageSize).Offset(offset).Find(&users)
```

### Order By

```go
db.Order("created_at desc").Find(&users)
```

### Preload Relationships

```go
// Load user with exchange keys
var user models.User
db.Preload("ExchangeKeys").First(&user, userID)
```

### Complex Queries

```go
db.Where("status = ?", "active").
   Where("subscription_end > ?", time.Now()).
   Order("created_at desc").
   Limit(10).
   Find(&users)
```

### Transaction

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // Create user
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    // Create related data
    if err := tx.Create(&exchangeKey).Error; err != nil {
        return err
    }

    return nil
})
```

## Testing

### SQLite (Default)

```bash
cd backend
DB_TYPE=sqlite go run main.go
```

### PostgreSQL

```bash
# Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=tradercoin \
  -e POSTGRES_PASSWORD=tradercoin123 \
  -e POSTGRES_DB=tradercoin_db \
  postgres:15-alpine

# Run backend
DB_TYPE=postgresql go run main.go
```

## Troubleshooting

### Type mismatch errors

- Đổi `int` thành `uint` cho IDs
- Đổi `time.Time` thành `*time.Time` cho nullable dates

### Foreign key errors

- Ensure relationships are properly defined in models
- Check CASCADE constraints

### Migration errors

```go
// Drop all tables and recreate
db.Migrator().DropTable(&models.User{}, &models.ExchangeKey{}, ...)
db.AutoMigrate(&models.User{}, &models.ExchangeKey{}, ...)
```

## Next Steps

1. ✅ Update all functions in `trading.go` to use GORM
2. Test each endpoint thoroughly
3. Update any other controllers if they exist
4. Add proper error handling
5. Add logging
6. Write unit tests

## Resources

- [GORM Documentation](https://gorm.io/docs/)
- [GORM Gen (Code generator)](https://gorm.io/gen/)
- [Best Practices](https://gorm.io/docs/conventions.html)
