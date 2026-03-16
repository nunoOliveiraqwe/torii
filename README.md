<p align="center">
<pre>
           _                ____
 _ __ ___ (_) ___ _ __ ___ |  _ \ _ __ _____  ___   _
| '_ ` _ \| |/ __| '__/ _ \| |_) | '__/ _ \ \/ / | | |
| | | | | | | (__| | | (_) |  __/| | | (_) >  <| |_| |
|_| |_| |_|_|\___|_|  \___/|_|   |_|  \___/_/\_\\__, |
                                                  |___/
</pre>
</p>

<p align="center">
  <b>A tiny reverse proxy for your home server. Keep it simple, keep it fun.</b>
</p>

---

> **Heads up!** microProxy is in early development. Things are being built, things will break, things will change. Feel
> free to poke around, but expect rough edges.

## Why microProxy?

Setting up a reverse proxy at home shouldn't feel like configuring a space shuttle.

There are some truly great proxies out there: Nginx, Traefik, HAProxy, Caddy... they're battle-tested, feature-rich, and
trusted by thousands of companies. microProxy does **not** aim to compete with any of them. Not even close.

So why build another one?

Because sometimes you just have a home server with a few services, and all you want is:

- Point a domain at a backend.
- Get TLS if you want it.
- Done. Go grab a coffee.

Those other proxies can absolutely do this, but they come with so many options and knobs that it can feel overwhelming
for a simple home setup. microProxy keeps things to the bare minimum on purpose. Less config, less confusion, more free
time.

## What it does

- **HTTP and HTTPS reverse proxy** with virtual hosting (route traffic by hostname).
- **Automatic TLS** via Let's Encrypt, or bring your own certificates.
- **Single-file YAML configuration.** One file, flat structure, no surprises.
- **A management API** that will eventually power a small **web GUI** for setting up routes, checking metrics, and
  seeing what's going on, all without touching a config file if you don't want to.

## Quick Start

### 1. Build

```bash
go build -o micro-proxy ./cmd/micro-proxy
```

### 2. Create a config file

Create a `config.yaml`:

```yaml
log:
  debug: false
  logLevel: INFO

apiServer:
  host: 127.0.0.1
  port: 27000

netConfig:
  http:
    - port: 80
      bind: 3          # 1 = IPv4, 2 = IPv6, 3 = both
      routes:
        - host: app.home.local
          backend: http://192.168.1.50:8080

    - port: 443
      bind: 3
      tls:
        use-acme: true
      routes:
        - host: app.home.local
          backend: http://192.168.1.50:8080

  acme:
    email: you@example.com
    cache: /var/lib/micro-proxy/certs
```

### 3. Run

```bash
./micro-proxy --config config.yaml
```

That's it. Seriously.

## Configuration Reference

### Top-level

| Key              | Type     | Default     | Description                                   |
|------------------|----------|-------------|-----------------------------------------------|
| `log.debug`      | `bool`   | `false`     | Enable development logging.                   |
| `log.logLevel`   | `string` | `INFO`      | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). |
| `log.logPath`    | `string` |             | File path to write logs to.                   |
| `apiServer.host` | `string` | `127.0.0.1` | Management API bind address.                  |
| `apiServer.port` | `int`    | `27000`     | Management API port.                          |

### HTTP Listener (`netConfig.http[]`)

| Key                | Type     | Description                              |
|--------------------|----------|------------------------------------------|
| `port`             | `int`    | Port to listen on.                       |
| `bind`             | `int`    | `1` = IPv4, `2` = IPv6, `3` = both.      |
| `interface`        | `string` | Network interface to bind to (optional). |
| `tls.use-acme`     | `bool`   | Use Let's Encrypt for automatic TLS.     |
| `tls.cert`         | `string` | Path to TLS certificate (manual TLS).    |
| `tls.key`          | `string` | Path to TLS private key (manual TLS).    |
| `routes[].host`    | `string` | Hostname to match (virtual hosting).     |
| `routes[].backend` | `string` | Backend URL to proxy to.                 |
| `default.backend`  | `string` | Catch-all backend when no route matches. |

### ACME (`netConfig.acme`)

| Key            | Type     | Description                           |
|----------------|----------|---------------------------------------|
| `email`        | `string` | Registration email for Let's Encrypt. |
| `cache`        | `string` | Directory to store certificates.      |
| `open-port-80` | `bool`   | Open port 80 for HTTP-01 challenges.  |

## CLI Flags

```
--config <path>     Path to the YAML configuration file
--debug             Enable debug logging (overrides config)
--log-level <level> Set log level (overrides config)
```

## Design Principles

1. **Keep it simple.** If a feature isn't useful for a home setup, it probably doesn't belong here.
2. **One config file.** No includes, no fragments, no directory of configs. Everything in one place.
3. **Sane defaults.** It should just work with as little configuration as possible.
4. **Stay out of the way.** Set it up, forget about it, enjoy your weekend.

## Requirements

- **Go 1.25+**

## License

This project is licensed under the [Apache License 2.0](LICENSE).
