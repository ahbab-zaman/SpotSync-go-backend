# Build Plan

## Core Principle

Build one layer at a time, bottom-up. Models first, then repositories, then services, then handlers. Each layer is complete and tested before the next is added. No handler is written without its service. No service is written without its repository. No skipping ahead.

---

## Phase 1 — Project Foundation

### 01 Project Scaffolding

Initialize the Go module and install all dependencies.

**Tasks:**

- `go mod init github.com/yourusername/spotsync`
- Install Echo: `go get github.com/labstack/echo/v4`
- Install GORM + PostgreSQL driver: `go get gorm.io/gorm gorm.io/driver/postgres`
- Install JWT: `go get github.com/golang-jwt/jwt/v5`
- Install bcrypt: `go get golang.org/x/crypto/bcrypt`
- Install validator: `go get github.com/go-playground/validator/v10`
- Install godotenv: `go get github.com/joho/godotenv`
- Create folder structure: `models/`, `dto/`, `repository/`, `service/`, `handler/`, `middleware/`
- Create `.env` with `DATABASE_URL`, `JWT_SECRET`, `PORT`
- Create `.env.example` with placeholder values
- Create `.gitignore` ignoring `.env` and binary outputs

---

### 02 Database Connection + Models

Define all three GORM models and confirm auto-migration creates the tables.

**Tasks:**

- `models/user.go` — User struct with `id`, `name`, `email`, `password`, `role` (default `driver`), `created_at`, `updated_at`
- `models/parking_zone.go` — ParkingZone struct with `id`, `name`, `type`, `total_capacity`, `price_per_hour`, `created_at`, `updated_at`
- `models/reservation.go` — Reservation struct with `id`, `user_id`, `zone_id`, `license_plate`, `status` (default `active`), `created_at`, `updated_at`. Include `BelongsTo` associations for User and ParkingZone.
- `main.go` — Open GORM PostgreSQL connection from `DATABASE_URL`. Call `AutoMigrate` on all three models. Confirm tables are created in the database.

**Verification:** Connect to the PostgreSQL database and confirm all three tables exist with the correct columns.

---

### 03 Middleware — JWT + Role

Build authentication and authorization middleware before any protected routes exist.

**Tasks:**

- `middleware/jwt_middleware.go` — Echo middleware that reads `Authorization: Bearer <token>` header, verifies the JWT signature using `JWT_SECRET`, extracts `id` and `role` claims, and stores them in the Echo context (`c.Set("userID", id)`, `c.Set("role", role)`). Returns 401 if token is missing or invalid.
- `middleware/role_middleware.go` — Echo middleware factory that accepts a required role string, reads `role` from Echo context, returns 403 if it does not match.

**Verification:** Write a test route protected by `jwt_middleware`. Confirm 401 is returned without a token and 200 with a valid token.

---

## Phase 2 — Auth Module

### 04 DTOs — Auth

Define all request and response types for authentication.

**Tasks:**

- `dto/auth_dto.go`:
  - `RegisterRequest` — `name`, `email`, `password`, `role` with validator tags (`required`, `email`, `oneof=driver admin`)
  - `LoginRequest` — `email`, `password` with validator tags
  - `UserResponse` — `id`, `name`, `email`, `role`, `created_at`, `updated_at` (no password)
  - `LoginResponse` — `token` string, `user` UserResponse

---

### 05 Repository — User

**Tasks:**

- `repository/user_repository.go`:
  - `CreateUser(user *models.User) error`
  - `FindByEmail(email string) (*models.User, error)`
  - `FindByID(id uint) (*models.User, error)`

---

### 06 Service — Auth

**Tasks:**

- `service/auth_service.go`:
  - `Register(req dto.RegisterRequest) (*dto.UserResponse, error)` — check email uniqueness, hash password with bcrypt (cost 10), save user, return UserResponse
  - `Login(req dto.LoginRequest) (*dto.LoginResponse, error)` — find user by email, compare bcrypt hash, sign JWT with `id` and `role` in payload, return LoginResponse

---

### 07 Handler — Auth

**Tasks:**

- `handler/auth_handler.go`:
  - `POST /api/v1/auth/register` — bind RegisterRequest, validate, call service, return 201 with UserResponse
  - `POST /api/v1/auth/login` — bind LoginRequest, validate, call service, return 200 with LoginResponse
- Register routes on Echo in `main.go`

**Verification:** `POST /api/v1/auth/register` creates a user and returns the correct shape. `POST /api/v1/auth/login` returns a JWT. Duplicate email returns 400. Wrong password returns 401.

---

## Phase 3 — Parking Zones Module

### 08 DTOs — Zones

**Tasks:**

- `dto/zone_dto.go`:
  - `CreateZoneRequest` — `name`, `type`, `total_capacity`, `price_per_hour` with validator tags (`required`, `oneof=general ev_charging covered`, `gt=0`)
  - `UpdateZoneRequest` — same fields, all optional for partial updates
  - `ZoneResponse` — `id`, `name`, `type`, `total_capacity`, `available_spots`, `price_per_hour`, `created_at`

---

### 09 Repository — Zone

**Tasks:**

- `repository/zone_repository.go`:
  - `FindAll() ([]models.ParkingZone, error)`
  - `FindByID(id uint) (*models.ParkingZone, error)`
  - `Create(zone *models.ParkingZone) error`
  - `Update(zone *models.ParkingZone) error`
  - `Delete(id uint) error`
  - `CountActiveReservations(zoneID uint) (int64, error)` — counts reservations WHERE zone_id = ? AND status = 'active'

