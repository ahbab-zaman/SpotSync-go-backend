# Library Docs

Project-specific usage patterns for every third-party library in SpotSync. Read the relevant section before implementing any feature that touches these libraries.

---

## Echo — Web Framework

Import: `github.com/labstack/echo/v4`

### Setup in main.go

```go
e := echo.New()
e.Validator = &customValidator{validator: validator.New()}
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
    AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
}))
```

### Binding and Validation

```go
func (h *AuthHandler) Register(c echo.Context) error {
    var req dto.RegisterRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "success": false,
            "message": "Invalid request body",
            "errors":  err.Error(),
        })
    }
    if err := c.Validate(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "success": false,
            "message": "Validation failed",
            "errors":  err.Error(),
        })
    }
    // proceed
}
```

### Reading JWT Claims from Context

```go
// After jwt_middleware has run, read claims like this:
userID := c.Get("userID").(uint)
role := c.Get("role").(string)
```

### Route Groups

```go
api := e.Group("/api/v1")

// Public
api.GET("/zones", zoneHandler.GetAll)

// Authenticated group
auth := api.Group("", middleware.JWTMiddleware())
auth.POST("/reservations", reservationHandler.Create)

// Admin group
admin := api.Group("", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))
admin.POST("/zones", zoneHandler.Create)
```

### JSON Response

```go
// Always use the standard shape
return c.JSON(http.StatusCreated, map[string]interface{}{
    "success": true,
    "message": "User registered successfully",
    "data":    result,
})
```

**Rules:**

- Always bind then validate — never skip either step
- Never return raw Go errors to the client — always wrap in the standard response shape
- Always use `c.Get("userID")` and `c.Get("role")` in handlers after JWT middleware — never re-parse the token
- Use route groups to apply middleware cleanly — never apply middleware per-handler manually

---

## GORM — ORM

Import: `gorm.io/gorm` and `gorm.io/driver/postgres`

### Database Connection

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "os"
)

func connectDB() *gorm.DB {
    dsn := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Connection pool — important for production
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    return db
}
```

### AutoMigrate

```go
db.AutoMigrate(
    &models.User{},
    &models.ParkingZone{},
    &models.Reservation{},
)
```

Call once on startup. GORM creates tables and adds missing columns but never deletes columns.

### Basic Queries

```go
// Find one by ID
var zone models.ParkingZone
if err := db.First(&zone, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrNotFound
    }
    return nil, err
}

// Find all
var zones []models.ParkingZone
if err := db.Find(&zones).Error; err != nil {
    return nil, err
}

// Find with condition
var user models.User
db.Where("email = ?", email).First(&user)

// Create
if err := db.Create(&user).Error; err != nil {
    return err
}

// Update specific fields
db.Model(&reservation).Update("status", "cancelled")

// Delete
db.Delete(&models.ParkingZone{}, id)

// Count
var count int64
db.Model(&models.Reservation{}).
    Where("zone_id = ? AND status = ?", zoneID, "active").
    Count(&count)
```

### Preloading Associations

```go
// Preload Zone on Reservation (for my-reservations endpoint)
var reservations []models.Reservation
db.Preload("Zone").Where("user_id = ?", userID).Find(&reservations)

// Preload both User and Zone (for admin all-reservations endpoint)
db.Preload("User").Preload("Zone").Find(&reservations)
```

### Transaction with Row-Level Lock (FOR UPDATE)

This is the most critical GORM pattern in the project. Used exclusively in `CreateWithLock`.

```go
import "gorm.io/gorm/clause"

