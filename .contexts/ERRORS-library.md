# LLM_CONTEXT.md

Usage context for application code that depends on `github.com/sirkon/errors` (requires Go 1.26+).
This file is for *consumers* of the library; for working on the library itself see `AGENTS.md`.

## Mental model

Errors are **processes, not values**. An `*errors.Error` is a mutable, append-only chain of
processing layers (`New` → `Wrap`/`Just` layers). Each layer carries two kinds of information:

1. **Text annotation** — the message of `New`/`Wrap`. These are what `Error()` joins into the
   classic `"outer: …: inner"` text chain (`Just` layers add no text).
2. **Structured context** — typed key/value pairs attached to the layer via the fluent setters
   (`Str`, `Int`, `Bool`, …). This mimics structured logging (the values are `slog.Value`s under
   the hood) and is rendered alongside the text by the slog handlers or any custom consumer — so
   details added deep in the stack surface at the final log site without re-logging the error at
   every stage.

The same instance is extended in place as it travels up the stack, and both parts are rendered at
the end (by `Error()`, by slog handlers, or by a custom consumer). Consequences:

- Never compare `*Error` instances and never use them as sentinels — `Is`/`As` deliberately
  refuse to match `*Error`. Use `errors.NewSentinel` / `errors.NewSentinelf` for comparable
  static errors.
- `Wrap`/`Just`/`Spec` called on an existing `*Error` mutate it and return the **same pointer**.
  Don't copy, don't cache an earlier reference expecting it to stay unchanged.
- Foreign errors (`fmt.Errorf("%w", ...)`, stdlib errors, third-party errors) can be wrapped at
  any point; they are traversed recursively by `Error()`, `Is`/`As`, and all context renderers.

## Constructors and annotation (package `errors`)

```go
errors.New("msg")                     // *Error; errors.Newf(format, args...)
errors.Wrap(err, "annotation")        // *Error; errors.Wrapf(err, format, args...)
errors.Just(err)                      // add context without a text annotation
errors.NewSentinel("static")          // comparable sentinel error
errors.NewSentinelf(format, args...)
errors.Spec(err, mark)                // attach invisible domain marker (any value)
errors.AsSpec[T](err) (T, bool)       // retrieve marker by type, descends through wraps
errors.IsSpec[T](err) bool
errors.Is / As / AsType / Join        // mirrors of the stdlib functions
```

Fluent context setters on `*Error` (all return the same `*Error`):
`Bool, Int, I8, I16, I32, I64, Uint, U8, U16, U32, U64, F32, F64, Str, Stg(fmt.Stringer),
Strs([]string), Bytes([]byte), Any`.

Typical flow:

```go
err := errors.New("read config").Str("path", path)
if err != nil { ... }
return errors.Wrap(err, "load service config").Int("retries", n)
// or without text: return errors.Just(err).Int("retries", n)
```

`Error()` renders outermost layer first: `"wrap no 2: wrap: new error"`.
`Bytes` renders printable UTF-8 as a string and keeps binary data raw (renderers hex-dump it).

### Formatting vs context

Choosing between formatting a value into the message and attaching it as structured context:

- A value that is the **reason** for the error → format it into the text:
  `errors.Newf("failed to do %s", action)`
- A value that is merely related context → attach it as structured context:
  `errors.New("failed to perform operation").Stg("failed-op", action)`
- If the reason consists of a **set of 2+ values**, it is context too — don't format several
  values into the message, attach them all as context instead.

### Locations

`errors.InsertLocations()` (global, opt-in) records `file:line` at every `New`/`Wrap`/`Just`.
**Dev-only** — it uses `runtime.Caller` and is expensive; leave it off in production. Default is
off (`DoNotInsertLocations()`).

### Specs (invisible markers)

`Spec` attaches a typed payload that never appears in any output. Retrieve with
`AsSpec[T]`/`IsSpec[T]`; lookup descends through this library's wraps and `fmt.Errorf("%w")`.
Idiomatic pattern: `errors.Spec(err, new(struct{ ... }))` or `errors.Spec(err, new(0))` with a
distinct type per case.

## slog integration (package `errorsctx`)

Log the error **directly as a slog argument** — a bare value or an attr; empty/`!BADKEY` keys are
normalized to `"err"`:

