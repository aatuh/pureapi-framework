# pureapi-framework

`pureapi-framework` adds a declarative layer on top of `pureapi-core`, making it easy to describe HTTP endpoints with generics, automatic request binding, response rendering, and consistent error mapping.

## Why use it?

- **Declarative endpoints** – describe method, path, metadata, success codes, and middleware in one place.
- **Automatic binding** – populate handler inputs from path/query/header/cookie/body with validation feedback.
- **Consistent errors** – ship a catalog driven error mapper so wire responses stay predictable.
- **Minimal surface** – import the root package, build an engine, declare endpoints, and register them with the provided facade helpers.

## Quick start

```go
package main

import (
    "context"
    "log"
    "net/http"

    framework "github.com/aatuh/pureapi-framework"
)

type helloInput struct {
    ID   string `path:"id"`
    Name string `query:"name" required:"true"`
}

type helloOutput struct {
    Greeting string `json:"greeting"`
}

func main() {
    engine := framework.NewEngine()

    helloEndpoint := framework.Endpoint[helloInput, helloOutput](
        engine,
        http.MethodGet,
        "/hello/{id}",
        func(ctx context.Context, in helloInput) (helloOutput, error) {
            return helloOutput{Greeting: "hello " + in.Name + " (#" + in.ID + ")"}, nil
        },
    )

    handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
    framework.RegisterEndpoints(handler, helloEndpoint)

    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatal(err)
    }
}
```

## Key concepts

- **Engine** – owns the default binder, renderer, error mapper, and shared middleware. Extend it via `WithBinder`, `WithRenderer`, `WithErrorMapper`, and `WithGlobalMiddlewares` options.
- **Endpoint declaration** – `Endpoint[TIn, TOut]` wires inputs/outputs and per-endpoint options like `WithMeta`, `WithSuccessStatus`, `WithEndpointBinder`, and `WithEndpointRenderer`.
- **Binder** – reflection-based `DefaultBinder` covers path/query/header/cookie/body sources, size limits, context cancellation, and detailed field errors.
- **Renderer** – `JSONRenderer` writes JSON responses with optional pretty printing.
- **Codec registry** – register additional renderers via `WithRenderer` (for example plain text) and negotiate responses with `Accept` headers.
- **Error handling** – `ErrorCatalog`, `ErrorMapper`, and `RenderError` stabilise wire errors and support custom mappings.
- **Input/output hooks** – attach reusable processors (e.g. validation) via `NewInputHook`, `NewOutputHook`, and the `WithEndpoint*Hooks` options.
- **Context enrichers** – inject principals or request metadata ahead of binding with `NewContextEnricher`, `WithContextEnrichers`, and `WithEndpointContextEnrichers`.
- **Authorization policies** – gate handlers using `AuthorizationPolicyFunc`, `WithAuthorizationPolicies`, and per-endpoint overrides.
- **Access logging** – ship structured request logs via `WithAccessLoggers` and the provided helpers.
- **Security middleware** – apply CORS via `NewCORSMiddleware` and common security headers via `NewDefaultSecurityHeadersMiddleware`.
- **Facade helpers** – use `NewHTTPHandler`, `RegisterEndpoints`, `NewMiddlewares`, and `RequestIDMiddleware` straight from the root package.

## Examples

Executable examples live in [`examples/`](examples) and progress from basic HTTP setups to more advanced scenarios (including optional validation hooks). Run them with `go test ./examples/...`.

## Testing

Run the library test suite from the module root:

```bash
go test ./...
```

Set `GOCACHE=$(pwd)/.gocache` when working inside sandboxes that restrict the default build cache location.

## Compatibility

- **Go support** – tested with Go 1.23+ and kept N-1 compatible; CI and docs will call out any breaking toolchain shifts.
- **CGO** – the module builds without CGO to simplify cross compilation.
- **SemVer** – releases follow semantic versioning; breaking changes land only in new major versions and release notes will highlight migration steps.

## Contributing

- Keep `README.md` and `examples/` updated as new features land.
- Add doc.go files when introducing new packages.
- Extend the `TODO.md` checklist as features move through the roadmap.
