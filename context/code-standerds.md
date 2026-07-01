# Code Standards

Implementation rules and conventions for the entire SpotSync project. Follow these in every session without exception. These rules prevent pattern drift across sessions.

---

## Engineering Mindset

- **Think before implementing** — understand what is being built and why before writing a single line
- **Read context files first** — never assume; always verify against architecture.md and project-overview.md
- **Scope is sacred** — only build what the current feature requires
- **Every feature must be testable** — if it cannot be verified immediately after implementation, it is incomplete
- **Clean over clever** — simple readable Go that a junior developer can understand is preferred over clever abstractions
- **One thing at a time** — complete one feature fully (DTO → Repository → Service → Handler) before touching the next
- **Layers must never be skipped** — Handler → Service → Repository → DB. No shortcuts.

---

## Go Conventions

- Go 1.22 or higher
- Always run `go fmt ./...` before committing
- Always run `go vet ./...` before committing
- Use `errors.New` or `fmt.Errorf` for error creation — never use blank identifiers to swallow errors
- All exported functions and types must have a doc comment
- Use named return values only when it genuinely improves clarity
- Prefer `if err != nil { return ..., err }` — never ignore errors
- Use `const` for fixed values (e.g., bcrypt cost, token expiry)
- Group imports: standard library first, then third-party, then internal packages — separated by blank lines

```go
import (
    "errors"
    "net/http"

    "github.com/labstack/echo/v4"
    "gorm.io/gorm"

    "github.com/yourusername/spotsync/dto"
    "github.com/yourusername/spotsync/models"
)
```

---

## Naming Conventions

| Thing                | Convention                            | Example                          |
| -------------------- | ------------------------------------- | -------------------------------- |
| Packages             | lowercase                             | `handler`, `service`, `dto`      |
| Files                | snake_case                            | `auth_handler.go`, `zone_dto.go` |
| Exported types       | PascalCase                            | `UserRepository`, `AuthService`  |
| Unexported variables | camelCase                             | `jwtSecret`, `dbConn`            |
| Constants            | PascalCase or SCREAMING_SNAKE for env | `BcryptCost`, `JWT_SECRET`       |
| Interfaces           | PascalCase, noun                      | `UserRepository`, `AuthService`  |

---

## Layer Rules

### Models (`models/`)

```go
// models/user.go
package models

import "time"

type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"not null" json:"name"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"`  // json:"-" hides password always
    Role      string    `gorm:"default:driver;not null" json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

- Models are pure GORM structs. No HTTP logic. No business logic.
- Password field always has `json:"-"` — it must never appear in any JSON output.
- Use `gorm` struct tags for constraints and defaults.
- Use `json` struct tags for API field names.

---

### DTOs (`dto/`)

```go
// dto/auth_dto.go
package dto

import "time"

type RegisterRequest struct {
    Name     string `json:"name"     validate:"required"`
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Role     string `json:"role"     validate:"required,oneof=driver admin"`
}

type UserResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

- DTOs are the API contract. Never return raw model structs from handlers.
- Request DTOs carry `validate` tags for go-playground/validator.
- Response DTOs never include passwords or internal GORM metadata.
- One file per module: `auth_dto.go`, `zone_dto.go`, `reservation_dto.go`.

---

### Repositories (`repository/`)

```go
// repository/user_repository.go
package repository

import (
    "gorm.io/gorm"
    "github.com/yourusername/spotsync/models"
)

type UserRepository interface {
    CreateUser(user *models.User) error
    FindByEmail(email string) (*models.User, error)
    FindByID(id uint) (*models.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

- Always define an interface and a concrete struct.
- Constructor: `NewXxxRepository(db *gorm.DB) XxxRepository`.
- Never return HTTP errors. Return Go errors only.
- All GORM access lives here. No GORM imports in handlers or services.
- Handle `gorm.ErrRecordNotFound` here — map to `ErrNotFound` sentinel.

---

### Services (`service/`)

```go
// service/auth_service.go
package service

import (
    "github.com/yourusername/spotsync/dto"
    "github.com/yourusername/spotsync/repository"
)

type AuthService interface {
    Register(req dto.RegisterRequest) (*dto.UserResponse, error)
    Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
    userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
    return &authService{userRepo: userRepo}
}
```

- Always define an interface and a concrete struct.
- Constructor: `NewXxxService(dep XxxRepository) XxxService`.
- Services receive and return plain Go types and DTOs — never Echo types.
- Business logic only: bcrypt, JWT, capacity checks, ownership checks.
- Call repository methods — never call GORM directly.

---

### Handlers (`handler/`)

```go
// handler/auth_handler.go
package handler

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/yourusername/spotsync/dto"
    "github.com/yourusername/spotsync/service"
)

type AuthHandler struct {
    authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

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
    result, err := h.authService.Register(req)
    if err != nil {
        // map sentinel errors to HTTP codes here
        return handleServiceError(c, err)
    }
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "success": true,
        "message": "User registered successfully",
        "data":    result,
    })
}
```

- Handlers only: bind, validate, call service, return JSON.
- No business logic. No GORM. No bcrypt. No JWT signing.
- Error mapping from service sentinel errors to HTTP codes happens in `handleServiceError` (a shared helper in the handler package).
- All responses follow the standard shape: `{ success, message, data }` or `{ success, message, errors }`.

---

## Standard Response Helper

Define a central response helper used by all handlers:

```go
// handler/response.go
package handler

import (
    "errors"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/yourusername/spotsync/service"
)

func success(c echo.Context, code int, message string, data interface{}) error {
    return c.JSON(code, map[string]interface{}{
        "success": true,
        "message": message,
        "data":    data,
    })
}

