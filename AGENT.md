# AGENTS.md

This file is read by any AI agent working on SpotSync before doing anything else. It describes the project, the context files, and the rules for working on this codebase.

---

## What Is This Project

SpotSync is a pure backend REST API written in Go. It is a Smart Parking and EV Charging Reservation platform. There is no frontend, no React, no TypeScript. The deliverable is a JSON API only.

**Tech stack:** Go 1.22+, Echo, GORM, PostgreSQL, JWT, bcrypt, go-playground/validator

---

## Read These Files First — In This Order

Before writing any code, read these context files in order:

| File                          | What It Contains                                                                         |
| ----------------------------- | ---------------------------------------------------------------------------------------- |
| `context/project-overview.md` | What SpotSync does, all endpoints, business rules, success criteria                      |
| `context/architecture.md`     | Folder structure, layer rules, database schema, concurrency pattern, invariants          |
| `context/build-plan.md`       | The exact build sequence — what to build, in what order, with what verification steps    |
| `context/code-standards.md`   | Go conventions, naming, layer templates, DI wiring, error handling, commit rules         |
| `context/library-docs.md`     | How to use Echo, GORM, JWT, bcrypt, validator, godotenv in this specific project         |
| `context/api-reference.md`    | Every endpoint's exact request shape, response shape, status codes, and validation rules |
| `context/progress-tracker.md` | What is done, what is in progress, what is next — update after every feature             |

---

## Before Writing Any Code

1. Read `progress-tracker.md` — know exactly where the build is
2. Read `build-plan.md` — understand the feature you are about to implement
3. Read `architecture.md` — confirm which files belong to which layer
4. Read `code-standards.md` — follow the layer templates exactly
5. Read `library-docs.md` — use the exact patterns defined there for GORM, JWT, Echo, etc.
6. Read `api-reference.md` — match the response shape exactly before writing handler code

---

## The Most Important Rules

### Clean Architecture — Non-Negotiable

```
Handler → Service → Repository → Database
```

- Handlers bind requests, call services, return JSON. Nothing else.
- Services contain business logic. They never import Echo or GORM directly.
- Repositories contain all GORM operations. They return domain types or sentinel errors.
- DTOs are the API contract. Models are never returned from handlers.

### The Concurrency Rule — Non-Negotiable

The `POST /api/v1/reservations` endpoint must use a GORM transaction with `FOR UPDATE` row lock on the parking zone. This is the central technical requirement of the assignment. See `architecture.md` and `library-docs.md` for the exact implementation pattern.

### Response Shape — Non-Negotiable

Every response must use:

```json
{ "success": true/false, "message": "...", "data": {} }
```

or for errors:

```json
{ "success": false, "message": "...", "errors": "..." }
```

No deviations. See `api-reference.md` for every endpoint's exact shape.

### Passwords — Non-Negotiable

The `password` field on the User model always has `json:"-"`. It never appears in any response, any log, or any DTO.

---

## What This Project Is NOT

- Not a frontend project — no React, no TypeScript, no Next.js, no Tailwind
- Not an AI agent project — no GPT-4o calls, no browser automation, no Adzuna
- Not a real-time project — no WebSockets, no subscriptions
- Not a file storage project — no PDF generation, no file uploads

---

## Dependency Injection Wiring (main.go)

```
NewUserRepository(db) → NewAuthService(userRepo) → NewAuthHandler(authSvc)
NewZoneRepository(db) → NewZoneService(zoneRepo) → NewZoneHandler(zoneSvc)
NewReservationRepository(db) → NewReservationService(reservationRepo, zoneRepo) → NewReservationHandler(reservationSvc)
```

---

## Sentinel Errors

Defined in the `service` package. Handlers map these to HTTP codes using `handleServiceError`.

| Sentinel            | HTTP Code | Scenario                                |
| ------------------- | --------- | --------------------------------------- |
| `ErrZoneFull`       | 409       | Reservation rejected — zone at capacity |
| `ErrNotFound`       | 404       | Resource does not exist                 |
| `ErrForbidden`      | 403       | Valid user but wrong ownership          |
| `ErrUnauthorized`   | 401       | Invalid or missing credentials          |
| `ErrDuplicateEmail` | 400       | Email already registered                |

---

## Skills Installed

All skills live in `skills/`. Use them — do not skip them because a task feels small.

| Skill               | When to run it                                                                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/architect`        | Before building any new feature — think through layer ownership, DTO shape, middleware chain, and transaction needs before writing a single line |
| `/contract`         | After building any handler — verify the response shape matches `context/api-reference.md` exactly and mark the endpoint as verified              |
| `/review`           | After finishing any feature — verify it against the plan, the Clean Architecture rules, and the API spec                                         |
| `/recover`          | The moment something goes wrong — diagnose the failure mode before re-prompting                                                                  |
| `/remember save`    | At the end of every session — save what was built, what was decided, and what comes next                                                         |
| `/remember restore` | At the start of every session — restore context before touching any code                                                                         |

There is no `/imprint` here — this project has no UI. `/contract` is the backend equivalent.

---

## After Every Feature

1. Run `/contract` — verify the handler response shape matches the spec
2. Run `/review` — verify Clean Architecture rules, error handling, and edge cases
3. Update `context/progress-tracker.md` — check off the completed item, update "Last completed" and "Next", note any decisions or deviations
4. Run `/remember save` at the end of the session
