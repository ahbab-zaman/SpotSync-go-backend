---
name: review
description: After building a feature, verify it matches what was planned, respects the Clean Architecture boundaries, follows Go and project conventions, and is ready to ship. Reports issues clearly so the developer decides what to fix.
---

Building is not done when the code compiles. It is done when the code is correct.

AI moves fast. Fast means things get built that work on the surface but drift from the architecture, skip validation, or miss the exact response shape the spec requires. This skill catches those things before they compound into bigger problems.

Run this after every feature. Before you move on.

## What This Skill Does Not Do

It does not fix anything. It reports what it finds and lets the developer decide what matters and what to do about it. Fixing without understanding is how problems get buried, not solved.

---

## Step 1 — Understand What Should Have Been Built

Before reviewing anything, establish the benchmark.

Read in this order:

- The implementation plan from `/architect` if one exists
- The feature entry in `context/build-plan.md`
- The endpoint shape in `context/api-reference.md`
- The layer rules in `context/architecture.md`
- The code conventions in `context/code-standards.md`

If no plan exists, ask the developer to describe what the feature was supposed to do before reviewing. You cannot verify correctness without knowing what correct looks like.

---

## Step 2 — Review in Three Layers

### Layer 1 — Does it match the plan and the spec?

Compare what was built against `context/api-reference.md` and the build plan.

Check:

- **Request shape** — do the DTO fields, JSON keys, and validate tags match exactly what `api-reference.md` specifies?
- **Response shape** — does the JSON response match the exact structure in `api-reference.md`? Field names, nesting, data types?
- **HTTP status codes** — does the handler return exactly the specified code for success and for each error case?
- **Scope** — did the implementation stay within bounds or add things that were not asked for?

Flag anything that was planned but missing. Flag anything that was built but not planned.

### Layer 2 — Does it respect Clean Architecture?

This is where drift most commonly happens. The feature works, but it violates the layer rules that the project depends on.

Check:

- **Handler** — does it only bind, validate, call the service, and return JSON? No GORM, no bcrypt, no JWT signing here.
- **Service** — does it only contain business logic? No Echo imports, no `c echo.Context` parameters, no GORM calls. Does it call the repository through the interface?
- **Repository** — does it contain all and only GORM operations? No HTTP status codes, no DTO types, no business rules.
- **DTO** — are request and response types in `dto/` and not mixed with model types? Is `json:"-"` on the password field in the User model?
- **Models** — are GORM structs clean of HTTP or business logic?
- **Dependency injection** — are all layers wired in `main.go`? Is the repository passed to the service, and the service passed to the handler?
- **Sentinel errors** — are domain errors defined in the `service` package and mapped to HTTP codes only in the handler via `handleServiceError`? No raw GORM errors returned from a handler.
- **Middleware chain** — does the route have the correct middleware? Public routes have none. Authenticated routes have `jwt_middleware`. Admin routes have both `jwt_middleware` and `role_middleware("admin")`.

### Layer 3 — Is it production ready?

Check:

- **Validation** — does every request DTO have `validate` tags on every field? Is `c.Validate(req)` called after `c.Bind(req)`?
- **Error handling** — is every `err != nil` check present? Are errors logged with a context prefix before returning? No empty catch equivalents (`_ = err`)?
- **Password safety** — does the User model have `json:"-"` on the password field? Does it appear in any response DTO or log?
- **JWT claims** — does the token payload include both `id` (uint) and `role` (string)?
- **Concurrency** — if this is the reservation endpoint: is `FOR UPDATE` used inside a transaction? Is `ErrZoneFull` returned as HTTP 409?
- **available_spots** — if zones are returned: is `available_spots` computed dynamically, not stored?
- **Edge cases** — what happens when the resource does not exist? When the user is not the owner? When the zone is full?

---

## Step 3 — Report What You Found

After completing all three layers, produce a clear report. Do not bury issues. Do not soften them. Report honestly so the developer can make informed decisions.

```
## Review — [Feature Name]

### Layer 1 — Plan and spec alignment
[PASS / ISSUES FOUND]
[List any gaps between what was planned and what was built]
[List any request/response shape mismatches against api-reference.md]
[List any wrong HTTP status codes]

### Layer 2 — Clean Architecture integrity
[PASS / ISSUES FOUND]
[List any layer boundary violations]
[List any missing or wrong middleware on routes]
[List any sentinel error mapping issues]

### Layer 3 — Production readiness
[PASS / ISSUES FOUND]
[List any missing validation tags or skipped c.Validate() calls]
[List any unhandled errors or missing nil checks]
[List any password exposure risks]
[List any concurrency safety issues on the reservation endpoint]

### Summary
[X] issues found across [Y] layers.

[If no issues: "No issues found. This feature is ready to ship."]
[If issues: "Resolve the above before moving to the next feature."]
```

---

## After a Clean Review

If the feature passes all three layers, update `context/progress-tracker.md`:

- Check off the completed feature
- Update "Last completed" and "Next"
- Note any decisions or deviations in the Decisions or Notes sections
