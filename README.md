# API Gateway

A production-ready API Gateway written in Go, featuring per-service rate limiting, configurable middleware pipelines, and circuit breakers — all driven by a single YAML file.

---

## Features

- **Circuit Breaker** — per-service, with Closed / Open / Half-Open state machine and configurable failure threshold and retry timeout
- **Rate Limiting** — two strategies per service: Token Bucket and Sliding Window
- **Middleware Pipeline** — register middleware globally, apply per-service; control which middlewares run before the circuit breaker check
- **Retry with Backoff** — four strategies (constant, linear, exponential, jitter) configurable per-service; integrates with the circuit breaker to abort early if the circuit opens mid-retry
- **YAML Configuration** — all behavior is declared in `gateway.yml`, no code changes required to add or modify services

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

The configuration file has two levels: root-level keys that apply to the gateway as a whole, and per-service keys under `services`.

```yaml
# services.yml

# --- Gateway-level configuration ---

service_lookup_header: "x-gateway-service"  # header used to identify the target service
                                             # default: "x-gateway-service"

cors_allowed_origins:
  - "http://127.0.0.1:5500"

# Global middlewares — run for every service before per-service middlewares.
# Registered via gtw.UseGlobal() at startup.
middlewares:
  - logger

# Global circuit breaker — applied to any service that does not define its own.
circuit_breaker:
  failure_threshold: 10      # consecutive failures before opening the circuit
  retry_timeout: "30s"       # how long to wait before trying again (Half-Open)
  before:
    - logger                 # runs before the circuit breaker check

# Global retry — applied to any service that does not define its own.
retry:
  attempts: 3
  backoff: exponential
  delay: "400ms"

# --- HTTP server configuration ---
server:
  listen_addr: "9091"
  timeout: "6s"              # default request timeout
  read_header_timeout: "500ms"

# --- Per-service configuration ---
services:
  users:
    name: users
    target: http://localhost:8000/api
    timeout: "10s"           # overrides server.timeout for this service

    # Per-service middlewares — run after global middlewares and after the
    # circuit breaker check. Must be registered via gtw.Use() at startup.
    middlewares:
      - auth

    # Supports multiple strategies applied in sequence.
    rate_limits:
      - type: bucket         # Token Bucket — smooth request flow
        rate: 20             # tokens added per second
        burst: 5             # max burst size

      - type: window         # Sliding Window — hard limit per interval
        limit: 60            # max requests allowed
        interval: "60s"      # window duration (e.g. "30s", "1m", "1h")
```

---

## Middleware Registration

Middlewares are registered once at startup and referenced by name in the YAML:

```go
package main

import (
	"log"
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/contracts"
	"github.com/serge1997/apigateway/internal/router"
	"github.com/serge1997/apigateway/internal/server"
)

func main() {
	gtw := apigateway.New()
	srv := server.New(
		server.WithApiGateway(gtw),
		server.WithRouter(router.New()),
	)
	defer srv.Close()

	// Global middleware — runs for every service.
	// Declared under "middlewares" in the root of services.yml.
	gtw.UseGlobal("logger", func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("%s - %s", ctx.Method(), ctx.Path())
		return next, nil
	})

	// Per-service middleware — declared under each service's "middlewares" in services.yml.
	gtw.Use("auth", func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		if ctx.Header("Authorization") != "Bearer my-secret-token" {
			return nil, ctx.Unauthorized()
		}
		return next, nil
	})

	log.Fatal(srv.Listen())
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

The circuit breaker can be declared globally (applies to all services) or per-service (overrides the global):

```yaml
# global — root level
circuit_breaker:
  failure_threshold: 10
  retry_timeout: "30s"
  before:
    - logger             # these middlewares run even when the circuit is Open

# per-service — overrides the global for this service only
services:
  orders:
    circuit_breaker:
      failure_threshold: 3
      retry_timeout: "10s"
      before:
        - logger
```

The `before` list controls which middlewares execute before the circuit breaker check. This is useful for logging and metrics — you may want to record rejected requests even when the circuit is open.

---

## Retry with Backoff

When a downstream service fails, the gateway can retry the request automatically using a configurable backoff strategy. Each retry checks the circuit breaker state before attempting — if the circuit opened during retries, the gateway aborts immediately instead of continuing to hammer an unhealthy service.

### Strategies

| Strategy | Behavior |
|---|---|
| **constant** | Fixed delay between every attempt |
| **linear** | Delay grows linearly with each attempt |
| **exponential** | Delay doubles with each attempt |
| **jitter** | Random delay up to a configured maximum — spreads retries across time to avoid thundering herd |

### Configuration

Retry can be declared globally (applies to all services) or per-service:

```yaml
# global — root level
retry:
  attempts: 3
  backoff: exponential
  delay: "400ms"