```go
logger.Error("failed to process", err)                 // bare error works
logger.Error("failed to process", slog.Any("err", err))
```

Handlers (wrap any existing `slog.Handler`):

- `errorsctx.NewSLogHandlerTree(h)` — replaces the error attr with a group:
  `err.@text` (message) + `err.@context` (layers: `NEW: …`, `WRAP: …`, `CTX`, each with
  `@location` when enabled and its key/values).
- `errorsctx.NewSLogHandlerFlat(h)` — message under `err`, all context values flattened under a
  single `@err` group (locations under `@err.@locations`).
- `errorsctx.NewSlogPrettyRenderer(dst, opts, isDark bool, hexLimit int)` — terminal handler for
  dev mode with colors, tree rendering, hex dumps, and auto-detection of embedded JSON/multiline
  strings. `hexLimit`: `0` → 32 bytes, `-1` → no truncation. Use it together with
  `errors.InsertLocations()` in development.
- `errorsctx.ForceTree()` — a marker `slog.Attr` you can add to any record to force tree-style
  rendering by the pretty renderer.

Non-`*errors.Error` values pass through all handlers untouched.

### Inspecting errors in tests

When a test needs to see what an error carries, don't assert on `err.Error()` — log it through
`errorsctx.NewLLMTestLogger(t)` (tree-shaped JSON over `testing.T` output; there is also
`NewHumanTestingLogger(t, isDark)` for humans):

```go
func TestShowcaseOfTreeJSONRenderer(t *testing.T) {
	errors.InsertLocations()

	err := errors.New("this is an error")
	err = errors.Wrap(err, "wrap 1").Int("count", 333)
	err = errors.Just(err).Bool("is-ctx-only-wrap", true)

	logger := errorsctx.NewLLMTestLogger(t)

	logger.Warn("example", slog.Any("err", err))
}
```

Output (`go test -vet=off -v -run TestShowcaseOfTreeJSONRenderer ./errorsctx/`):

```json
{
  "time": "2026-08-08T16:13:49.359723991+03:00",
  "level": "WARN",
  "msg": "example",
  "err": {
    "@text": "wrap 1: this is an error",
    "@context": {
      "NEW: this is an error": {
        "@location": ".../errorsctx/testing_loggers_test.go:13"
      },
      "WRAP: wrap 1": {
        "@location": ".../errorsctx/testing_loggers_test.go:14",
        "count": 333
      },
      "CTX": {
        "@location": ".../errorsctx/testing_loggers_test.go:15",
        "is-ctx-only-wrap": true
      }
    }
  }
}
```

`@text` is the rendered message; each `@context` group is one processing layer (`NEW:` /
`WRAP:` / `CTX`) with its structured values and `@location` (only when `InsertLocations()` is
enabled). Times and paths vary per run.

### Custom logging backends

Two routes:

1. Direct (no interface overhead):
   `errors.SLogTreeContext(err) []slog.Attr` / `errors.SLogFlatContext(err) []slog.Attr` —
   only when you already hold a `*errors.Error`.
2. Generic, works on any `error` chain:
   `dlr := errors.GetContextDeliverer(err)` returns `nil` for foreign errors; otherwise
   `dlr.Deliver(consumer)` replays layers and values into your implementation of
   `errors.ErrorContextConsumer` / `errors.ErrorContextBuilder` (contracts in
   `logging_contracts.go`; reference implementation: `errorsctx.Consumer`, which collects
   `[]errorsctx.Layer` with `Kind` (`NEW`/`WRAP`/`CTX`), `What`, `Pos`, `Pairs []slog.Attr`).

## Gotchas for application code

- `go vet` flags bare errors passed to slog (`logger.Error("msg", err)` should be a string or a
  slog.Attr). This is the *intended* usage here; silence vet for those call sites if needed.
- Don't use `errors.New`/`Newf` for sentinel errors — they are untargetable by `Is`. Use
  `NewSentinel`/`NewSentinelf`.
- Context keys are conventionally kebab-case in this ecosystem (`"is-wrap-layer"`,
  `"text-bytes"`); match it when writing new code.
- `Stg` calls `String()` immediately at attach time; lazy rendering is not supported.
- The library is faster than `fmt.Errorf` and scales better with context size (see README
  benchmarks), so prefer attaching context over formatting it into messages.
