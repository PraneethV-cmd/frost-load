# frost-load

A load balancer written in Go, combining **L7 (HTTP)** and **L4 (TCP)**
proxying with pluggable balancing strategies and automatic health checking.

## What it does

- **L7 reverse proxy** — forwards real HTTP requests to a chosen backend and
  streams the response back (`net/http/httputil`). Listens on `:9000`.
- **L4 TCP proxy** — balances raw TCP connections by piping bytes both ways,
  protocol-agnostic (works for anything over TCP, not just HTTP). Listens on
  `:9090`. Reuses the same backend pool as L7.
- **Pluggable strategies** (behind a `Strategy` interface):
  - **Weighted round-robin** — higher-weight backends get proportionally more
    traffic.
  - **Least-connections** — routes to the backend with the fewest in-flight
    requests; adapts to real load. (Default in `main.go`.)
- **Automatic health checks** — a background goroutine probes each backend's
  `/health` every few seconds and takes dead ones out of rotation, adding them
  back when they recover. No manual intervention.
- **Concurrency-safe** — backend liveness, connection counts, and the
  round-robin cursor use `sync/atomic`; health probes fan out with a
  `sync.WaitGroup`.

## Layout

```
cmd/backend/       tiny test backend — one binary, run many copies (NAME/PORT env)
cmd/frostload/     the load balancer entry point (L7 :9000, L4 :9090)
internal/pool/     pool.go      backend pool: weighted rotation, skip-dead selection
                   strategy.go  Strategy interface + RoundRobin / LeastConns
                   health.go    background health checker
internal/lb/       lb7.go       L7 reverse-proxy handler
                   lb4.go       L4 TCP proxy
```

## Configuring backends

Backends come from the `BACKENDS` environment variable — a comma-separated list
of full URLs:

```sh
BACKENDS="http://localhost:8080,http://localhost:8081,http://localhost:8082"
```

If `BACKENDS` is unset, it falls back to those three localhost addresses, so
local dev works with zero setup. All backends currently get weight 1.

## Run it locally

Start three backends (each identifies itself by NAME):

```sh
NAME=web-1 PORT=8080 go run ./cmd/backend
NAME=web-2 PORT=8081 go run ./cmd/backend
NAME=web-3 PORT=8082 go run ./cmd/backend
```

Start the load balancer (uses the localhost fallback):

```sh
go run ./cmd/frostload
```

Send traffic and watch the distribution:

```sh
for i in $(seq 12); do curl -s localhost:9000/; echo; done   # L7
curl -s localhost:9090/                                       # L4
```

Kill a backend and, after a few seconds, it drops out of rotation
automatically; restart it and it comes back.

## Run it with Docker

`docker-compose.yml` runs three backends + the load balancer as isolated
containers (a single multi-stage image, ~36 MB, holds both binaries):

```sh
docker compose up --build
```

Then hit `localhost:9000` (L7) or `localhost:9090` (L4). The backends are
internal to the compose network — only the load balancer is exposed, matching a
real front-door topology. `BACKENDS` is set in the compose file to the backend
service names.

## Backend endpoints

- `GET /` — returns the backend's `{name, port}` (so you can see who answered)
- `GET /health` — returns `{"status":"ok"}`; used by the health checker
- `GET /sleep` — sleeps 3s, for testing slow responses

## Roadmap

- [x] Test backend
- [x] Backend pool + weighted round-robin
- [x] L7 reverse proxy
- [x] Health checks
- [x] Least-connections balancing
- [x] L4 (TCP) proxy
- [x] Docker Compose
- [ ] Nix (dev shell + reproducible image)

## Future scope

Planned but not yet built:

- **Named, prioritized backends** — give each backend a human-readable name and
  a priority (weight), instead of the current uniform weight-1 from `BACKENDS`.
- **Interactive TUI config** — a [Charm](https://charm.sh) `huh` form to define
  backends (name / address / priority) when run interactively, while `BACKENDS`
  env stays the path for Docker/CI. (Adds the Charm dependency, which the
  project has so far avoided.)
- **Richer `BACKENDS` format** — extend the env syntax to carry name and weight
  per backend (e.g. `name=url=weight`) so Docker can run weighted backends too.
- **Unit tests** — cover the pool: weighted rotation, skip-dead, all-dead
  error, least-connections selection.
- **Nix flake** — pinned dev shell + reproducible container image via
  `dockerTools`.
