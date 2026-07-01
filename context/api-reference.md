# API Reference

Living document. Updated after every endpoint is implemented. Read this before building any new handler — match existing response shapes exactly before writing new ones.

---

## How to Use

Before building any handler:

1. Check if the endpoint shape is already defined here
2. Match the exact request body fields, response fields, and HTTP status codes
3. After implementing an endpoint — mark it as done and note any deviations

---

## Standard Response Envelope

Every response — success or error — uses this envelope. No exceptions.

**Success:**

```json
{
  "success": true,
  "message": "Human readable description",
  "data": {}
}
```

**Error:**

```json
{
  "success": false,
  "message": "Human readable description",
  "errors": "Error details or null"
}
```

---

## Auth Module

### POST /api/v1/auth/register

**Status:** ✅ Implemented and verified  
**Access:** Public  
**Middleware:** None

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john.doe@spotsync.com",
  "password": "securePassword123",
  "role": "driver"
}
```

**Validation Rules:**

- `name` — required
- `email` — required, valid email format
- `password` — required, min 8 characters
- `role` — required, must be `driver` or `admin`

**Success Response — 201 Created:**

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

**Error Cases:**
| Scenario | Status | Message |
| ------------------- | ------ | -------------------------- |
| Validation failure | 400 | Validation failed |
| Email already taken | 400 | Email already registered |
| Server error | 500 | Internal server error |

---

### POST /api/v1/auth/login

**Status:** ✅ Implemented and verified  
**Access:** Public  
**Middleware:** None

**Request Body:**

```json
{
  "email": "john.doe@spotsync.com",
  "password": "securePassword123"
}
```

**Success Response — 200 OK:**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john.doe@spotsync.com",
      "role": "driver"
    }
  }
}
```

**JWT Payload must contain:** `id` (uint), `role` (string)

**Status:** ⚠️ Shape note — Login response `user` object includes `created_at` and `updated_at` (reuses `UserResponse` DTO). Spec omits these fields. Decision recorded in progress-tracker.md. Adjust if frontend expects exact spec.

**Error Cases:**
| Scenario | Status | Message |
| --------------------- | ------ | -------------------------- |
| Validation failure | 400 | Validation failed |
| Wrong email/password | 401 | Invalid credentials |
| Server error | 500 | Internal server error |

---

## Parking Zones Module

### GET /api/v1/zones

**Status:** ✅ Implemented — shape note (see below)  
**Access:** Public  
**Middleware:** None

**Success Response — 200 OK:**

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

**Note:** `available_spots` = `total_capacity` − COUNT of `active` reservations for that zone. Always computed dynamically, never stored.

**Shape note:** Response includes `updated_at` field (with `omitempty`) not shown in spec. Extra field, non-breaking.

---

### GET /api/v1/zones/:id

**Status:** ✅ Implemented — shape note (see GET /zones)  
**Access:** Public  
**Middleware:** None

**Success Response — 200 OK:**

```json
{
  "success": true,
  "message": "Parking zone retrieved successfully",
  "data": {
    "id": 5,
    "name": "Terminal 1 EV Charging",
    "type": "ev_charging",
    "total_capacity": 20,
    "available_spots": 14,
    "price_per_hour": 5.5,
    "created_at": "2026-06-20T10:30:00Z"
  }
}
```

**Shape note:** Same `updated_at` deviation as GET /zones.

**Error Cases:**
| Scenario | Status | Message |
| ------------- | ------ | ------------------- |
| Zone not found | 404 | Resource not found |

---

### POST /api/v1/zones

**Status:** ✅ Implemented and verified  
**Access:** Admin only  
**Middleware:** `jwt_middleware` → `role_middleware("admin")`

**Request Body:**

```json
{
  "name": "Terminal 1 EV Charging",
  "type": "ev_charging",
  "total_capacity": 20,
  "price_per_hour": 5.5
}
```

**Validation Rules:**

- `name` — required
- `type` — required, must be `general`, `ev_charging`, or `covered`
- `total_capacity` — required, integer, greater than 0
- `price_per_hour` — required, float, greater than 0

**Success Response — 201 Created:**

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

