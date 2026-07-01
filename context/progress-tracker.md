# Progress Tracker

Update this file after every completed feature. Any AI agent reading this should immediately know what is done, what is in progress, and what is next.

---

## Current Status

**Phase:** Not started  
**Last completed:** —  
**Next:** 01 Project Scaffolding

---

## Progress

### Phase 1 — Foundation

- [ ] 01 Project Scaffolding — `go mod init`, install all dependencies, create folder structure, `.env`, `.gitignore`
- [ ] 02 Database Connection + Models — User, ParkingZone, Reservation GORM structs, AutoMigrate confirmed
- [ ] 03 Middleware — JWT verification middleware, role-check middleware

### Phase 2 — Auth Module

- [ ] 04 DTOs — Auth — RegisterRequest, LoginRequest, UserResponse, LoginResponse
- [ ] 05 Repository — User — CreateUser, FindByEmail, FindByID
- [ ] 06 Service — Auth — Register (bcrypt hash), Login (hash verify + JWT sign)
- [ ] 07 Handler — Auth — POST /auth/register, POST /auth/login, routes registered

### Phase 3 — Parking Zones Module

- [ ] 08 DTOs — Zones — CreateZoneRequest, UpdateZoneRequest, ZoneResponse (with available_spots)
- [ ] 09 Repository — Zone — FindAll, FindByID, Create, Update, Delete, CountActiveReservations
- [ ] 10 Service — Zone — GetAll (with available_spots), GetByID, Create, Update, Delete
- [ ] 11 Handler — Zone — all 5 zone endpoints, correct middleware chains registered

### Phase 4 — Reservations Module

- [ ] 12 DTOs — Reservations — CreateReservationRequest, ReservationResponse, MyReservationResponse, AdminReservationResponse
- [ ] 13 Repository — Reservation — CreateWithLock (FOR UPDATE transaction), FindByUserID, FindByID, UpdateStatus, FindAll
- [ ] 14 Service — Reservation — Reserve, GetMyReservations, CancelReservation (ownership check), GetAllReservations
- [ ] 15 Handler — Reservation — all 4 reservation endpoints, correct middleware chains registered

### Phase 5 — Deployment

- [ ] 16 Deployment — PostgreSQL provisioned, backend deployed, env vars set, live URL confirmed, README.md written

---

## Decisions Made During Build

_Add decisions here as they are made during implementation._

---

## Notes

_Add notes here as the build progresses — workarounds, patterns, anything that differs from the context files._
