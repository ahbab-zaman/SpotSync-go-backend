---
name: recover
description: When something goes wrong during a build, diagnose what type of failure it is before deciding how to respond. Targeted fix, hard reset, or full rethink — the right response depends on the right diagnosis.
---

Not every problem is a bug. Not every bug needs debugging.

When something goes wrong with AI-assisted development, the instinct is to keep prompting — describe the problem, ask for a fix, get another broken version, describe that problem, ask for another fix. The session gets longer. The context gets polluted. The code gets worse.

The problem is not the code. The problem is not knowing what type of failure you are dealing with.

This skill diagnoses the failure first. Then it prescribes the right response. Those are two separate steps and they cannot be swapped.

---

## Step 1 — Describe What Went Wrong

Ask the developer:

```
Describe what is wrong. Be specific:
- What did you expect to happen?
- What happened instead?
- What is the exact error message or wrong behaviour?
- How many times have you tried to fix it already?
```

Read the answer carefully. The number of fix attempts matters — it tells you whether this is a fresh problem or a session that has already gone wrong.

---

## Step 2 — Identify the Failure Mode

Based on the description, determine which of three failure modes this is.

### Failure Mode 1 — A specific thing is broken

**Signs:**

- The problem is isolated — one function, one route, one layer
- The rest of the project compiles and behaves correctly
- This is the first or second attempt at fixing it
- The error message is specific — a compile error, a GORM error, a wrong HTTP status code, a test failure

**Common SpotSync examples:**

- GORM `clause.Locking` import missing — `undefined: clause`
- `c.Get("userID")` type assertion panic — missing type cast to `uint`
- Route registered without middleware — returns 200 when it should return 401
- `AutoMigrate` not including a new model — table missing in DB
- Validator tag wrong — `oneof=general ev_charging covered` vs `oneof=general ev_charging covered` (space-separated, not comma)

**Response:** Targeted fix — go to Step 3A.

---

### Failure Mode 2 — The session has gone wrong

**Signs:**

- Multiple fix attempts have made things worse or introduced new problems
- The code has become tangled — fixes patching fixes
- Context in this session is full of failed attempts and contradictions
- It is no longer clear what the original problem was
- The handler, service, and repository are starting to share responsibilities they should not

**Response:** Hard reset — go to Step 3B.

---

### Failure Mode 3 — The foundation is wrong

**Signs:**

- The code compiles but produces fundamentally wrong behaviour
- A core architectural rule has been violated — for example, GORM calls inside a handler, business logic inside a repository, JWT signing inside a handler instead of the service
- The concurrency pattern is wrong — reservation created outside a transaction, or `FOR UPDATE` lock missing
- The response shape does not match `context/api-reference.md` in a structural way — not a typo, a fundamental misread of the spec
- Fixing individual pieces will not help because the approach itself is incorrect

**Response:** Rethink — go to Step 3C.

---

Tell the developer which failure mode this is before proceeding:

```
This looks like Failure Mode [1/2/3] — [name].

[One sentence explaining why you identified it this way.]

Here is how we handle this:
```

---

## Step 3A — Targeted Fix

For Failure Mode 1.

### Diagnose before touching code

Ask for:

- The exact compiler error, runtime panic, or wrong API response
- The specific file and function where the problem occurs
- What the code is supposed to do versus what it actually does

Read only the relevant file — the handler function, the repository method, or the middleware. Do not read the entire codebase.

### Find the root cause

Identify the root cause before suggesting any fix. State it clearly:

```
Root cause: [specific explanation of why this is happening]

This is different from the symptom because: [explanation]
```

### Suggest a precise fix

```
Fix: [what needs to change and why]

This will resolve the root cause because: [explanation]
```

Wait for the developer to confirm before making any changes.

### If the fix does not work

Stop. Do not suggest another fix immediately. Re-examine the root cause diagnosis — if the fix did not work, the diagnosis was probably wrong. Diagnose again from the beginning.

If two diagnoses have both been wrong — re-evaluate whether this is actually Failure Mode 2 or 3.

---

## Step 3B — Hard Reset

For Failure Mode 2.

### Acknowledge the situation honestly

```
This session has gone too far in the wrong direction
to recover by patching. The right move is a clean start.

This is not a failure — it is the correct response
to a polluted context. A fresh session with clear intent
will be faster than continuing here.
```

### Save what is worth keeping

Extract anything valuable from the current state:

```
## Reset Note — [Feature Name]

### What we were building
[Original feature description — which phase and feature number from build-plan.md]

### What went wrong
[Honest summary of how the session went off track]

### What to avoid next time
[Specific approaches or patterns that did not work]

### Clean starting point
[Exactly where to begin fresh — which file, which function, what the first line of correct code should be]
```

### Instruct the developer

```
Next steps:

1. Save this reset note
2. End this session completely
3. Start a fresh session
4. Run /remember restore
5. Approach [feature name] again with the reset note as context

Do not continue in this session.
```

---

## Step 3C — Rethink

For Failure Mode 3.

### Name the wrong assumption

```
The core issue is not a bug — it is a wrong assumption:

Assumed: [what was assumed]
Reality: [what is actually true]

This means the current implementation cannot be fixed
by patching. The approach needs to change.
```

### Common SpotSync rethink triggers

- **Handler calling GORM directly** — the handler must call a service method; the service calls the repository; the repository calls GORM. Rebuild the repository method and service method first, then rewrite the handler.
- **Transaction missing on reservation** — `CreateWithLock` must use `db.Transaction` with `clause.Locking{Strength: "UPDATE"}`. A reservation created outside a transaction cannot be made safe by patching — the repository method must be rewritten.
- **Response shape built from model struct** — if the handler is marshalling a `models.Reservation` directly, the entire response layer needs to be rebuilt using `dto.ReservationResponse`. Models must never be returned from handlers.
- **Middleware applied globally instead of per-group** — if `jwt_middleware` was added to the Echo instance instead of route groups, the public endpoints (`GET /zones`, `POST /auth/register`, etc.) will break. The route registration in `main.go` needs to be rewritten with proper groups.

### Propose the correct approach

```
Correct approach: [description based on context/architecture.md]

Key difference from current approach: [explanation]

What needs to be discarded: [specific files or functions to delete or rewrite]
What can be kept: [what is still valid]
```

Do not start rebuilding immediately. Present the analysis and wait for confirmation.

```
Does this diagnosis match your understanding?

If yes — we can start fresh with the correct approach.
If no — tell me what I am getting wrong.
```

Only after the developer confirms does any rebuilding begin.

---

## The Principle

The worst thing you can do when something is broken is keep doing the same thing faster.

Diagnose first. Respond correctly. Different failures need different responses — and knowing which failure you are dealing with is more than half the solution.
