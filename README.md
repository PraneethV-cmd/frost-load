# frost-load

A load balancer written in Go. Currently an **L7 (HTTP) reverse proxy** with
weighted round-robin balancing and automatic health checking.

## What it does

- **Weighted round-robin** — spreads requests across backends; higher-weight
  backends receive proportionally more traffic.
- **L7 reverse proxy** — forwards real HTTP requests to a chosen backend and
  streams the response back (`net/http/httputil`).
- **Automatic health checks** — a background goroutine probes each backend's
  `/health` every few seconds and takes dead ones out of rotation, adding them
  back when they recover. No manual intervention.
- **Concurrency-safe** — backend liveness, connection counts, and the
  round-robin cursor use `sync/atomic`; health probes fan out with a
  `sync.WaitGroup`.

## Layout

```
cmd/backend/       tiny test backend — one binary, run many copies (NAME/PORT env)
cmd/frostload/     the load balancer entry point (listens on :9000)
internal/pool/     backend pool: weighted rotation, skip-dead selection (pool.go)
                   + background health checker (health.go)
internal/lb/       L7 reverse-proxy handler (lb7.go)
```

## Run it

Start three backends (each identifies itself by NAME):

```sh
NAME=web-1 PORT=8080 go run ./cmd/backend
NAME=web-2 PORT=8081 go run ./cmd/backend
NAME=web-3 PORT=8082 go run ./cmd/backend
```

Start the load balancer (backend addresses/weights are set in
`cmd/frostload/main.go`):

```sh
go run ./cmd/frostload
```

Send some traffic and watch the distribution:

```sh
for i in $(seq 12); do curl -s localhost:9000/; echo; done
```

The weighted backend appears proportionally more often. Kill a backend and,
after a few seconds, it drops out of rotation automatically; restart it and it
comes back.

## Backend endpoints

- `GET /` — returns the backend's `{name, port}` (so you can see who answered)
- `GET /health` — returns `{"status":"ok"}`; used by the health checker
- `GET /sleep` — sleeps 3s, for testing slow responses

## Roadmap

- [x] Test backend
- [x] Backend pool + weighted round-robin
- [x] L7 reverse proxy
- [x] Health checks
- [ ] Least-connections balancing
- [ ] L4 (TCP) proxy
- [ ] Docker Compose
- [ ] Nix (dev shell + image)