func (r *reservationRepository) CreateWithLock(reservation *models.Reservation, zoneID uint) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Step 1: Lock the zone row — no other transaction can read or modify it until this one commits
        var zone models.ParkingZone
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&zone, zoneID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return ErrNotFound
            }
            return err
        }

        // Step 2: Count active reservations for this zone
        var activeCount int64
        if err := tx.Model(&models.Reservation{}).
            Where("zone_id = ? AND status = ?", zoneID, "active").
            Count(&activeCount).Error; err != nil {
            return err
        }

        // Step 3: Enforce capacity — reject if full
        if activeCount >= int64(zone.TotalCapacity) {
            return ErrZoneFull
        }

        // Step 4: Create the reservation — only reachable if capacity allows
        return tx.Create(reservation).Error
    })
}
```

**Rules:**

- `clause.Locking{Strength: "UPDATE"}` must be imported from `gorm.io/gorm/clause`
- The lock must be acquired on the zone row before counting reservations — never count before locking
- The transaction automatically commits if the callback returns nil, and rolls back if it returns an error
- `ErrZoneFull` and `ErrNotFound` are sentinel errors — return them directly, the transaction rolls back automatically
- Always check `gorm.ErrRecordNotFound` inside repositories and map to domain sentinel `ErrNotFound`
- Never call `db.Transaction` from a service or handler — only from the repository layer

---

## golang-jwt/jwt — Authentication

Import: `github.com/golang-jwt/jwt/v5`

### JWT Claims Struct

```go
import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
    UserID uint   `json:"id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

### Signing a Token (in auth_service.go)

```go
func generateToken(userID uint, role string) (string, error) {
    claims := JWTClaims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

### Verifying a Token (in jwt_middleware.go)

```go
func JWTMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "success": false,
                    "message": "Missing or invalid authorization header",
                })
            }

            tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
            token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return []byte(os.Getenv("JWT_SECRET")), nil
            })

            if err != nil || !token.Valid {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "success": false,
                    "message": "Invalid or expired token",
                })
            }

            claims := token.Claims.(*JWTClaims)
            c.Set("userID", claims.UserID)
            c.Set("role", claims.Role)

            return next(c)
        }
    }
}
```

### Role Middleware (in role_middleware.go)

```go
func RoleMiddleware(requiredRole string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            role, ok := c.Get("role").(string)
            if !ok || role != requiredRole {
                return c.JSON(http.StatusForbidden, map[string]interface{}{
                    "success": false,
                    "message": "Forbidden: insufficient permissions",
                })
            }
            return next(c)
        }
    }
}
```

**Rules:**

- JWT payload must always contain `id` (uint) and `role` (string) — no exceptions
- Always use `HS256` signing method
- Token expiry is 24 hours — defined once as a constant
- `JWT_SECRET` always comes from environment — never hardcoded
- Always validate the signing method inside the key function — prevents algorithm confusion attacks
- `jwt_middleware` always runs before `role_middleware` — never apply role check without first verifying the token

---

## bcrypt — Password Hashing

Import: `golang.org/x/crypto/bcrypt`

### Hashing

```go
const BcryptCost = 12

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
    return string(bytes), err
}
```

### Verifying

```go
func checkPassword(hash, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**Rules:**

- Cost is always 10–12. Use 12 for production.
- Always store only the hash — never the plaintext password
- Never log or return the password or its hash
- `CompareHashAndPassword` returns `bcrypt.ErrMismatchedHashAndPassword` on wrong password — map to a generic "invalid credentials" error, never expose which field was wrong

---

## go-playground/validator — Request Validation

Import: `github.com/go-playground/validator/v10`

### Registration with Echo

```go
type customValidator struct {
    validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        return err
    }
    return nil
}

// In main():
e.Validator = &customValidator{validator: validator.New()}
```

### Validator Tags on DTOs

```go
type RegisterRequest struct {
    Name     string `json:"name"     validate:"required"`
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Role     string `json:"role"     validate:"required,oneof=driver admin"`
}

type CreateZoneRequest struct {
    Name           string  `json:"name"            validate:"required"`
    Type           string  `json:"type"            validate:"required,oneof=general ev_charging covered"`
    TotalCapacity  int     `json:"total_capacity"  validate:"required,gt=0"`
    PricePerHour   float64 `json:"price_per_hour"  validate:"required,gt=0"`
}

type CreateReservationRequest struct {
    ZoneID       uint   `json:"zone_id"       validate:"required,gt=0"`
    LicensePlate string `json:"license_plate" validate:"required,max=15"`
}
```

### Common Tags

| Tag           | Meaning                                |
| ------------- | -------------------------------------- |
| `required`    | Field must be present and non-zero     |
| `email`       | Must be a valid email format           |
| `min=8`       | String must be at least 8 characters   |
| `max=15`      | String must be at most 15 characters   |
| `gt=0`        | Number must be greater than 0          |
| `oneof=a b c` | Value must be one of the listed values |

**Rules:**

- Every request DTO must have `validate` tags on every field
- Always call `c.Validate(req)` immediately after `c.Bind(req)` — never skip validation
- Validation errors return HTTP 400 — never 500
- Never add business logic validation in DTO tags — only format/type validation belongs here. Business rules (e.g. zone full, email taken) belong in the service layer

---

## godotenv — Environment Loading

Import: `github.com/joho/godotenv`

### Usage

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file in development — silently ignored if file doesn't exist (production uses real env vars)
    godotenv.Load()

    // Read values
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
}
```

**Rules:**

- Call `godotenv.Load()` once at the top of `main()` — before any `os.Getenv` call
- Never use `godotenv.MustLoad()` in production — it panics if `.env` is missing, which is expected in deployed environments
- Never commit `.env` — always commit `.env.example` with placeholder values
- Production environment variables are set directly on the platform (Render / Railway / Fly.io)
