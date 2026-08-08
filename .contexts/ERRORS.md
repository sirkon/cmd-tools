# Go Error Handling Conventions

These are the project-wide rules for dealing with errors in Go. Follow them for every
error path. The three operations involved — **annotating**, **logging**, and **returning** —
have distinct responsibilities. Annotating is not opposed to returning: **annotation is a
complement that, in certain cases, must accompany a return** (see rule 4).

**Scoping note.** Rules 3 and 9 are scoped **per function**: they govern what that
function does with an error before handing it off. Rules 2 and 4 describe the journey
across the whole call stack. The terminal boundary (typically the HTTP handler) is the
function that logs the full chain — including every annotation it carries — and that
logging is always permitted there.

## The rules

1. **Never lose an error.** Every error must either be logged, annotated, returned, or a
   combination that satisfies the rest of these rules. An error must never silently vanish.

2. **Log errors that end their lifecycle in scope.** If an error is exhausted within its
   scope, or leaves the responsibility zone of business logic (for example, when we reply
   to a client in a handler), it **must be logged**. At the boundary the error is logged
   and translated into a client-facing response (e.g. an HTTP status code); the response
   is not a Go error return, so rule 3 is not implicated.

3. **Never return an error you logged.** Within a single function, logging an error and
   returning it (in either order) is forbidden — it is duplicative and produces noisy
   double reporting. An error logged in a function ends its journey there.

4. **Annotate "outer" errors when there are multiple error-return sites.** Any function
   that has **two or more `return` statements returning a non-nil error** must annotate
   the "outer" error — the one obtained from the result of a called function — with
   context, via `fmt.Errorf("....: %w", err)`. Count error-return sites only: happy-path
   returns (`return u, nil` and the like) do not count. Always use `%w` (never `%v`) so
   that `errors.Is`/`errors.As` keep working through the chain. **This rule has exactly
   one exception: recursive functions** (rule 8 takes precedence and they return bare).
   In every other multi error-site function, annotation is mandatory.

5. **Annotation text describes the action, not the failure.**
   The annotation should say *what we were doing* when the error arrived, not *that an
   error occurred*.
   - Correct: `fmt.Errorf("open file: %w", err)`
   - Incorrect: `fmt.Errorf("failed to open file: %w", err)`

6. **Annotation texts must be unique within a function.** If a function has several
   error-return sites, each gets its own annotation text. Annotations play two roles, and
   both depend on this uniqueness:
   1. **Traceability** — uniqueness is the key. Reading the chain at the boundary must
      identify exactly one code site for each hop, and grepping the codebase for an
      annotation text must land on exactly one site.
   2. **Researchability** — provided by a *proper* annotation (rule 5): it tells what we
      were doing when the error happened, and it may also attach pieces of the execution
      state as context — identifiers, inputs, keys — so the failure can be researched
      without the source at hand, e.g. `fmt.Errorf("fetch user %d: %w", id, err)`.
   The same text may legitimately recur in different functions; the chain itself carries
   the cross-function path. **Practical pattern** (rules 5 + 6 combined): the annotation
   is an action phrase plus the distinguishing parameters of that site — `fetch user %d`,
   `validate user %d`. The action phrase says *what we were doing*; the parameters both
   pin down *which* site it was (uniqueness, traceability) and record the state needed to
   research the failure (researchability).

7. **Log text states the failure openly.**
   In direct contrast to rule 5, logging should report the error plainly — the actual
   `"failed to open file"`. Why: the log text is the **essence** of the problem, while the
   error text is only the **context** of the problem. In logs, essence comes first,
   followed by context, which is the cognitively correct order:
   `log.Printf("failed to open file: %v", err)`.

8. **Never annotate errors in recursive functions.** Annotating inside recursion is
   forbidden (it would repeat/duplicate context on every frame). This rule takes
   precedence over rule 4: a recursive function with multiple error-return sites returns
   received errors bare, unwrapped.

9. **Never log an error you just annotated.** Within a single function, an error you
   just wrapped must be returned, not logged — logging it there would duplicate the
   annotation you just wrote. Return it and let the caller decide. This does **not**
   apply to terminal boundaries: the handler may (and, per rule 2, must) log the fully
   annotated chain.

## Formal state model

