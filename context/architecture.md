# Architecture

## Stack

| Layer         | Tool                                   | Purpose                                              |
| ------------- | -------------------------------------- | ---------------------------------------------------- |
| Language      | Go 1.22+                               | Entire backend                                       |
| Web Framework | Echo (`github.com/labstack/echo/v4`)   | HTTP routing, middleware, request/response           |
| ORM           | GORM (`gorm.io/gorm`)                  | Database access, transactions, row locking           |
| Database      | PostgreSQL (NeonDB / Supabase / Aiven) | Relational data storage                              |
| Validation    | go-playground/validator v10            | Struct-level request validation integrated with Echo |
| Auth          | golang-jwt/jwt v5                      | JWT signing and verification                         |
| Password Hash | golang.org/x/crypto/bcrypt             | Bcrypt password hashing, cost 10–12                  |
| Config        | godotenv or os.Getenv                  | Environment variable loading from `.env`             |

---

## Clean Architecture — Layer Rules (Strict)

Every layer has one job. Layers only communicate downward. **Handlers never touch the database.**

| Layer          | Directory     | Responsibility                                                                                       |
| -------------- | ------------- | ---------------------------------------------------------------------------------------------------- |
| **DTO**        | `dto/`        | Request payloads and response shapes. Never expose GORM models directly to the API.                  |
| **Handler**    | `handler/`    | HTTP layer. Bind and validate DTOs. Extract JWT claims from Echo context. Call Service. Return JSON. |
| **Service**    | `service/`    | Business logic. Hash passwords. Generate JWTs. Enforce capacity rules. Call Repository.              |
| **Repository** | `repository/` | Data access only. All GORM operations — CRUD, transactions, row locks.                               |
| **Models**     | `models/`     | GORM structs representing database tables. No HTTP or business logic here.                           |

---

## Dependency Injection

Layers are wired manually in `main.go`. No DI framework.

```
Instantiate Repository → Pass to Service → Pass to Handler → Register on Echo router
```

---

## Folder Structure

```
/
├── main.go                          → Entry point. DI wiring, Echo setup, route registration
├── .env                             → Environment variables (never committed)
├── .env.example                     → Example env file (committed)
├── go.mod
├── go.sum
│
├── models/
│   ├── user.go                      → User GORM struct
│   ├── parking_zone.go              → ParkingZone GORM struct
│   └── reservation.go               → Reservation GORM struct
│
├── dto/
│   ├── auth_dto.go                  → RegisterRequest, LoginRequest, LoginResponse, UserResponse
│   ├── zone_dto.go                  → CreateZoneRequest, UpdateZoneRequest, ZoneResponse, ZoneListResponse
│   └── reservation_dto.go           → CreateReservationRequest, ReservationResponse, MyReservationResponse, AdminReservationResponse
│
├── repository/
│   ├── user_repository.go           → FindByEmail, CreateUser
│   ├── zone_repository.go           → FindAll, FindByID, Create, Update, Delete, CountActiveReservations
│   └── reservation_repository.go    → CreateWithLock (transaction + FOR UPDATE), FindByUserID, FindByID, UpdateStatus, FindAll
│
├── service/
│   ├── auth_service.go              → Register (hash password, save user), Login (verify hash, issue JWT)
│   ├── zone_service.go              → List zones with available_spots, Get zone, Create, Update, Delete
│   └── reservation_service.go       → Reserve (capacity check inside transaction), GetMyReservations, CancelReservation, GetAll
│
├── handler/
│   ├── auth_handler.go              → POST /auth/register, POST /auth/login
│   ├── zone_handler.go              → GET /zones, GET /zones/:id, POST /zones, PUT /zones/:id, DELETE /zones/:id
│   └── reservation_handler.go       → POST /reservations, GET /reservations/my-reservations, DELETE /reservations/:id, GET /reservations
│
└── middleware/
    ├── jwt_middleware.go            → Verify JWT signature, inject claims into Echo context
    └── role_middleware.go           → Check role claim from context, return 403 if insufficient
```

---

## Database Schema

### Table: `users`

| Column       | Type        | Constraints                                              |
| ------------ | ----------- | -------------------------------------------------------- |
| `id`         | serial      | PRIMARY KEY                                              |
| `name`       | varchar     | NOT NULL                                                 |
| `email`      | varchar     | NOT NULL, UNIQUE                                         |
| `password`   | varchar     | NOT NULL (bcrypt hash)                                   |
| `role`       | varchar     | NOT NULL, DEFAULT `driver`, CHECK IN (`driver`, `admin`) |
| `created_at` | timestamptz | Auto-generated                                           |
| `updated_at` | timestamptz | Auto-refreshed                                           |

### Table: `parking_zones`

