# API Gateway

A production-ready API Gateway written in Go, featuring per-service rate limiting, configurable middleware pipelines, and circuit breakers — all driven by a single YAML file.

---

## Features

- **Circuit Breaker** — per-service, with Closed / Open / Half-Open state machine and configurable failure threshold and retry timeout
- **Rate Limiting** — two strategies per service: Token Bucket and Sliding Window
- **Middleware Pipeline** — register middleware globally, apply per-service; control which middlewares run before the circuit breaker check
- **YAML Configuration** — all behavior is declared in `services.yml`, no code changes required to add or modify services

---

## Architecture

Every incoming request flows through a layered pipeline. The order is deliberate: cheap checks (circuit breaker state) happen before expensive ones (downstream calls).

```
Incoming Request
       │
       ▼
 ┌─────────────┐
 │  Before MWs  │  ← middlewares declared in circuit_breaker.before (e.g. logging)
 └──────┬──────┘
        │
        ▼
 ┌──────────────────┐
 │  Circuit Breaker  │  ← if Open and retryTimeout not elapsed: reject immediately
 └──────┬───────────┘
        │
        ▼
 ┌─────────────┐
 │  After MWs   │  ← remaining middlewares (e.g. auth, cors)
 └──────┬──────┘
        │
        ▼
 ┌──────────────┐
 │  Rate Limiter │  ← token bucket or sliding window
 └──────┬───────┘
        │
        ▼
 ┌──────────────┐
 │  Proxy Call   │  ← reverse proxy to the target service
 └──────────────┘
```

### Circuit Breaker States

```
          failure >= threshold
 Closed ─────────────────────► Open
   ▲                              │
   │                              │ retryTimeout elapsed
   │                              ▼
   └──────────────────────── HalfOpen
          success                 │
                                  │ failure
                                  ▼
                                Open
```

---

## Configuration

```yaml
# services.yml

services:
  users:
    name: users
    target: http://localhost:8000/api

    # Middlewares applied after the circuit breaker check.
    # Must be registered via proxy.Use() before the gateway starts.
    middlewares:
      - auth_jwt
      - cors
      - logger

    # Supports multiple strategies applied in sequence.
    rate_limits:
      - type: bucket       # Token Bucket — smooth request flow
        rate: 20           # tokens added per second
        burst: 5           # max burst size

      - type: window       # Sliding Window — hard limit per interval
        limit: 60          # max requests allowed
        interval: "1m"     # window duration (e.g. "30s", "1m", "1h")

    circuit_breaker:
      failure_threshold: 5       # consecutive failures before opening the circuit
      retry_timeout: "10s"       # how long to wait before trying again (Half-Open)
      before:
        - logger                 # runs before the circuit breaker check
                                 # must also be declared in middlewares above

  orders:
    name: orders
    target: http://localhost:9000/api

    middlewares:
      - logger
      - auth_jwt

    rate_limits:
      - type: bucket
        rate: 50
        burst: 10

    circuit_breaker:
      failure_threshold: 3
      retry_timeout: "30s"
      before:
        - logger
```

---

## Middleware Registration

Middlewares are registered once at startup and referenced by name in the YAML:

```go
package main

import (
    "log"
    "net/http"

    "github.com/you/api-gateway/internal/proxy"
)

func main() {
    // Register a logging middleware
    proxy.Use("logger", func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            log.Printf("%s %s", r.Method, r.URL.Path)
            next(w, r)
        }
    })

    // Register a JWT auth middleware
    proxy.Use("auth_jwt", func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            next(w, r)
        }
    })

    // Register a CORS middleware
    proxy.Use("cors", func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            next(w, r)
        }
    })

    http.ListenAndServe(":8080", proxy.Router())
}
```

---

## Rate Limiting Strategies

### Token Bucket

Best for APIs that need to allow short bursts while keeping a steady average rate.

```yaml
rate_limits:
  - type: bucket
    rate: 20     # 20 tokens added per second
    burst: 5     # allows up to 5 requests instantly before throttling
```

Backed by `golang.org/x/time/rate`. Thread-safe without additional locking.

### Sliding Window

Best for enforcing hard quotas over a fixed period (e.g. 60 requests per minute).

```yaml
rate_limits:
  - type: window
    limit: 60
    interval: "1m"
```

The window resets after `interval` elapses from the first request in the current window.

---

## Circuit Breaker

The circuit breaker protects downstream services from cascading failures. It operates as a state machine with three states:

| State | Behavior |
|---|---|
| **Closed** | Requests pass through normally |
| **Open** | Requests are rejected immediately with `503` |
| **Half-Open** | One request is allowed through to test recovery; success closes, failure reopens |

```yaml
circuit_breaker:
  failure_threshold: 5   # open after 5 consecutive failures
  retry_timeout: "10s"   # try Half-Open after 10 seconds
  before:
    - logger             # these middlewares run even when the circuit is Open
```

The `before` list controls which middlewares execute before the circuit breaker check. This is useful for logging and metrics — you may want to record rejected requests even when the circuit is open.

---

## Project Structure

```
api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── proxy/
│   │   └── proxy.go       # reverse proxy, pipeline assembly, middleware registry
│   ├── ratelimit/
│   │   ├── bucket.go      # token bucket strategy
│   │   ├── window.go      # sliding window strategy
│   │   └── config.go
│   ├── circuitbreaker/
│   │   ├── breaker.go     # state machine: Closed, Open, HalfOpen
│   │   └── config.go
└── go.mod
└── services.yml
```

---

## Running Locally

```bash
# Clone the repository
git clone https://github.com/you/api-gateway
cd api-gateway

# Run the gateway
go run ./cmd/gateway main.go
```

The gateway listens on `:8080` by default.

---

## Running with Docker Compose

```yaml
# docker-compose.yml
services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./services.yml:/app/services.yml

  users:
    image: your-users-service
    ports:
      - "8000:8000"

  orders:
    image: your-orders-service
    ports:
      - "9000:9000"
```

```bash
docker compose up
```

---

## Design Decisions

**Why `before` in the circuit breaker config?**
Most gateways apply all middlewares before or after the circuit breaker — with no per-service control. Here, `before` lets you declare exactly which middlewares should run even when the circuit is open. Logging and metrics make sense there; auth and business logic do not.

**Why two rate limiting strategies?**
Token bucket and sliding window solve different problems. Bucket is better for smoothing bursty traffic; window is better for strict quotas. Declaring both in sequence lets you enforce a burst limit and a per-minute cap independently on the same service.

**Why parse `time.Duration` from strings?**
Accepting `"10s"`, `"500ms"`, and `"1m"` instead of plain integers avoids ambiguity about units and matches the convention used by Docker Compose, Kubernetes, and the Go standard library itself.