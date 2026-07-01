# Project Overview

## About the Project

SpotSync is a fully backend REST API built in Go. It is a Smart Parking and EV Charging Reservation platform designed for high-demand locations such as airports and malls. The system manages parking zones and handles the reservation of limited EV charging spots, with a critical focus on concurrency safety to prevent overbooking.

There is no frontend. The deliverable is a clean, well-structured JSON API that follows strict Clean Architecture principles.

---

## The Problem It Solves

Busy airports and malls have limited EV charging spots. Without a centralized reservation system, multiple drivers can attempt to claim the last available spot simultaneously, causing overbooking. SpotSync solves this with a transactional, row-locked reservation flow that guarantees no zone ever exceeds its capacity — even under concurrent load.

---

## API Base Path

All endpoints are prefixed with `/api/v1`.

---

## Roles

| Role   | Permissions                                                                                       |
| ------ | ------------------------------------------------------------------------------------------------- |
| driver | Register, login, view all zones and availability, reserve a spot, view and cancel own reservations |
| admin  | All driver permissions plus: create/update/delete zones, set pricing, view all reservations        |

---

## Endpoints

```
POST   /api/v1/auth/register            → Public. Register a new user
POST   /api/v1/auth/login               → Public. Login and receive JWT

GET    /api/v1/zones                    → Public. List all parking zones with available_spots
GET    /api/v1/zones/:id                → Public. Get single parking zone with available_spots
POST   /api/v1/zones                    → Admin only. Create a new parking zone
PUT    /api/v1/zones/:id                → Admin only. Update a parking zone
DELETE /api/v1/zones/:id                → Admin only. Delete a parking zone

POST   /api/v1/reservations             → Authenticated. Reserve a parking spot (concurrency-critical)
GET    /api/v1/reservations/my-reservations → Authenticated. Get current user's reservations
DELETE /api/v1/reservations/:id         → Authenticated. Cancel own reservation
GET    /api/v1/reservations             → Admin only. Get all reservations in the system
```

---

## Core Business Rules

- A parking zone must never be over its `total_capacity`. The 21st reservation for a 20-capacity zone must be rejected with HTTP 409.
- The reservation endpoint uses a GORM database transaction combined with a `FOR UPDATE` row-level lock on the zone record. This is non-negotiable.
- `available_spots` is always calculated dynamically: `total_capacity` minus the count of `active` reservations for that zone.
- A driver can only cancel their own reservation. Cancelling another user's reservation returns HTTP 403.
- Changing a reservation's status to `cancelled` frees up the spot.
- Passwords are never exposed in any response or log.
- Protected endpoints reject requests without a valid JWT with HTTP 401.
- Role violations return HTTP 403.

---

## Database Tables

```
users            → id, name, email, password, role, created_at, updated_at
parking_zones    → id, name, type, total_capacity, price_per_hour, created_at, updated_at
reservations     → id, user_id, zone_id, license_plate, status, created_at, updated_at
```

---

## Standard Response Shape

**Success:**
```json
{
  "success": true,
  "message": "Operation description",
  "data": {}
}
```

**Error:**
```json
{
  "success": false,
  "message": "Error description",
  "errors": "Error details"
}
```

---

## HTTP Status Codes

| Code | Usage                                                          |
| ---- | -------------------------------------------------------------- |
| 200  | Successful GET, PUT, DELETE                                    |
| 201  | Successful POST (resource created)                             |
| 400  | Validation errors, invalid input, duplicate resource           |
| 401  | Missing, expired, or invalid JWT token                         |
| 403  | Valid token but insufficient role or wrong ownership           |
| 404  | Requested resource does not exist                              |
| 409  | Business logic conflict — zone is full, duplicate license plate |
| 500  | Unexpected server or database error                            |

---

## Target Audience

The assessment panel and technical reviewers evaluating clean Go architecture, correct JWT implementation, proper bcrypt usage, GORM transaction handling, and concurrency safety under the "EV Spot Bottleneck" problem.

---

## Success Criteria

- All 9 endpoints return exactly the specified request/response shapes
- Zone capacity is never exceeded, even under concurrent load
- `FOR UPDATE` row lock is correctly implemented in the reservation transaction
- Clean Architecture layers are strictly separated — handlers never touch the DB
- JWT payload includes `id` and `role`
- Passwords never appear in any response
- All validation errors return structured, human-readable messages
- Deployment is live and publicly accessible