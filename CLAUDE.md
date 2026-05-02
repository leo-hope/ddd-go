# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test ./core/... -run TestCommandBusDispatch

# Vet and format
go vet ./...
gofmt -w .
```

## Architecture

This is a Go library (`github.com/leo-hope/ddd-go`) with no external dependencies that provides DDD building blocks in two packages.

### `common/` — Domain primitives

| Type | Purpose |
|---|---|
| `Entity`, `IdentityEntity`, `UUIDEntity`, `ConcurrencySafeEntity` | Base structs for domain entities (embedded, not implemented as interfaces) |
| `ValueObject` | Marker base for immutable value types |
| `BaseCommand[R]`, `Command[R]`, `Query[R]`, `PagingQuery[R]` | Generic command/query hierarchy; embed `BaseCommand[R]` in every concrete command struct |
| `BaseEvent` | Embed in domain event structs; provides `OccurredTime()` and a no-op `Tag()` |
| `Result[T]`, `MultiResult[T]`, `PagingResult[T]` | Typed response envelopes returned by handlers |
| `BizError`, `ErrorCode` | Structured domain errors; use `IsBizError`/`IsConcurrencyConflict` to inspect errors |
| `CommandValidator` / `GlobalCommandValidator` | Validator interfaces wired into the bus via `RegisterValidator` |

### `core/` — Bus infrastructure

**CommandBus** (`CommandBus` interface, `DefaultCommandBus` impl):
- Create with `NewCommandBus()`
- Register handlers (`RegisterHandler`), interceptors (`RegisterInterceptor`), validators (`RegisterValidator`)
- Call `Init()` **after** all registrations — it wires the validate interceptor
- Dispatch with the type-safe package function: `core.Dispatch[R](ctx, bus, cmd)`

**Handlers** — two styles:
- Struct-based: implement `Handler` interface (`SupportCommand() reflect.Type` + `Handle`)
- Function-based: `core.NewHandlerFunc(func(ctx, cmd) (R, error))` — no boilerplate

**Interceptors** (`CommandInterceptor` interface):
- Embed `CommandInterceptorBase` and override only `BeforeHandle`/`AfterHandle`
- Scope: `NewGlobalInterceptorBase(order)` for all commands; `NewInterceptorBase[*CmdType](order)` for one command type
- Lower `Order()` values run first

**EventBus** (`EventBus` interface, `DefaultEventBus` impl):
- Create with `NewEventBus()`
- Register handlers with `RegisterHandler`; implement `EventHandler` by embedding `BaseEventHandler[*YourEvent]`
- `Publish` is synchronous; `AsyncPublish` fires each handler in its own goroutine
- Override `WithExceptionHandler` to replace the default logging exception handler

**Assembler** (`Assembler[Source, Target]`): thin interface for domain→DTO conversions.

### Typical wiring sequence

```go
eventBus := core.NewEventBus()
eventBus.RegisterHandler(myEventHandler)

bus := core.NewCommandBus()
bus.RegisterHandler(myCommandHandler)
bus.RegisterValidator(myValidator)
bus.RegisterInterceptor(myInterceptor)
bus.Init() // must be last

result, err := core.Dispatch[common.Result[string]](ctx, bus, &MyCommand{})
```