---

### 10 Service — Zone

**Tasks:**

- `service/zone_service.go`:
  - `GetAll() ([]dto.ZoneResponse, error)` — fetch all zones, calculate `available_spots` for each using `CountActiveReservations`
  - `GetByID(id uint) (*dto.ZoneResponse, error)` — same as above for single zone
  - `Create(req dto.CreateZoneRequest) (*dto.ZoneResponse, error)`
  - `Update(id uint, req dto.UpdateZoneRequest) (*dto.ZoneResponse, error)`
  - `Delete(id uint) error`

---

### 11 Handler — Zone

**Tasks:**

- `handler/zone_handler.go`:
  - `GET /api/v1/zones` — public, return all zones with `available_spots`
  - `GET /api/v1/zones/:id` — public, return single zone with `available_spots`
  - `POST /api/v1/zones` — admin only (jwt_middleware + role_middleware), return 201
  - `PUT /api/v1/zones/:id` — admin only, return 200
  - `DELETE /api/v1/zones/:id` — admin only, return 200
- Register all routes in `main.go` with correct middleware chains

**Verification:** Create a zone as admin. List zones and confirm `available_spots` equals `total_capacity` (no reservations yet). Attempt create as driver — confirm 403.

---

## Phase 4 — Reservations Module (Concurrency-Critical)

### 12 DTOs — Reservations

**Tasks:**

- `dto/reservation_dto.go`:
  - `CreateReservationRequest` — `zone_id`, `license_plate` with validator tags (`required`, `max=15`)
  - `ReservationResponse` — `id`, `user_id`, `zone_id`, `license_plate`, `status`, `created_at`, `updated_at`
  - `MyReservationResponse` — `id`, `license_plate`, `status`, `zone` (nested: `id`, `name`, `type`), `created_at`
  - `AdminReservationResponse` — full reservation with preloaded `user` (id, name, email) and `zone` (id, name, type)

---

### 13 Repository — Reservation (The Critical One)

**Tasks:**

- `repository/reservation_repository.go`:
  - `CreateWithLock(reservation *models.Reservation, zoneID uint) error` — opens a GORM transaction, locks the zone row with `clause.Locking{Strength: "UPDATE"}`, counts active reservations, checks capacity, creates reservation if capacity allows, returns `ErrZoneFull` if full
  - `FindByUserID(userID uint) ([]models.Reservation, error)` — preload Zone association
  - `FindByID(id uint) (*models.Reservation, error)`
  - `UpdateStatus(id uint, status string) error`
  - `FindAll() ([]models.Reservation, error)` — preload User and Zone associations

**The transaction pattern is mandatory:**

```go
func (r *reservationRepository) CreateWithLock(reservation *models.Reservation, zoneID uint) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        var zone models.ParkingZone
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&zone, zoneID).Error; err != nil {
            return err
        }
        var activeCount int64
        tx.Model(&models.Reservation{}).
            Where("zone_id = ? AND status = ?", zoneID, "active").
            Count(&activeCount)
        if activeCount >= int64(zone.TotalCapacity) {
            return ErrZoneFull
        }
        return tx.Create(reservation).Error
    })
}
```

---

### 14 Service — Reservation

**Tasks:**

- `service/reservation_service.go`:
  - `Reserve(userID uint, req dto.CreateReservationRequest) (*dto.ReservationResponse, error)` — verify zone exists, call `CreateWithLock`, map to response DTO
  - `GetMyReservations(userID uint) ([]dto.MyReservationResponse, error)`
  - `CancelReservation(reservationID uint, userID uint) error` — find reservation, verify ownership (return ErrForbidden if not owner), update status to `cancelled`
  - `GetAllReservations() ([]dto.AdminReservationResponse, error)`

---

### 15 Handler — Reservation

**Tasks:**

- `handler/reservation_handler.go`:
  - `POST /api/v1/reservations` — jwt_middleware required, extract userID from context, bind and validate request, call service, return 201
  - `GET /api/v1/reservations/my-reservations` — jwt_middleware required, extract userID, return 200
  - `DELETE /api/v1/reservations/:id` — jwt_middleware required, extract userID, call service, handle ErrForbidden → 403, return 200
  - `GET /api/v1/reservations` — jwt_middleware + role_middleware(admin), return 200
- Register all routes in `main.go`

**Verification:** Create two concurrent reservation requests for the last available spot. Confirm exactly one succeeds (201) and one is rejected (409). Confirm a driver cannot cancel another driver's reservation (403).

---

## Phase 5 — Deployment

### 16 Deployment

**Tasks:**

- Push final code to GitHub (public repository, minimum 10 meaningful commits)
- Provision PostgreSQL on NeonDB, Supabase, or Aiven
- Deploy backend to Render, Railway, or Fly.io
- Set environment variables (`DATABASE_URL`, `JWT_SECRET`, `PORT`) in the deployment platform
- Confirm CORS is configured if any web client will hit the API
- Confirm all 9 endpoints are reachable on the live URL
- Write `README.md` with: project name, live URL, features, tech stack, architecture explanation, setup steps, required `.env` variables, full API endpoint list

---

## Feature Count

| Phase                   | Features |
| ----------------------- | -------- |
| Phase 1 — Foundation    | 3        |
| Phase 2 — Auth Module   | 4        |
| Phase 3 — Parking Zones | 4        |
| Phase 4 — Reservations  | 4        |
| Phase 5 — Deployment    | 1        |
| **Total**               | **16**   |
