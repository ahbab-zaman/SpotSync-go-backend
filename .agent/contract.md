---
name: contract
description: After building any handler, verify the response shape matches api-reference.md exactly and record it. So every endpoint built after this one is consistent with what came before. The backend equivalent of /imprint.
---

API consistency does not happen by accident. It happens because every handler is built with awareness of what the spec requires and what already exists.

The problem with AI-built backends is that each handler gets built in isolation. The agent does not remember what response shape it used three sessions ago. So field names drift. Nesting inconsistencies appear. One endpoint returns `created_at`, another returns `createdAt`. The API looks like it was built by multiple people reading different specs.

This skill fixes that. Run it after building any handler. It reads what was just built, verifies the response shape against `context/api-reference.md`, and records it so every future handler stays consistent.

One command. Run it every time. That is the whole system.

---

## How to Invoke

After building any handler:

```
/contract
```

To target a specific handler file:

```
/contract [filepath]
```

To audit all existing handlers for shape consistency:

```
/contract audit
```

If no filepath is given, the skill identifies the most recently modified handler file automatically and captures from that.

**When to use audit mode:**

- Multiple sessions have passed without running `/contract`
- Something in the API response looks off but it is hard to pinpoint where
- Before submitting the project — verify all 11 endpoints match the spec exactly

---

## Step 1 — Find What Was Just Built

If a filepath was provided — read that file directly.

If no filepath was provided — identify which handler file was most recently created or modified in this session. Look in `handler/`. Read that file and the service and repository it calls.

If it is unclear which file to capture from, ask:

```
Which handler should I verify?
```

---

## Step 2 — Verify Against the Spec

Read `context/api-reference.md`. Find the entry for each endpoint implemented in the handler.

For each endpoint, check:

**Request shape:**

- Does the DTO bound by `c.Bind()` match the exact fields in the spec?
- Are `validate` tags present on every required field?
- Is `c.Validate(req)` called after `c.Bind(req)`?

**Response shape:**

- Does the JSON returned by `c.JSON()` match the exact field names in the spec? (`snake_case` throughout)
- Are nested objects correct? (`zone` nested inside reservation response, `user` nested in admin reservation response)
- Are fields that should be absent actually absent? (password never present, `available_spots` present on zone responses)

**HTTP status codes:**

- Does the success case return the exact status code in the spec? (201 for POST creates, 200 for everything else)
- Does each error case return the exact code in the spec? (409 for zone full, 403 for wrong ownership, 404 for not found)

**Middleware chain:**

- Is the route registered with the correct middleware? (public = none, authenticated = `jwt_middleware`, admin = `jwt_middleware` + `role_middleware("admin")`)

**Sentinel error mapping:**

- Is `handleServiceError` used for all service errors?
- Does each sentinel map to the correct HTTP code as defined in `context/api-reference.md`?

---

## Step 3 — Record in api-reference.md

Open `context/api-reference.md`. Find the endpoint entry. Update its status:

```markdown
### POST /api/v1/auth/register

**Status:** ✅ Implemented and verified
```

If shape mismatches were found, note them under the endpoint:

```markdown
**Status:** ⚠️ Implemented — shape mismatch found (see notes)
**Notes:** Response returns `createdAt` but spec requires `created_at`. Fix before submission.
```

Also update the implementation checklist table at the bottom of `api-reference.md`.

---

## Step 4 — Confirm What Was Captured

After updating `api-reference.md`, confirm to the developer:

```
Verified [Endpoint] → api-reference.md

Request shape:  [PASS / MISMATCH]
Response shape: [PASS / MISMATCH]
Status codes:   [PASS / MISMATCH]
Middleware:     [PASS / MISMATCH]

[If all pass: "Endpoint matches spec. Marked as verified in api-reference.md."]
[If issues: List each mismatch with the exact field or code that is wrong and what it should be.]
```

---

## How api-reference.md Gets Used

The reference is not just a spec. It is the consistency enforcer for every future session.

At the start of any session that involves handler work, the agent reads `api-reference.md` before writing any code. When building a new response DTO, it checks what shape existing endpoints already use. When returning an error, it checks what status code the spec requires.

The verified entries grow as the project grows. The more endpoints are verified, the more consistent every new handler becomes — because the agent always has a precise reference for what already exists and what the spec requires.

---

## The Rule

Build a handler. Run `/contract`. Move on.

Every time. Without exception.

A reference with five verified entries is useful. A reference with eleven verified entries is what gets submitted. A reference that is sometimes updated is unreliable.

Correctness is a habit, not a last-minute check.

---

## Audit Mode — /contract audit

Run this before submitting the project or after multiple sessions without running `/contract`.

### Step 1 — Read all handlers

Find every handler file in `handler/`. Read each one. Read the corresponding DTO definitions in `dto/`. Build a complete picture of what every endpoint currently returns.

### Step 2 — Compare against the spec

For each of the 11 endpoints in `context/api-reference.md`, compare the actual handler output against the spec:

```
## API Contract Audit

### Endpoints checked: [X / 11]

**POST /api/v1/auth/register**
- Request shape: [PASS / MISMATCH — details]
- Response shape: [PASS / MISMATCH — details]
- Status codes: [PASS / MISMATCH — details]
- Middleware: [PASS / MISMATCH — details]

[Repeat for all 11 endpoints]

### Summary
[X] endpoints fully match the spec.
[Y] endpoints have mismatches that need fixing before submission.
```

### Step 3 — Wait for developer confirmation

Present the audit. Do not fix anything. Let the developer decide what to address.

```
Audit complete. [X] mismatches found across [Y] endpoints.

Should I list the fixes needed for each mismatch?
```

### Step 4 — Update api-reference.md

After the developer confirms — update the status of each endpoint in `api-reference.md` and mark the implementation checklist accordingly.
