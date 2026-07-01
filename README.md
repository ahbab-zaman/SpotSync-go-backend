# SpotSync — Parking Reservation API

A RESTful parking reservation backend built with Go (Echo), GORM, and PostgreSQL. Provides JWT-authenticated endpoints for user registration, parking zone management, and concurrent-safe reservation booking.

**Live URL:** `https://spotsync.onrender.com` (or your deployed URL)

---

## Features

- **User authentication** — Register and login with bcrypt-hashed passwords and JWT tokens (24h expiry)
- **Role-based access** — `driver` and `admin` roles; admin-only endpoints for managing parking zones
- **Parking zones** — CRUD for zones with `general`, `ev_charging`, and `covered` types
- **Dynamic availability** — `available_spots` computed in real time from active reservations
- **Concurrent reservations** — `FOR UPDATE` row locking inside a database transaction prevents over-booking
- **Ownership enforcement** — Drivers can only cancel their own reservations
- **Validation** — Request payloads validated with `go-playground/validator`

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| HTTP framework | Echo v4 |
| ORM | GORM v1.31 |
| Database | PostgreSQL 16+ |
| Auth | JWT (golang-jwt v5) + bcrypt |
| Validation | go-playground/validator v10 |
| Hosting | Render / Railway / Fly.io |

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                      HTTP Client                     │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                    Echo Router                       │
│           /api/v1/*  +  Middleware Chain             │
│   (Logger, Recover, CORS, JWT, Role-based auth)     │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                   Handlers                           │
│        Bind JSON → Validate → Call Service           │
│        Map sentinel errors → HTTP status codes       │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                   Services                           │
│      Business logic, sentinel errors, DTO mapping    │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                 Repositories                         │
│              GORM database operations                │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                  PostgreSQL                          │
│          users, parking_zones, reservations          │
└─────────────────────────────────────────────────────┘
```

### Clean Architecture Boundaries

- **Models** — GORM structs only. No JSON tags for responses, no validation logic.
- **DTOs** — Request/response types, validation tags, JSON shape definitions.
- **Repositories** — Pure GORM operations. No business rules, no HTTP concerns.
- **Services** — Business logic, sentinel errors, DTO mapping. No Echo/GORM imports.
- **Handlers** — Bind, validate, call service, return JSON. No GORM/bcrypt/JWT logic.

---

## Setup

### Prerequisites

- Go 1.26+
- PostgreSQL 16+
- Git

### Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/spotsync.git
cd spotsync

# Copy environment file and fill in your values
cp .env.example .env

# Run the project
go run main.go
```

### Default Credentials

After running the project, you can log in with these pre-seeded credentials:

| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@main.com` | `AdminPass123!` |

> **Note:** Any user can register via `POST /api/v1/auth/register`. These credentials are for the admin account that already exists in the database.

---

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/db?sslmode=require` |
| `JWT_SECRET` | Secret key for signing JWT tokens | `your-secret-key-change-in-production` |
| `PORT` | Server port (default: 8080) | `8080` |

---

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Standard Response Envelope

**Success:**
```json
{
  "success": true,
  "message": "Human readable message",
  "data": {}
}
```

**Error:**
```json
{
  "success": false,
  "message": "Error description",
  "errors": null
}
```

---

### Auth

#### POST /api/v1/auth/register

Public. Create a new user account.

**Request:**
```json
{
  "name": "John Doe",
  "email": "john.doe@spotsync.com",
  "password": "securePassword123",
  "role": "driver"
}
```

**Response — 201 Created:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@spotsync.com",
    "role": "driver",
    "created_at": "2026-06-20T09:00:00Z",
    "updated_at": "2026-06-20T09:00:00Z"
  }
}
```

| Error | Status | Message |
|-------|--------|---------|
| Validation failure | 400 | Validation failed |
| Email already taken | 400 | Email already registered |

---

#### POST /api/v1/auth/login

Public. Authenticate and receive a JWT token.

**Request:**
```json
{
  "email": "john.doe@spotsync.com",
  "password": "securePassword123"
}
```

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john.doe@spotsync.com",
      "role": "driver"
    }
  }
}
```

| Error | Status | Message |
|-------|--------|---------|
| Validation failure | 400 | Validation failed |
| Wrong credentials | 401 | Invalid credentials |

**JWT Payload:** `{ "id": uint, "role": string }` — expires in 24 hours.

---

### Parking Zones

#### GET /api/v1/zones

Public. List all parking zones with real-time available spots.

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "Parking zones retrieved successfully",
  "data": [
    {
      "id": 5,
      "name": "Terminal 1 EV Charging",
      "type": "ev_charging",
      "total_capacity": 20,
      "available_spots": 14,
      "price_per_hour": 5.5,
      "created_at": "2026-06-20T10:30:00Z"
    }
  ]
}
```

---

#### GET /api/v1/zones/:id

Public. Get a single parking zone.

**Response — 200 OK:** Same shape as list item.

| Error | Status | Message |
|-------|--------|---------|
| Zone not found | 404 | Resource not found |

---

#### POST /api/v1/zones

**Admin only.** Requires `Authorization: Bearer <token>` header.

**Request:**
```json
{
  "name": "Terminal 1 EV Charging",
  "type": "ev_charging",
  "total_capacity": 20,
  "price_per_hour": 5.5
}
```

**Response — 201 Created:**
```json
{
  "success": true,
  "message": "Parking zone created successfully",
  "data": {
    "id": 5,
    "name": "Terminal 1 EV Charging",
    "type": "ev_charging",
    "total_capacity": 20,
    "available_spots": 20,
    "price_per_hour": 5.5,
    "created_at": "2026-06-20T10:30:00Z",
    "updated_at": "2026-06-20T10:30:00Z"
  }
}
```

| Error | Status | Message |
|-------|--------|---------|
| No/invalid token | 401 | Invalid or expired token |
| Not admin | 403 | Forbidden: insufficient permissions |
| Validation failure | 400 | Validation failed |

---

#### PUT /api/v1/zones/:id

**Admin only.** Partial update — only send fields to change.

**Request:** All fields optional.
```json
{
  "name": "Updated Name",
  "price_per_hour": 6.0
}
```

**Response — 200 OK:** Same shape as Create.

---

#### DELETE /api/v1/zones/:id

**Admin only.**

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "Parking zone deleted successfully"
}
```

---

### Reservations

All reservation endpoints require `Authorization: Bearer <token>` header.

#### POST /api/v1/reservations

Authenticated (driver or admin). ⚠️ **Concurrency-safe** — uses `FOR UPDATE` row lock.

**Request:**
```json
{
  "zone_id": 5,
  "license_plate": "ABC-1234"
}
```

**Response — 201 Created:**
```json
{
  "success": true,
  "message": "Reservation confirmed successfully",
  "data": {
    "id": 105,
    "user_id": 1,
    "zone_id": 5,
    "license_plate": "ABC-1234",
    "status": "active",
    "created_at": "2026-06-20T15:30:00Z",
    "updated_at": "2026-06-20T15:30:00Z"
  }
}
```

| Error | Status | Message |
|-------|--------|---------|
| No/invalid token | 401 | Invalid or expired token |
| Zone not found | 404 | Resource not found |
| Zone is full | 409 | Zone is at full capacity |
| Validation failure | 400 | Validation failed |

---

#### GET /api/v1/reservations/my-reservations

Authenticated. Get the current user's reservations.

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "My reservations retrieved successfully",
  "data": [
    {
      "id": 105,
      "license_plate": "ABC-1234",
      "status": "active",
      "zone": {
        "id": 5,
        "name": "Terminal 1 EV Charging",
        "type": "ev_charging"
      },
      "created_at": "2026-06-20T15:30:00Z"
    }
  ]
}
```

---

#### DELETE /api/v1/reservations/:id

Authenticated. Cancel your own reservation.

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "Reservation cancelled successfully"
}
```

| Error | Status | Message |
|-------|--------|---------|
| No/invalid token | 401 | Invalid or expired token |
| Not found | 404 | Resource not found |
| Not your reservation | 403 | Forbidden |

---

#### GET /api/v1/reservations

**Admin only.** List all reservations with user and zone details.

**Response — 200 OK:**
```json
{
  "success": true,
  "message": "All reservations retrieved successfully",
  "data": [
    {
      "id": 105,
      "license_plate": "ABC-1234",
      "status": "active",
      "user": {
        "id": 1,
        "name": "John Doe",
        "email": "john.doe@spotsync.com"
      },
      "zone": {
        "id": 5,
        "name": "Terminal 1 EV Charging",
        "type": "ev_charging"
      },
      "created_at": "2026-06-20T15:30:00Z"
    }
  ]
}
```

---

### Endpoint Summary

| Method | Path | Auth | Role |
|--------|------|------|------|
| POST | /api/v1/auth/register | Public | — |
| POST | /api/v1/auth/login | Public | — |
| GET | /api/v1/zones | Public | — |
| GET | /api/v1/zones/:id | Public | — |
| POST | /api/v1/zones | JWT | admin |
| PUT | /api/v1/zones/:id | JWT | admin |
| DELETE | /api/v1/zones/:id | JWT | admin |
| POST | /api/v1/reservations | JWT | — |
| GET | /api/v1/reservations/my-reservations | JWT | — |
| DELETE | /api/v1/reservations/:id | JWT | — |
| GET | /api/v1/reservations | JWT | admin |

---

## Deployment

### 1. PostgreSQL (NeonDB / Supabase / Aiven)

Provision a PostgreSQL instance and copy the connection string to `DATABASE_URL`.

### 2. Deploy to Render / Railway / Fly.io

```bash
# Set the following environment variables on your platform:
#   DATABASE_URL  — PostgreSQL connection string
#   JWT_SECRET    — A strong random secret
#   PORT          — 8080 (or your platform's assigned port)
```

Build command: `go build -o server ./main.go`  
Start command: `./server`

### 3. Verify

```bash
curl https://your-app.onrender.com/api/v1/zones
```

All 9 unique endpoints should return valid responses.

---

## CORS

The API allows all origins (`*`) with `GET`, `POST`, `PUT`, `DELETE` methods and `Content-Type`, `Authorization` headers. For production, restrict `AllowOrigins` to your frontend domain in `main.go`:

```go
AllowOrigins: []string{"https://your-frontend.com"},
```

---

## Project Structure

```
├── dto/                 # Request/response DTOs with validation tags
├── handler/             # Echo handlers — bind, validate, respond
├── middleware/          # JWT verification, role-based access
├── models/              # GORM model definitions
├── repository/          # Database operations
├── service/             # Business logic
├── main.go              # Entry point, DI wiring, route registration
├── go.mod / go.sum      # Go module dependencies
└── .env.example         # Environment variable template
```
