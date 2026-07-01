# Postman API Testing Guide

All endpoints are prefixed with `http://localhost:8080/api/v1` (or your deployed URL).

---

## 1. Import as Collection (Optional)

Create a new collection in Postman named **SpotSync**. Add the following requests inside it.

---

## 2. Auth Endpoints

### 2.1 Register a Driver

| Setting | Value |
|---------|-------|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/register` |
| Headers | `Content-Type: application/json` |

**Body (raw JSON):**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "role": "driver"
}
```

**Expected Response — 201:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "driver",
    "created_at": "2026-07-02T12:00:00Z",
    "updated_at": "2026-07-02T12:00:00Z"
  }
}
```

---

### 2.2 Register an Admin

| Setting | Value |
|---------|-------|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/register` |
| Headers | `Content-Type: application/json` |

**Body (raw JSON):**
```json
{
  "name": "Admin User",
  "email": "admin@spotsync.com",
  "password": "AdminPass123!",
  "role": "admin"
}
```

**Expected Response — 201**

---

### 2.3 Login (Driver)

| Setting | Value |
|---------|-------|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/login` |
| Headers | `Content-Type: application/json` |

**Body:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Expected Response — 200:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "driver"
    }
  }
}
```

> **Postman Tip:** Copy the `token` value. In Postman, create a collection variable `driver_token` and set it to this value. Then for all authenticated requests, add a header `Authorization: Bearer {{driver_token}}`.

---

### 2.4 Login (Admin)

Same as above with admin credentials (`admin@spotsync.com` / `AdminPass123!`). Save token as `admin_token`.

---

## 3. Parking Zones

### 3.1 List All Zones (Public)

| Setting | Value |
|---------|-------|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/zones` |

**Expected Response — 200:** (empty array if no zones exist)
```json
{
  "success": true,
  "message": "Parking zones retrieved successfully",
  "data": []
}
```

---

### 3.2 Create a Zone (Admin Only)

| Setting | Value |
|---------|-------|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/zones` |
| Headers | `Content-Type: application/json`, `Authorization: Bearer {{admin_token}}` |

**Body:**
```json
{
  "name": "Terminal 1 General",
  "type": "general",
  "total_capacity": 50,
  "price_per_hour": 2.5
}
```

**Expected Response — 201:**
```json
{
  "success": true,
  "message": "Parking zone created successfully",
  "data": {
    "id": 1,
    "name": "Terminal 1 General",
    "type": "general",
    "total_capacity": 50,
    "available_spots": 50,
    "price_per_hour": 2.5,
    "created_at": "2026-07-02T12:00:00Z",
    "updated_at": "2026-07-02T12:00:00Z"
  }
}
```

Create 2 more zones for testing:

```json
{ "name": "Terminal 2 EV", "type": "ev_charging", "total_capacity": 10, "price_per_hour": 5.0 }
{ "name": "Terminal 3 Covered", "type": "covered", "total_capacity": 20, "price_per_hour": 3.0 }
```

---

### 3.3 Get Zone by ID (Public)

| Setting | Value |
|---------|-------|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/zones/1` |

**Expected Response — 200:** Single zone object.

---

### 3.4 Update a Zone (Admin Only)

| Setting | Value |
|---------|-------|
| Method | `PUT` |
| URL | `{{base_url}}/api/v1/zones/1` |
| Headers | `Content-Type: application/json`, `Authorization: Bearer {{admin_token}}` |

**Body (partial update):**
```json
{
  "price_per_hour": 3.0
}
```

**Expected Response — 200:** Updated zone with new `price_per_hour`.

---

### 3.5 Delete a Zone (Admin Only)

| Setting | Value |
|---------|-------|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/zones/1` |
| Headers | `Authorization: Bearer {{admin_token}}` |

**Expected Response — 200:**
```json
{
  "success": true,
  "message": "Parking zone deleted successfully"
}
```

> Re-create zone 1 before proceeding to reservation tests.

---

## 4. Reservations

### 4.1 Create Reservation (Driver)

| Setting | Value |
|---------|-------|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/reservations` |
| Headers | `Content-Type: application/json`, `Authorization: Bearer {{driver_token}}` |

**Body:**
```json
{
  "zone_id": 2,
  "license_plate": "ABC-1234"
}
```

**Expected Response — 201:**
```json
{
  "success": true,
  "message": "Reservation confirmed successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "zone_id": 2,
    "license_plate": "ABC-1234",
    "status": "active",
    "created_at": "2026-07-02T12:00:00Z",
    "updated_at": "2026-07-02T12:00:00Z"
  }
}
```

Create a second reservation (same driver, different zone):
```json
{
  "zone_id": 3,
  "license_plate": "XYZ-9876"
}
```

---

### 4.2 Get My Reservations (Driver)

| Setting | Value |
|---------|-------|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/reservations/my-reservations` |
| Headers | `Authorization: Bearer {{driver_token}}` |

