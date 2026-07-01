---
name: remember
description: Save what matters at the end of a session so the next session picks up exactly where you left off. Or restore context at the start of a new session so nothing is lost between them.
---

AI has no memory between sessions. Every new session starts blank. This skill fixes that.

Run it at the end of a session to save. Run it at the start of a new session to restore. Done consistently, nothing is ever lost between sessions.

## Security Boundary

This skill must never persist secrets. If any sensitive value appears in the conversation — database connection strings, JWT secrets, API keys — do not copy it to `memory.md`.

Sensitive data includes (non-exhaustive):

- `DATABASE_URL` or any connection string
- `JWT_SECRET` or any signing key
- Any API key or token
- Passwords or passphrases

If a detail is useful but sensitive, store a redacted placeholder instead — for example `[REDACTED_JWT_SECRET]`.

If unsure whether something is sensitive, treat it as sensitive and omit or redact it.

---

## How to Invoke

**To save at end of session:**

```
/remember save
```

**To restore at start of new session:**

```
/remember restore
```

If the developer just runs `/remember` without specifying — ask them which one they need.

---

## Save Mode

When the developer runs `/remember save`:

### What to capture

Review the current conversation to extract only what a developer would genuinely need to continue this work in a completely fresh context. Not a transcript. Not a summary of everything that happened. The essential state.

Think like someone handing off the project to a Go developer who is equally skilled but knows nothing about what happened today. What would they need to continue without losing anything?

**What was built** — specific files created or modified, features completed. Be precise. Not "built the auth module" — "created `dto/auth_dto.go`, `repository/user_repository.go`, `service/auth_service.go`, `handler/auth_handler.go`. Register and Login endpoints working. Routes registered in `main.go`. Confirmed with Postman."

**Decisions made** — choices that would be hard to reverse or that future work depends on. Not implementation details — architectural choices that shaped the layer structure. For example: "Chose to put `ErrZoneFull` in the `service` package so all sentinel errors live in one place. Handlers import from `service` to do the mapping."

**Problems solved** — any issue that took time to figure out. So the next session does not solve the same problem twice. For example: "GORM `clause.Locking` must be imported from `gorm.io/gorm/clause` not from `gorm.io/gorm`. The import path is not obvious — this caught us."

**Current state** — exactly where things stand. Which features from `context/progress-tracker.md` are done. What compiles. What is known to be broken or incomplete.

**What comes next** — the very next thing that needs to happen, specific enough that the next session can start immediately. For example: "Next: Feature 12 — DTOs for Reservations. Start with `dto/reservation_dto.go`. Fields defined in `context/api-reference.md`."

**Open questions** — anything unresolved that the next session needs to address.

### What not to capture

- Implementation details that are visible in the code
- Decisions already documented in context files
- Anything that can be inferred by reading the codebase
- The process of how something was built — only what was built and what was decided
- Any secrets, connection strings, JWT secrets, or credential-like values

### Safety check before writing

Before writing `memory.md`, run a final pass over the content to ensure no sensitive value is present. If found, remove or redact before writing.

### Where to save

Write the memory to `memory.md` in the project root. This file always contains only the most recent session state.

If `memory.md` already exists, show a brief summary of what is currently saved and ask for confirmation before overwriting:

```
memory.md already exists from a previous session.
Current memory covers: [one-line summary of existing content].

Overwrite with this session's memory? (yes / no)
```

Only write after confirmation. If the developer says no, reply:

```
No changes made. memory.md is unchanged.
```

### Format

```markdown
# Memory — SpotSync Session [date]

Last updated: [date and time]

## What was built

[Specific files created or modified, features completed this session.
Reference feature numbers from context/build-plan.md where possible.]

## Decisions made

[Architectural and implementation decisions that future work depends on.
Not code details — choices that shaped layer structure, error handling, route organisation.]

## Problems solved

[Issues resolved this session — exact error messages, root causes, fixes.
So the next session does not solve the same problem twice.]

## Current state

[Which features from context/progress-tracker.md are done.
What compiles and passes manual testing. What is partial or broken.]

## Next session starts with

[The very first thing to do — feature number, file name, and first step.
Specific enough that work can start immediately without re-reading everything.]

## Open questions

[Anything unresolved that needs addressing in the next session.]
```

After writing the file, confirm:

```
Memory saved to memory.md.

Next session: run /remember restore to pick up from here.
```

---

## Restore Mode

When the developer runs `/remember restore` at the start of a new session:

### Step 1 — Find the memory

Look for `memory.md` in the project root. If it does not exist:

```
No memory.md found in this project.

Either this is the first session, or the file was not saved.
To save memory at the end of a session, run /remember save.
```

### Step 2 — Read context files

Read `memory.md` first. Then read these context files in order:

1. `context/progress-tracker.md` — confirms what is actually done
2. `context/build-plan.md` — confirms what comes next
3. `context/architecture.md` — refreshes layer rules and folder structure
4. `AGENTS.md` — refreshes project rules and invariants

Do not scan or read source code files unless `memory.md` specifically points to one.

When restoring, never surface raw secrets from any source — summarise in redacted form only.

### Step 3 — Confirm what was restored

Do not start building. Summarise what was restored so the developer can verify:

```
Memory restored. Here is where we are:

**Last session:** [what was built — feature numbers and file names]
**Current state:** [what works right now, what is partial]
**Decisions in place:** [key decisions that are locked in]
**Next up:** [feature number, file, and first step]

Is this correct? Say yes to continue, or correct anything
that does not look right before we proceed.
```

Only after the developer confirms does the session continue.

### If memory is incomplete or unclear

```
I found memory.md but some context seems missing —
[what is unclear or absent].

Should we continue with what we have, or do you want
to fill in the gaps before we start?
```

Do not guess. Do not assume. Surface the gap and let the developer decide.

---

## The Rule

Every session ends with `/remember save`.
Every session starts with `/remember restore`.

That is the whole system. Consistent use is what makes it work.
A skill used sometimes is a skill that cannot be relied on.
