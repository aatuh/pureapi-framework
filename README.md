# pureapi/framework

Opinionated API framework building blocks for Go.
Combine endpoint definitions, error handling, middleware, and CRUD scaffolding
into consistent, testable APIs.

> **Status:** pre-release `v0.0.1`. APIs may change. Feedback welcome.

---

## ✨ Features

* **Endpoint scaffolding**: generic input/logic/output stacks with defaults.
* **Error handling**: `errutil` with expected/masked errors and clean JSON output.
* **Middleware**: CORS with credentials/wildcards, request logging, more to come.
* **CRUD templates**: configurable get/update endpoints with selectors, orders,
  pagination, and validation.
* **DB abstractions**: lightweight predicate/order types and helpers.

---

## 🚀 Quickstart

```go
package main

import (
  "net/http"

  "github.com/pureapi/framework/api"
  "github.com/pureapi/framework/api/errutil"
)

func main() {
  // Define a generic endpoint
  def := api.NewGenericEndpointDefinition(
    "hello",
    func(r *http.Request) (any, error) {
      return map[string]string{"msg": "hello"}, nil
    },
  )

  // Wrap with error handler
  handler := api.NewGenericEndpointHandler(def).
    WithErrorHandler(errutil.NewErrorHandler())

  http.Handle("/hello", handler)
  http.ListenAndServe(":8080", nil)
}
```

---

## 📦 Modules

* `api/errutil`: error factory & masking
* `api/middleware`: CORS, logging, etc.
* `api/input`: generic input handler with validators
* `crud`: CRUD configs and handlers
* `db`: predicates, orders, pagination
* `defaults`: default error/output handlers

---

## 🔮 Roadmap

* More middlewares (auth, metrics)
* Transaction helpers
* Example apps with SQLite/MySQL
* Stabilize APIs and cut `v1.0.0`