| Column           | Type        | Constraints                                              |
| ---------------- | ----------- | -------------------------------------------------------- |
| `id`             | serial      | PRIMARY KEY                                              |
| `name`           | varchar     | NOT NULL                                                 |
| `type`           | varchar     | NOT NULL, CHECK IN (`general`, `ev_charging`, `covered`) |
| `total_capacity` | integer     | NOT NULL, CHECK > 0                                      |
| `price_per_hour` | decimal     | NOT NULL, CHECK > 0                                      |
| `created_at`     | timestamptz | Auto-generated                                           |
| `updated_at`     | timestamptz | Auto-refreshed                                           |

### Table: `reservations`

| Column          | Type        | Constraints                                                               |
| --------------- | ----------- | ------------------------------------------------------------------------- |
| `id`            | serial      | PRIMARY KEY                                                               |
| `user_id`       | integer     | NOT NULL, FOREIGN KEY → users(id)                                         |
| `zone_id`       | integer     | NOT NULL, FOREIGN KEY → parking_zones(id)                                 |
| `license_plate` | varchar(15) | NOT NULL                                                                  |
| `status`        | varchar     | NOT NULL, DEFAULT `active`, CHECK IN (`active`, `completed`, `cancelled`) |
| `created_at`    | timestamptz | Auto-generated                                                            |
| `updated_at`    | timestamptz | Auto-refreshed                                                            |

---

## Authentication Flow

```
Client → POST /auth/login (email + password)
       → Service verifies bcrypt hash
       → Service signs JWT with { id, role } in payload
       → Response returns token

Client → Attach token: Authorization: Bearer <token>
       → jwt_middleware.go verifies signature
       → Claims injected into Echo context
       → Handler or role_middleware checks role
```

JWT payload must always include:

- `id` — user's integer ID
- `role` — `driver` or `admin`

---

## Concurrency Safety — The EV Spot Bottleneck

The reservation endpoint is the most critical part of the system. Two simultaneous requests for the last available spot must result in exactly one success and one 409 Conflict.

**Implementation in `repository/reservation_repository.go`:**

```go
db.Transaction(func(tx *gorm.DB) error {
    var zone models.ParkingZone
    // 1. Lock the zone row exclusively
    if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        First(&zone, zoneID).Error; err != nil {
        return err
    }
    // 2. Count active reservations for this zone
    var activeCount int64
    tx.Model(&models.Reservation{}).
        Where("zone_id = ? AND status = ?", zoneID, "active").
        Count(&activeCount)
    // 3. Enforce capacity
    if activeCount >= int64(zone.TotalCapacity) {
        return ErrZoneFull // custom sentinel error
    }
    // 4. Create reservation — only reachable if capacity allows
    return tx.Create(&reservation).Error
})
```

This pattern guarantees atomicity. The `FOR UPDATE` lock prevents any other transaction from reading a stale capacity count until this transaction commits or rolls back.

---

## available_spots Calculation

`available_spots` is never stored. It is calculated dynamically every time zones are fetched:

```
available_spots = total_capacity - COUNT(reservations WHERE zone_id = ? AND status = 'active')
```

Implemented in the Service layer or via a GORM subquery in the Select clause.

---

## Error Sentinel Values

Define custom errors in the service or repository layer for business logic failures. These are mapped to HTTP status codes in the handler.

```go
var (
    ErrZoneFull         = errors.New("zone is at full capacity")
    ErrNotFound         = errors.New("resource not found")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrForbidden        = errors.New("forbidden")
    ErrDuplicateEmail   = errors.New("email already registered")
)
```

---

## Middleware Chain

```
Public routes:           No middleware
Authenticated routes:    jwt_middleware → handler
Admin-only routes:       jwt_middleware → role_middleware(admin) → handler
```

---

## Environment Variables

| Variable       | Used In                     |
| -------------- | --------------------------- |
| `DATABASE_URL` | main.go GORM connection     |
| `JWT_SECRET`   | auth_service.go JWT signing |
| `PORT`         | main.go Echo server listen  |

Never hardcode any of these values. Always read from environment.

---

## Invariants

Rules that must never be violated:

- Handlers never import or call GORM directly. All DB access goes through Repository.
- Services never import Echo or return `echo.Context`. They receive plain Go types and return plain Go types or errors.
- Repositories never contain HTTP logic, status codes, or JSON.
- DTOs are never GORM models. Models are never returned directly from handlers.
- The `FOR UPDATE` lock inside a transaction is required on every reservation creation. No exceptions.
- Passwords are never logged, returned in responses, or stored in plaintext.
- Every endpoint that requires auth must pass through `jwt_middleware`.
- Every admin-only endpoint must pass through `role_middleware` after `jwt_middleware`.
- `available_spots` is always computed at read time — never stored in the database.
- Custom error sentinels are mapped to HTTP status codes only in the handler layer.
