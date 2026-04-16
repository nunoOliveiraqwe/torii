<p align="center">
<pre>
  _              _ _
 | |_ ___  _ __ (_|_)
 | __/ _ \| '_ \| | |
 | || (_) | |   | | |
  \__\___/|_|   |_|_|
</pre>
</p>

<p align="center">
  <b>A reverse proxy built for your home lab that isn't afraid of the open internet.</b>
</p>

---

> **`v0.1.0` — early alpha.** Things will break. Things will change. But it works, and it's already running on a Pi exposed to the internet. If you find bugs, open an issue.

## Why Torii?

You're running services at home and you want to expose some of them to the internet without managing nginx configs, certbot cron jobs, or a monitoring stack with three dependencies just to see what's going on.

Torii is a single binary that handles routing, TLS, rate limiting, bot defense, geo-blocking, and observability. No Docker required. No dependencies. Just the binary and a config file.

The dashboard was the first thing built, before the proxy even forwarded a single request. I wanted to see what was happening without spinning up Grafana, Prometheus, or ELK.

## Screenshots

|                  Dashboard                   |                    Proxy Routes                    |
|:--------------------------------------------:|:--------------------------------------------------:|
| ![Dashboard](docs/screenshots/dashboard.png) | ![Proxy Routes](docs/screenshots/proxy-routes.png) |

<p align="center">
  <img src="docs/screenshots/activity.png" alt="Activity" width="80%"/>
</p>

## Features