# per-service — overrides the global for this service only
services:
  orders:
    retry:
      attempts: 5
      backoff: jitter
      delay: "200ms"
```

### How it integrates with the Circuit Breaker

At the start of each attempt, the gateway checks whether the circuit breaker is open. If it is, the retry loop exits immediately and returns `503 Service Unavailable` — no further calls are made to the downstream service.

```
attempt 1 → cb open? no  → call service → failure → RecordFailure()
attempt 2 → cb open? no  → call service → failure → RecordFailure() → cb opens
attempt 3 → cb open? yes → abort immediately → 503
```

Every success calls `RecordSuccess()` and every failure calls `RecordFailure()` on the circuit breaker, so the two mechanisms stay in sync without any manual coordination.

### Backoff intervals

```
constant:     100ms ─── 100ms ─── 100ms
linear:       100ms ─── 200ms ─── 300ms
exponential:  100ms ─── 200ms ─── 400ms
jitter:       ~300ms ── ~750ms ── ~100ms  (random, up to max_delay)
```

Jitter is recommended for high-traffic services — when many clients retry at the same interval they hit the recovering service simultaneously. Randomizing the delay spreads the load.

---

## Project Structure

```
api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── apiGateway/
│   │   └── gateway.go     # gateway core — middleware registry, pipeline assembly
│   ├── server/
│   │   └── server.go      # HTTP server setup, options pattern
│   ├── proxy/
│   │   └── proxy.go       # reverse proxy and request pipeline
│   ├── contracts/
│   │   └── *.go           # shared interfaces — Context, Service, Router
│   ├── ratelimit/
│   │   ├── bucket.go      # token bucket strategy
│   │   ├── window.go      # sliding window strategy
│   │   └── config.go
│   ├── circuitbreaker/
│   │   ├── circuitbreaker.go  # state machine: Closed, Open, HalfOpen
│   │   └── config.go
│   ├── retry/
│   │   ├── retry.go       # shared RecordFailure/RecordSuccess logic
│   │   ├── constant.go
│   │   ├── linear.go
│   │   ├── exponential.go
│   │   └── jitter.go
│   └── router/
│       └── router.go      # request routing by service_lookup_header
├── go.mod
└── services.yml           # gateway configuration
```

---

## Running Locally

```bash
# Clone the repository
git clone https://github.com/serge1997/apigateway
cd apigateway

# Run the gateway
go run ./cmd/gateway/main.go
```

The gateway listens on the port declared in `server.listen_addr` in `services.yml`.

---

## Running with Docker Compose

```yaml
# docker-compose.yml
services:
  gateway:
    build: .
    ports:
      - "9091:9091"
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

**Why global circuit breaker and retry with per-service override?**
Declaring a global default eliminates repetition — most services share the same resilience policy. A service that needs different behavior declares only what changes. This follows the same convention as Docker Compose and Kubernetes, where a base config is extended rather than duplicated.

**Why `before` in the circuit breaker config?**
Most gateways apply all middlewares before or after the circuit breaker — with no per-service control. Here, `before` lets you declare exactly which middlewares should run even when the circuit is open. Logging and metrics make sense there; auth and business logic do not.

**Why two rate limiting strategies?**
Token bucket and sliding window solve different problems. Bucket is better for smoothing bursty traffic; window is better for strict quotas. Declaring both in sequence lets you enforce a burst limit and a per-minute cap independently on the same service.

**Why does retry check the circuit breaker at the start of each attempt instead of after?**
If the circuit opens during a retry loop, there is no point executing the next attempt — the downstream service is already considered unhealthy. Checking at the top of each iteration exits early without an unnecessary network call, which is exactly what the circuit breaker is designed to prevent.

**Why a configurable `service_lookup_header`?**
Different organizations have different header naming conventions. Instead of hardcoding `x-gateway-service`, the gateway lets the operator declare which header to use for service routing — making it easier to integrate into existing infrastructure without changing client code.

**Why parse `time.Duration` from strings?**
Accepting `"10s"`, `"500ms"`, and `"1m"` instead of plain integers avoids ambiguity about units and matches the convention used by Docker Compose, Kubernetes, and the Go standard library itself.