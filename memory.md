# Memory — SpotSync Session 2026-07-01

Last updated: 2026-07-01 18:30

## What was built

- **Feature 03 — Middleware**:
  - `middleware/jwt_middleware.go` — `JWTMiddleware()` factory returning Echo middleware. Reads `Authorization: Bearer <token>`, parses JWT with `JWTClaims` (UserID uint + Role string), verifies with `JWT_SECRET` env var, injects `userID` and `role` into Echo context. Returns 401 on missing/invalid token.
  - `middleware/role_middleware.go` — `RoleMiddleware(requiredRole string)` factory returning Echo middleware. Reads `role` from context, returns 403 if not matching.
  - `middleware/jwt_middleware.go` also exports `JWTClaims` struct for use by future auth service.
  - `main.go` — Added `/api/v1/protected-test` test route protected by `JWTMiddleware()` for manual verification. Echo's built-in middleware aliased as `echomw` to avoid package name collision.

## Decisions made

- Echo's `github.com/labstack/echo/v4/middleware` imported as `echomw` alias to prevent name collision with custom `middleware` package.
- `JWTClaims` placed in `middleware` package since that's where verification happens; auth service will import from here.

## Problems solved

- `go build` failed initially because `github.com/golang-jwt/jwt/v5` was not in go.mod — fixed with `go get`.
- Package name collision between `github.com/labstack/echo/v4/middleware` and local `middleware` package — resolved by aliasing Echo's middleware as `echomw`.

## Current state

- All of Phase 1 (Foundation) is complete: scaffolding, models, database connection, middleware.
- Project compiles cleanly (`go build ./...` passes).
- No running database or server tests yet.

## Next session starts with

**Feature 04 — DTOs for Auth**: `dto/auth_dto.go`
- `RegisterRequest` — name, email, password, role with validator tags (required, email, oneof=driver admin)
- `LoginRequest` — email, password with validator tags
- `UserResponse` — id, name, email, role, created_at, updated_at (no password)
- `LoginResponse` — token string, user UserResponse

## Open questions

- None