**Error Cases:**
| Scenario | Status | Message |
| ------------------ | ------ | ----------------------------- |
| No/invalid token | 401 | Invalid or expired token |
| Role is not admin | 403 | Forbidden: insufficient permissions |
| Validation failure | 400 | Validation failed |

---

### PUT /api/v1/zones/:id

**Status:** ✅ Implemented — shape note (same as POST)  
**Access:** Admin only  
**Middleware:** `jwt_middleware` → `role_middleware("admin")`

**Request Body:** Same fields as Create — all optional for partial update.

**Success Response — 200 OK:** Same shape as Create response.

---

### DELETE /api/v1/zones/:id

**Status:** ✅ Implemented — shape note (see below)  
**Access:** Admin only  
**Middleware:** `jwt_middleware` → `role_middleware("admin")`

**Success Response — 200 OK:**

```json
{
  "success": true,
  "message": "Parking zone deleted successfully"
}
```

**Shape note:** Response includes `"data": null` in addition to the fields shown above (reuses standard `successResponse` envelope). Non-breaking extra field.

---

## Reservations Module

### POST /api/v1/reservations

**Status:** ✅ Implemented and verified  
**Access:** Authenticated (driver or admin)  
**Middleware:** `jwt_middleware`

⚠️ **Concurrency-critical.** Must use `FOR UPDATE` row lock inside a GORM transaction. See `CreateWithLock` in `repository/reservation_repository.go`.

**Request Body:**

```json
{
  "zone_id": 5,
  "license_plate": "ABC-1234"
}
```

**Validation Rules:**

- `zone_id` — required, greater than 0
- `license_plate` — required, max 15 characters

**Success Response — 201 Created:**

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

**Error Cases:**
| Scenario | Status | Message |
| ------------------ | ------ | ------------------------------- |
| No/invalid token | 401 | Invalid or expired token |
| Zone not found | 404 | Resource not found |
| Zone is full | 409 | Zone is at full capacity |
| Validation failure | 400 | Validation failed |

---

### GET /api/v1/reservations/my-reservations

**Status:** ✅ Implemented and verified  
**Access:** Authenticated  
**Middleware:** `jwt_middleware`

**Success Response — 200 OK:**

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

**Note:** Zone is loaded via GORM `Preload("Zone")`.

---

### DELETE /api/v1/reservations/:id

**Status:** ✅ Implemented — shape note (see below)  
**Access:** Authenticated (own reservations only)  
**Middleware:** `jwt_middleware`

**Success Response — 200 OK:**

```json
{
  "success": true,
  "message": "Reservation cancelled successfully"
}
```

**Shape note:** Response includes `"data": null` in addition to the fields shown above (reuses standard `successResponse` envelope). Non-breaking extra field.

**Error Cases:**
| Scenario | Status | Message |
| -------------------------------- | ------ | ------------------------------- |
| No/invalid token | 401 | Invalid or expired token |
| Reservation not found | 404 | Resource not found |
| Trying to cancel someone else's | 403 | Forbidden |

---

### GET /api/v1/reservations

**Status:** ✅ Implemented and verified  
**Access:** Admin only  
**Middleware:** `jwt_middleware` → `role_middleware("admin")`

**Success Response — 200 OK:**

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

**Note:** User and Zone loaded via GORM `Preload("User").Preload("Zone")`.

---

## Implementation Checklist

After implementing each endpoint, mark it done and confirm the response shape matches exactly.

| Endpoint                                 | Done | Shape Verified |
| ---------------------------------------- | ---- | -------------- |
| POST /api/v1/auth/register               | [x]  | [x]            |
| POST /api/v1/auth/login                  | [x]  | [x]            |
| GET /api/v1/zones                        | [x]  | [x]  |
| GET /api/v1/zones/:id                    | [x]  | [x]  |
| POST /api/v1/zones                       | [x]  | [x]  |
| PUT /api/v1/zones/:id                    | [x]  | [x]  |
| DELETE /api/v1/zones/:id                 | [x]  | [x]  |
| POST /api/v1/reservations                | [x]  | [x]  |
| GET /api/v1/reservations/my-reservations | [x]  | [x]  |
| DELETE /api/v1/reservations/:id          | [x]  | [x]  |
| GET /api/v1/reservations                 | [x]  | [x]  |