- **Routing** — HTTP/HTTPS, virtual-host routing, path-based routing with wildcards, per-path backend overrides, path stripping
- **Automatic TLS** — Let's Encrypt via DNS-01. No port 80 required. Per-domain SNI. Background renewal.
- **Hierarchical metrics** — global → route → path. Drill down from one dropdown. Live request/error logs via SSE.
- **Web UI** — HTMX, no build pipeline. Dashboard, route viewer, server wizard, ACME management, API keys.
- **Bot defense** — UA blocking (scanners, scrapers, AI crawlers), honeypot traps (`.git/config`, `wp-login.php`), optional trickster mode (tarpits, fake creds, infinite streams), GeoIP blocking
- **Rate limiting** — token-bucket, global or per-client IP, proper `429` + `Retry-After`
- **Circuit breaker** — stops forwarding to unhealthy backends until they recover
- **Header validation** — set, strip, or require headers. Values from files or env vars.
- **API keys** — scoped keys (e.g. `read_stats` for [Homepage](https://gethomepage.dev) integration)
- **14 middlewares** — all composable, per-route or per-path. [Full list →](docs/configuration.md#middleware-reference)
- **SQLite** — single-file DB, WAL mode, embedded migrations
- **Debug mode** — `--debug` spins up stub backends for testing config without real services

## Quick Start

```bash
go build -o torii ./cmd/torii
./torii -config config.yaml
```

Open `http://127.0.0.1:27000/ui` to set your admin password.

### Minimal config

```yaml
log:
  log-level: INFO

api-server:
  host: 127.0.0.1
  port: 27000

net-config:
  http:
    - port: 80
      bind: 3
      default:
        backend: http://192.168.1.50:8080
        middlewares:
          - type: "RequestId"
          - type: "RequestLog"
          - type: "Metrics"
```
Full configuration reference: [docs/configuration.md](docs/configuration.md)

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │              torii                       │
  Incoming          │                                         │
  Request ─────────►│  Global Middlewares                      │
                    │    ▼                                     │
                    │  Host Dispatcher ──► Route Middlewares    │
                    │    ▼                                     │
                    │  Path Dispatcher ──► Path Middlewares     │
                    │    ▼                                     │
                    │  Reverse Proxy ──────────────────► Backend│
                    └─────────────────────────────────────────┘

  Management        ┌─────────────────────────────────────────┐
  (localhost)       │  API Server (:27000)                     │
                    │    ├── /ui          Web dashboard        │
                    │    ├── /api/v1      REST API             │
                    │    └── SSE stream   Real-time metrics    │
                    │                                         │
                    │  SQLite (torii.db)                       │
                    └─────────────────────────────────────────┘
```

## Installation

### Binary

```bash
go build -o torii ./cmd/torii
./torii -config config.yaml
```

### Docker

```bash
docker run -d \
  --network host \
  -v torii-data:/data \
  -v /path/to/config.yaml:/etc/torii/config.yaml \
  ghcr.io/nunoOliveiraqwe/torii:latest
```

`--network host` is recommended so Torii can bind directly to host interfaces.

| Mount | Purpose |
|---|---|
| `torii-data:/data` | SQLite database (persists across restarts) |
| `/path/to/config.yaml:/etc/torii/config.yaml` | Config file (read-write — UI can write changes back) |

<details>
<summary>Building from source (Docker)</summary>

```bash
# Multi-stage
docker build -f docker/Multistage.Dockerfile -t torii .

# Single-stage (pre-built binary)
CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o torii ./cmd/torii
cp torii docker/
docker build -f docker/Dockerfile -t torii docker/
```

> Torii uses `mattn/go-sqlite3` which requires CGO. The binary must be built with `CGO_ENABLED=1`.

</details>

### Debian / Ubuntu

```bash
dpkg -i torii_<version>_amd64.deb
```

## Performance

Tested on a **Raspberry Pi 4** (4-core ARM, 906 MB RAM) over HTTPS with full middleware chain, using [`hey`](https://github.com/rakyll/hey):

| Concurrency | Req/s | p50 | p95 | p99 |
|:-----------:|------:|----:|----:|----:|
| 10 | **663** | 12 ms | 24 ms | 98 ms |
| 100 | **656** | 146 ms | 239 ms | 359 ms |

~660 req/s sustained (~57M requests/day). Throughput stays flat as concurrency increases — latency scales linearly, which means the Pi's CPU is the bottleneck, not the proxy.

Resource usage under load: ~3 of 4 CPU cores, ~100 MB RSS. Goroutines drain back to baseline (~20) within a minute after traffic stops.

```bash
# Run your own benchmarks
go install github.com/rakyll/hey@latest
./torii -config config.yaml --debug
hey -z 2m -c 50 http://localhost:<proxy-port>/
```

## Documentation

- [Configuration & Middleware Reference](docs/configuration.md)
- [Management API](docs/api.md)

## TODO

Keeping track of what's done and what's next. This is not a roadmap, just my personal scratchpad for the project.


### Needs work
- [x] IP block lists: middleware is registered, filtering logic is stubbed out
- [ ] Create HTTP Proxy UI: works, but the UX needs another pass
- [ ] ACME UI: need a delete button for individual certificates, the reset button placement is bad
- [ ] Proxy Routes UI: host names should be clickable links that open in a new tab
- [ ] Server timeouts: review the defaults for the Create HTTP Proxy wizard

### Up next
- [ ] Config persistence: proxy routes created or deleted via the UI are memory-only — they need to be written back to the config file so they survive restarts - not sure if I will. maybe with an overridable flag, because if I want to expose the proxy API to the internet for dynamic route management, I don't want those changes written to disk, and if I want to manage the config through the file, I don't want the UI overwriting it. Maybe a `persist-ui-changes` flag that defaults to `false` and can be set per-route or globally. Also maybe don't even allow proxy route changes to be created. If the session is hijacked, then not much can be done if no config can be done. Ask for password when creating a proxy!!!!!
- [ ] Wildcard host matching: `VirtualHostDispatcher` uses exact map lookup, no support for `*.home.example.com`. Replace the `map[string]http.Handler` in `VirtualHostDispatcher` with a reversed-label trie so wildcard routes match with DNS-style semantics (one label only, most-specific wins). Priority chain: exact match → longest wildcard match → `default` → 502. Normalize to lowercase, strip trailing dots. Watch out for ACME domain collection (preserve `*.example.com` for wildcard cert issuance) and SNI matching in TLS config.
- [ ] CPU usage in system activity: add CPU usage metrics to the system health/activity dashboard alongside the existing memory and other system stats.
- [ ] Terminating middleware support in UI: update the Create/Edit Proxy UI so that when a terminating middleware (e.g., Redirect) is present in the chain, the backend field becomes optional. Currently the UI always requires a backend.
- [ ] Route editing in UI: add edit support for existing proxy routes through the web UI (there is already a comment/placeholder for this in the codebase). Users should be able to modify host, backend, middlewares, and path rules for an existing route without deleting and recreating it.
- [ ] TCP proxying: config schema is there, implementation is not
- [x] Make `RequestId`, `RequestLog`, and `Metrics` default on all endpoints (too easy to forget and then the dashboard shows nothing)
- [x] Blocked IP observability: surface blocked IPs (from honeypot, UA blocker, country block) in the UI with timestamps and metadata, probably as a rolling log
- [ ] API keys: Homepage integration endpoint, `read_config` and `write_config` scopes (may be dropped depending on how useful they turn out to be)
- [ ] UA fingerprint rotation detection: bots that rotate user agents mid-scan are easy to spot — a real client doesn't switch from macOS to Linux to Windows between requests. Track UA consistency per IP and flag or block IPs whose OS/browser family changes unnaturally fast. I encountered this with a bot that rotated through 20+ UAs in a single scan, hitting 100+ endpoints in minutes. The honeypot caught it, but this would be another layer of defense against UA rotation.
- [ ] Coraza WAF integration: [Coraza](https://coraza.io/) is a full-featured open-source WAF (OWASP CRS compatible). Add it as a middleware so routes can opt into proper WAF rules alongside the existing bot defense. There's overlap with what the honeypot and UA blocker already do, but Coraza covers a much wider surface (SQLi, XSS, protocol violations, etc.).
- [x] AbuseIPDB middleware: check client IPs against [AbuseIPDB](https://www.abuseipdb.com/) and block or flag IPs with a high abuse confidence score. Optionally report blocked IPs back (honeypot hits, rate-limit violations, etc.) so the community benefits too.


### Known Bugs
- [x] **ACME port leak on startup:** if a route has `use-acme: true` but ACME is not configured, the server fails to start. Starting the proxy manually afterwards returns an ACME error, but the port is already bound from the first attempt. A second manual start then fails with a "bind: address already in use" error. The listener from the failed first start is never closed.
- [ ] **ACME config not loadable from config file:** there is no way to specify ACME configuration (provider, credentials, etc.) in the YAML config file. On first start with `use-acme: true`, it should be possible to seed the ACME configuration from the config file instead of requiring it to be set up through the UI first.

### Maybe
- [ ] Proxy-level authentication: login pages so the proxy handles auth before forwarding to backends
- [ ] Dedicated tar pitting middleware (separate from the honeypot trickster mode)
- [ ] Login endpoint rate limiting

## Requirements

- Go 1.25+

## License

[Apache License 2.0](LICENSE)