**Expected Response — 200:**
```json
{
  "success": true,
  "message": "My reservations retrieved successfully",
  "data": [
    {
      "id": 1,
      "license_plate": "ABC-1234",
      "status": "active",
      "zone": {
        "id": 2,
        "name": "Terminal 2 EV",
        "type": "ev_charging"
      },
      "created_at": "2026-07-02T12:00:00Z"
    },
    {
      "id": 2,
      "license_plate": "XYZ-9876",
      "status": "active",
      "zone": {
        "id": 3,
        "name": "Terminal 3 Covered",
        "type": "covered"
      },
      "created_at": "2026-07-02T12:00:00Z"
    }
  ]
}
```

---

### 4.3 Cancel Reservation (Driver — Own Reservation)

| Setting | Value |
|---------|-------|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/reservations/2` |
| Headers | `Authorization: Bearer {{driver_token}}` |

**Expected Response — 200:**
```json
{
  "success": true,
  "message": "Reservation cancelled successfully"
}
```

---

### 4.4 Cancel Someone Else's Reservation (Negative Test)

| Setting | Value |
|---------|-------|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/reservations/1` |
| Headers | `Authorization: Bearer {{driver_token}}` |

> If reservation 1 belongs to a different user, expected **403:**
```json
{
  "success": false,
  "message": "Forbidden",
  "errors": null
}
```

---

### 4.5 Get All Reservations (Admin Only)

| Setting | Value |
|---------|-------|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/reservations` |
| Headers | `Authorization: Bearer {{admin_token}}` |

**Expected Response — 200:**
```json
{
  "success": true,
  "message": "All reservations retrieved successfully",
  "data": [
    {
      "id": 1,
      "license_plate": "ABC-1234",
      "status": "active",
      "user": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
      },
      "zone": {
        "id": 2,
        "name": "Terminal 2 EV",
        "type": "ev_charging"
      },
      "created_at": "2026-07-02T12:00:00Z"
    }
  ]
}
```

---

## 5. Concurrency Test (Critical)

This tests that two simultaneous requests to book the last spot don't both succeed.

### Setup
1. Find a zone with only 1 available spot (the EV zone with capacity 10, minus existing reservations)
2. Or create a new zone with capacity 1

Create a zone with capacity 1 (admin):
```json
POST {{base_url}}/api/v1/zones
{
  "name": "Single Spot Zone",
  "type": "general",
  "total_capacity": 1,
  "price_per_hour": 10.0
}
```
Note the new zone ID (e.g., 5).

### Test
Send **two** reservation requests **simultaneously** (use Postman's "Send" button on two tabs at the same time):

```json
POST {{base_url}}/api/v1/reservations
{
  "zone_id": 5,
  "license_plate": "CAR-0001"
}
```

```json
POST {{base_url}}/api/v1/reservations
{
  "zone_id": 5,
  "license_plate": "CAR-0002"
}
```

**Expected result:**
- One request → **201 Created** (reservation confirmed)
- The other → **409 Conflict**:
```json
{
  "success": false,
  "message": "Zone is at full capacity",
  "errors": null
}
```

If both return 201, the `FOR UPDATE` lock is not working correctly.

---

## 6. Error Scenarios

| Test | Request | Expected |
|------|---------|----------|
| Register with existing email | Same email as step 2.1 | 400 "Email already registered" |
| Login with wrong password | `password: "wrong"` | 401 "Invalid credentials" |
| Create zone without token | POST /zones (no auth) | 401 "Missing or invalid authorization header" |
| Create zone with driver token | POST /zones (driver_token) | 403 "Forbidden: insufficient permissions" |
| Get non-existent zone | GET /zones/9999 | 404 "Resource not found" |
| Reserve non-existent zone | POST /reservations with `zone_id: 9999` | 404 "Resource not found" |
| Cancel non-existent reservation | DELETE /reservations/9999 | 404 "Resource not found" |
| Invalid zone_id | POST /zones with `total_capacity: 0` | 400 "Validation failed" |

---

## Postman Variables Summary

Create these collection variables for convenience:

| Variable | Value |
|----------|-------|
| `base_url` | `http://localhost:8080` |
| `driver_token` | (from login response) |
| `admin_token` | (from admin login response) |

Then use `{{base_url}}`, `{{driver_token}}`, `{{admin_token}}` in all request URLs and headers.