An error's state describes how it was created/transformed within the **current function**
only. `return err` moves it to `Rtrd`; when another function receives that same error via a
call, it is `Rcvd` for that receiver. Allowed state transitions depend on context.
Rule 1 (never lose an error) means the final state of every error must be `Rtrd` or `Lggd`
(or a `Wrpd`-then-`Rtrd` chain).

| Mnemonic | Meaning |
|----------|---------|
| `Crtd` | Created by a constructor or an error-sentinel (e.g. `io.EOF`) |
| `Rcvd` | Received from outside (e.g. from calling `os.Open(fileName)`) |
| `Wrpd` | Annotated |
| `Lggd` | Logged |
| `Rtrd` | Returned |

### Allowed transitions by context

**Exactly one error-return site in the function:**

- `Crtd -> Rtrd`
- `Rcvd -> Rtrd`
- `Rcvd -> Wrpd -> Rtrd`
- `Rcvd -> Lggd`

**Multiple error-return sites in the function:**

- `Crtd -> Rtrd`
- `Rcvd -> Wrpd -> Rtrd`
- `Rcvd -> Lggd`
- (Notice: a bare `Rcvd -> Rtrd` — returning without annotation — is **not** allowed here;
  this is rule 4.)

Note: there is no `Crtd -> Lggd` transition. An error you create locally and then swallow
adds nothing to the log — just log a plain message without an error object. If you create
an error, return it.

**Recursive function (rule 8 takes precedence over rule 4):**

- `Crtd -> Rtrd`
- `Rcvd -> Rtrd`
- `Rcvd -> Lggd`
- (Notice: `Rcvd -> Wrpd -> Rtrd` is **not** allowed here; this is rule 8. Bare
  `Rcvd -> Rtrd` **is** allowed even with multiple error-return sites, precisely because
  annotation is forbidden in recursion.)

## Quick reference

| Operation | Text style | When |
|-----------|-----------|------|
| Annotate | "do action" (`open file`) | multi error-site functions; describes the action; unique per site; never in recursion |
| Log      | "failed to..." (`failed to open file`) | error ends in scope / leaves business-logic boundary; states the failure; full annotated chain |
| Return   | — | never after logging within the same function (rule 3); never after annotating within the same function (rule 9) |

## Worked example: repository -> service -> handler

```go
// Repository layer: two error-return sites -> annotate both, with unique texts
// (rules 4, 5, 6).
func (r *repo) FetchUser(ctx context.Context, id int64) (*User, error) {
    row := r.db.QueryRowContext(ctx, userQuery, id)

    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        return nil, fmt.Errorf("fetch user %d: %w", id, err)
    }

    if err := u.Validate(); err != nil {
        return nil, fmt.Errorf("validate user %d: %w", id, err)
    }

    return &u, nil
}

// Service layer: two error-return sites again -> annotate both uniquely; context
// accumulates.
func (s *svc) GetUser(ctx context.Context, id int64) (*User, error) {
    u, err := s.repo.FetchUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user profile: %w", err)
    }

    if u.Suspended {
        return nil, fmt.Errorf("check user suspension: %w", ErrSuspended)
    }

    return u, nil
}

// Handler: terminal boundary -> log the full chain (rules 2, 7), then answer the
// client. Never return the logged error (rule 3).
func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    u, err := h.svc.GetUser(r.Context(), parseID(id))
    if err != nil {
        log.Printf("failed to get user %s: %v", id, err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, u)
}
```

The boundary log line then reads, for example:

    failed to get user 42: get user profile: fetch user 42: validate user 42: <cause>

Essence first (the log message), then context (the annotation chain).

## Practical flow

- Single error-return site → any of `Crtd -> Rtrd`, `Rcvd -> Rtrd`,
  `Rcvd -> Wrpd -> Rtrd`, or `Rcvd -> Lggd` is valid. You may freely return; annotation
  here is optional.
- Function with 2+ error-return sites → wrap returned errors with context describing
  the action (rules 4, 5) and do **not** log them in the same function (rule 9). Return them
  and let the caller decide. The only exception to wrapping is recursion (see below).
- Recursive code → pass the error up without annotation (rule 8 takes precedence over
  rule 4); log at the boundary.
- Terminal boundary (handler) → log the full annotated chain with a "failed to..."
  message (rules 2, 7), then translate it into a client-facing response. Do not return
  the error any further.