func fail(c echo.Context, code int, message string, errs interface{}) error {
    return c.JSON(code, map[string]interface{}{
        "success": false,
        "message": message,
        "errors":  errs,
    })
}

func handleServiceError(c echo.Context, err error) error {
    switch {
    case errors.Is(err, service.ErrZoneFull):
        return fail(c, http.StatusConflict, "Zone is at full capacity", err.Error())
    case errors.Is(err, service.ErrNotFound):
        return fail(c, http.StatusNotFound, "Resource not found", err.Error())
    case errors.Is(err, service.ErrForbidden):
        return fail(c, http.StatusForbidden, "Forbidden", err.Error())
    case errors.Is(err, service.ErrDuplicateEmail):
        return fail(c, http.StatusBadRequest, "Email already registered", err.Error())
    default:
        return fail(c, http.StatusInternalServerError, "Internal server error", nil)
    }
}
```

---

## Validation Setup

Register a custom validator with Echo in `main.go`:

```go
type customValidator struct {
    validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

// In main():
e.Validator = &customValidator{validator: validator.New()}
```

---

## JWT Claims

```go
type JWTClaims struct {
    UserID uint   `json:"id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

Always include `id` and `role` in the JWT payload. Standard claims (exp, iat) are included via `RegisteredClaims`.

---

## Bcrypt

```go
const BcryptCost = 12

// Hash
hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)

// Verify
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
```

Cost is always 10–12. Never store raw passwords. Never log passwords.

---

## Dependency Injection in main.go

```go
func main() {
    // 1. Load env
    godotenv.Load()

    // 2. Connect DB
    db := connectDB()
    db.AutoMigrate(&models.User{}, &models.ParkingZone{}, &models.Reservation{})

    // 3. Instantiate repositories
    userRepo := repository.NewUserRepository(db)
    zoneRepo := repository.NewZoneRepository(db)
    reservationRepo := repository.NewReservationRepository(db)

    // 4. Instantiate services
    authSvc := service.NewAuthService(userRepo)
    zoneSvc := service.NewZoneService(zoneRepo)
    reservationSvc := service.NewReservationService(reservationRepo, zoneRepo)

    // 5. Instantiate handlers
    authHandler := handler.NewAuthHandler(authSvc)
    zoneHandler := handler.NewZoneHandler(zoneSvc)
    reservationHandler := handler.NewReservationHandler(reservationSvc)

    // 6. Echo setup
    e := echo.New()
    e.Validator = &customValidator{validator: validator.New()}

    // 7. Route registration
    registerRoutes(e, authHandler, zoneHandler, reservationHandler)

    // 8. Start server
    e.Start(":" + os.Getenv("PORT"))
}
```

---

## Route Registration

```go
func registerRoutes(e *echo.Echo, auth *handler.AuthHandler, zone *handler.ZoneHandler, res *handler.ReservationHandler) {
    api := e.Group("/api/v1")

    // Public auth
    api.POST("/auth/register", auth.Register)
    api.POST("/auth/login", auth.Login)

    // Public zones
    api.GET("/zones", zone.GetAll)
    api.GET("/zones/:id", zone.GetByID)

    // Admin zones
    adminZones := api.Group("/zones", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))
    adminZones.POST("", zone.Create)
    adminZones.PUT("/:id", zone.Update)
    adminZones.DELETE("/:id", zone.Delete)

    // Authenticated reservations
    authRes := api.Group("/reservations", middleware.JWTMiddleware())
    authRes.POST("", res.Create)
    authRes.GET("/my-reservations", res.GetMyReservations)
    authRes.DELETE("/:id", res.Cancel)

    // Admin reservations
    api.GET("/reservations", res.GetAll, middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))
}
```

---

## Error Handling Rules

- Never use empty catch/ignore patterns — always handle `err != nil`
- Never expose raw GORM errors to the client — map to sentinel errors in the repository, map sentinels to HTTP codes in the handler
- Log internal errors server-side with context: `log.Printf("[auth_service.Register] %v", err)`
- User-facing error messages must be human-readable — never include stack traces or GORM internals

---

## Commit Discipline

- Minimum 10 meaningful commits
- Commit message format: `<scope>: <short description>`
  - `setup: initialize Go module and install dependencies`
  - `models: add User, ParkingZone, Reservation GORM structs`
  - `middleware: implement JWT and role middleware`
  - `auth: add register and login endpoints`
  - `zones: implement CRUD with available_spots calculation`
  - `reservations: implement FOR UPDATE transaction to fix race condition`
  - `deploy: configure Railway deployment and environment variables`
- Never commit `.env` — only `.env.example`
- Never commit binary outputs

---

## Environment Variables

| Variable       | Description                       |
| -------------- | --------------------------------- |
| `DATABASE_URL` | Full PostgreSQL connection string |
| `JWT_SECRET`   | Secret key for signing JWTs       |
| `PORT`         | Port for Echo HTTP server         |

Read with `os.Getenv`. Load `.env` in development with `godotenv.Load()`.

---

## Sentinel Errors

Defined in the `service` package:

```go
package service

import "errors"

var (
    ErrZoneFull       = errors.New("zone is at full capacity")
    ErrNotFound       = errors.New("resource not found")
    ErrForbidden      = errors.New("forbidden")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrDuplicateEmail = errors.New("email already registered")
)
```

- Repositories return `ErrNotFound` when `gorm.ErrRecordNotFound` occurs
- Services return domain sentinels based on business rules
- Handlers map sentinels to HTTP status codes using `handleServiceError`
- Never return GORM errors directly from a service or handler
