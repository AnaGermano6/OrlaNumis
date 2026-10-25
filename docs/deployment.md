# Deployment Guide

This guide covers installing Orla Numis from its published image (Docker Compose, `docker run` or Unraid), updating it, building it from the source, and what to do before exposing it outside a home network. For a first walkthrough of the app, see the [Getting Started Guide](getting-started.md).

> **Public exposure warning:** the app suits a single computer or a trusted home network as installed here. Do **not** put it on the public internet without TLS, a hardened reverse proxy, trusted proxy configuration, closed registration, backups, monitoring and firewall rules from this guide. Anonymous attackers can brute-force sign-in, upload payloads, scrape media and probe the Python agent boundary continuously.

## Install

The published image is `anagermano/orlanumis` on Docker Hub, for `linux/amd64` and `linux/arm64`. It holds everything: the app and its AI agent. On the first start it makes its secrets and keeps them in the data folder.

### Docker Compose

Put the project's [`docker-compose.yml`](../docker-compose.yml) in a folder of its own, set `PUID`, `PGID` and `TZ` in it, then:

```sh
docker compose up -d
```

The app is at `http://<this computer>:8080`; the first account registered becomes the administrator. The data stays in the `data`, `uploads` and `backups` folders next to the file.

### docker run

```sh
docker run -d --name orlanumis --restart unless-stopped \
  -p 8080:8080 \
  -e PUID=1000 -e PGID=1000 -e TZ=Europe/Lisbon \
  -v /srv/orlanumis/data:/app/data \
  -v /srv/orlanumis/uploads:/app/uploads \
  -v /srv/orlanumis/backups:/app/backups \
  anagermano/orlanumis:latest
```

### Unraid

In **Apps** (Community Applications), search for **OrlaNumis** and install it. The template ([`templates/orlanumis.xml`](../templates/orlanumis.xml)) keeps the three folders in `/mnt/user/appdata/orlanumis/`, runs as `nobody:users` (99:100) and takes Unraid's time zone. Point **Backups** to a folder on another share to keep the backups apart from appdata.

### Folders and owner

| Path in the container | What it holds |
|---|---|
| `/app/data` | SQLite database, settings, and the secrets (`.orlanumis-secrets`) |
| `/app/uploads` | Uploaded images referenced by the database |
| `/app/backups` | Full backups, made every day |

Losing the first two loses the collection. The container starts as root only to make the three folders belong to `PUID:PGID` (only what is not theirs yet), then runs the app and the agent as that user. Use the owner of the host folders: `id -u` and `id -g` on Linux, 99 and 100 on Unraid. A network share that refuses the change of owner must already be writable by them.

### Image settings

| Variable | Default | Description |
|---|---|---|
| `PUID`, `PGID` | `10001` | User and group that run the app and own the folders |
| `TZ` | UTC | Time zone of the daily backup and the schedules |
| `AI_AGENT` | `on` | `off` leaves the AI agent out, and with it the AI features; without AI set up, it saves the agent's memory and processor time |
| `TRUSTED_PROXIES` | `none` | Behind a reverse proxy, its address; see [Trusted Proxy and Client IP Handling](#trusted-proxy-and-client-ip-handling) |
| `JWT_SECRET`, `AGENT_INTERNAL_SERVICE_TOKEN` | made on the first start | Set them only to choose your own; the given ones win over the saved ones |

The image runs with `GIN_MODE=release`. The other variables of the [full reference](#full-variable-reference) apply too.

---

## Updating

**Admin → Updates** shows the version running and the latest one, from the project's GitHub releases, with what the new one brings; administrators also see a notice on the overview page. It checks every few hours, and can be turned off there. Updating keeps the data: the app migrates it when it starts on the new version.

| Installed with | To update |
|---|---|
| Docker Compose | `docker compose pull && docker compose up -d` in the folder of `docker-compose.yml` |
| Unraid | Docker tab: **Check for Updates**, then **Apply Update** on OrlaNumis |
| `docker run` | `docker pull anagermano/orlanumis:latest`, then remove the container and create it again with the same options and folders |

Make a backup first (**Settings → Backups → Back up now**); the automatic one of the night before is also there.

### Updating from the app (Watchtower)

[Watchtower](https://github.com/nicholas-fedor/watchtower) is a container that updates other containers. With its HTTP API on, **Admin → Updates** shows **Update now**: the app makes a backup (*Before updating*), asks Watchtower to update `anagermano/orlanumis`, and reloads when the new version answers.

1. Make a long random token: `openssl rand -hex 32`.
2. With the project's `docker-compose.yml`, put `WATCHTOWER_TOKEN=<the token>` in a `.env` file next to it and run `docker compose --profile watchtower up -d`. Elsewhere, run `nickfedor/watchtower` with the Docker socket, `WATCHTOWER_HTTP_API_ENDPOINTS=update` and `WATCHTOWER_HTTP_API_TOKEN=<the token>`, on a network the app reaches.
3. In **Admin → Updates**, set the address (`http://watchtower:8080` in the same Compose project) and the same token.

With its HTTP API on, Watchtower updates only when asked, and the app asks only for its own image. The token lets whoever has it update containers on the host: keep it as private as the Docker socket, and do not publish Watchtower's port.

### Going back

The migrations only go forward. To return to an earlier version, run its image by version tag (`anagermano/orlanumis:1.0.0` instead of `latest`; on Unraid, in the container's **Repository**), then restore the *Before updating* backup in **Settings → Backups**: the data goes back to how it was before the update.

---

## Architecture Overview

The published image runs two processes under [tini](https://github.com/krallin/tini):

| Process | Port | Purpose |
|---|---|---|
| Go API + Vue SPA | 8080 | The app |
| Python LangGraph agent | 8081, on the container's `127.0.0.1` only | The AI features |

Built from the source with `docker-compose.dev.yml`, they are two containers:

| Container | Image | Port | Purpose |
|---|---|---|---|
| `app` | `aurearia-app:dev` | 8080 | Go API + Vue SPA |
| `agent` | `aurearia-agent:dev` | 8081 (internal only) | Python LangGraph agent for the AI features |

The `Dockerfile` builds the Vue frontend with Node 24 and the API with Go 1.26.6, then either the published image (`release` target, on Python 3.12) or the app alone (`app` target, on Alpine). The agent's dependencies come from `src/agent/uv.lock` with uv 0.11.22, so CI and Docker builds use the same locked resolution.

Data is stored in **SQLite** (via GORM, WAL mode) and images on the filesystem. Schema changes and data migrations run at startup; a failed step stops the startup with its name instead of serving a half-migrated database.

Image deletion commits the metadata plus a durable cleanup record before removing files. Pending cleanup of the original, thumbnail and medium image is retried on startup or by repeating the same DELETE.

The images built from the source run as a non-root user with UID/GID `10001:10001`; the published one as `PUID:PGID`.

### Background jobs during restart

Run only one API process against a database. Before starting the AI-job, Set Builder and wishlist-search-alert workers, the process marks previous `running` jobs as failed with an interruption message, then resumes `queued` work with bounded workers. Interrupted jobs are not replayed automatically: review existing results before requesting a new run.

Shutdown drains HTTP requests, stops scheduling, then cancels and waits for the worker pools within an eight-second budget. Unclaimed rows stay queued; rows still running after a forced exit reconcile on the next startup.

---

## Building from the Source

```sh
task init                                                # creates .env with random secrets
docker compose -f docker-compose.dev.yml up -d --build   # build and start
```

The app is at `http://localhost:8080`, bound to `127.0.0.1`. Other commands:

```sh
docker compose -f docker-compose.dev.yml logs -f app   # follow the API logs
docker compose -f docker-compose.dev.yml down          # stop (keeps the data)
docker compose -f docker-compose.dev.yml down -v       # stop and wipe the database and uploads (not the backups)
```

A second, separate stack can run alongside, for example for testing:

```sh
APP_PORT=8090 docker compose -p aurearia-qa -f docker-compose.dev.yml up -d --build
```

> **⚠️ Multi-container networking:** the API and the agent call each other.
> - `AGENT_SERVICE_URL` (API → agent) is used by every AI feature
> - `AGENT_INTERNAL_CALLBACK_URL` (agent → API) is used by the collection tools of the chat
>
> In containers both **must** use Docker service names (`http://agent:8081`, `http://app:8080`), as `docker-compose.dev.yml` does. The agent is not published on a host port; every endpoint except `/health` requires `AGENT_INTERNAL_SERVICE_TOKEN`.

To build the published image locally: `docker build --target release -t orlanumis:local .`

### Upgrading a source build

```sh
git pull
docker compose -f docker-compose.dev.yml up -d --build
```

The migrations run when the new app starts. The automatic full backup of the night before is your way back; for a bigger upgrade, make one first with **Settings → Backups → Back up now**.

---

## Environment Variables

### Required `.env` values (source builds)

Copy `.env.example` to `.env` (or run `task init`) and replace the placeholders:

```env
JWT_SECRET=<48-byte random value>
AGENT_INTERNAL_SERVICE_TOKEN=<48-byte random value shared by app and agent>
TRUSTED_PROXIES=none
```

Generate a value with `openssl rand -base64 48`. The published image makes both secrets itself.

`AGENT_INTERNAL_SERVICE_TOKEN` is **not** an Anthropic or Ollama API key. It is a private credential shared by the Go API and the Python agent. If either process lacks it, coin analysis and agent chat fail with an internal agent service credential error.

For local `task up-all` development (no containers), the API and the agent both load the repo-root `.env`. Use localhost URLs:

```env
AGENT_SERVICE_URL=http://localhost:8081
AGENT_INTERNAL_CALLBACK_URL=http://localhost:8080
AGENT_TRUSTED_OUTBOUND_ORIGINS=http://localhost:8080
AGENT_ALLOW_LOCAL_OUTBOUND=true
```

### Full variable reference

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | — | JWT signing key (**at least 32 characters**) |
| `GIN_MODE` | debug (`release` in the published image) | Set `release` for an internet-facing deployment: it turns on the production startup checks |
| `DB_PATH` | `./ancientcoins.db` | SQLite database file |
| `PORT` | `8080` | HTTP port |
| `UPLOAD_DIR` | `./uploads` | Uploaded images |
| `BACKUP_DIR` | — | Folder of the full backups (`/app/backups` in the images) |
| `AUREARIA_BACKUP_DIR` | `~/Aurearia-backups/<compose project>` | Host folder mounted as `/app/backups` by `docker-compose.dev.yml` |
| `APP_PORT` | `8080` | Host port of the app in `docker-compose.dev.yml` (bound to `127.0.0.1`) |
| `CORS_ORIGINS` | localhost origins outside release mode | Comma-separated browser origins allowed to call the API |
| `TRUSTED_PROXIES` / `GIN_TRUSTED_PROXIES` | *(required in release mode)* | Proxy IPs/CIDRs whose `X-Forwarded-*` headers the API trusts. Use the exact reverse-proxy address, `none` only without a proxy, never `0.0.0.0/0`. |
| `AGENT_SERVICE_URL` | `http://agent:8081` | Python agent URL (API → agent) |
| `AGENT_INTERNAL_CALLBACK_URL` | `http://localhost:8080` | URL the agent uses to call the API for the chat's collection tools. **In containers, set it to `http://app:8080`.** |
| `AGENT_INTERNAL_SERVICE_TOKEN` | — | Shared API ↔ agent credential |
| `AGENT_LOG_LEVEL` | `INFO` | Agent log level |
| `AGENT_DEBUG` | `false` | Debug mode of the agent (exposes `/docs`) |
| `AGENT_SEARXNG_URL` | — | SearXNG URL for Ollama web search |
| `AGENT_TRUSTED_OUTBOUND_ORIGINS` | `http://app:8080` | Exact origins the agent may call (Ollama, SearXNG, the API) |
| `AGENT_ALLOW_LOCAL_OUTBOUND` | `false` | `true` only when those origins are local or private, as with Ollama on the same computer |

In the published image the agent settings point to the app on `127.0.0.1` by themselves; set `AGENT_TRUSTED_OUTBOUND_ORIGINS` to add Ollama or SearXNG (keep `http://127.0.0.1:8080` in the list).

> **⚠️ Important:** Never use a weak or default JWT secret: it compromises every token. Never expose `AGENT_INTERNAL_SERVICE_TOKEN` to browsers or third-party services.

---

## Docker Volumes (source builds)

| Mount | Purpose |
|---|---|
| `/app/data` | SQLite database — collection data, users and settings |
| `/app/uploads` | Uploaded images referenced by the database |
| `/app/backups` | Full backups — a host folder, so they survive `down -v` |

Losing the first two loses the collection. `docker-compose.dev.yml` uses named volumes for them. For bind mounts, create the host directories writable by UID/GID `10001:10001`:

```sh
sudo mkdir -p /opt/aurearia/data /opt/aurearia/uploads
sudo chown -R 10001:10001 /opt/aurearia/data /opt/aurearia/uploads
```

---

## Running on a Server

With the published image, bind the port to the host's loopback and set the proxy in `docker-compose.yml`:

```yaml
services:
  orlanumis:
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      # The address the proxy's requests arrive from: the gateway of the
      # Docker network (docker network inspect orlanumis_default).
      TRUSTED_PROXIES: 172.18.0.1/32
      CORS_ORIGINS: https://coins.example.com
```

Check it with **API sees your IP as**, as [Trusted Proxy and Client IP Handling](#trusted-proxy-and-client-ip-handling) explains.

With a source build, keep `docker-compose.dev.yml` as it is and add the production settings in an override file, for example `docker-compose.prod.yml`:

```yaml
services:
  app:
    environment:
      GIN_MODE: release
      TRUSTED_PROXIES: 127.0.0.1/32
      CORS_ORIGINS: https://coins.example.com
  agent:
    environment:
      AGENT_TRUSTED_OUTBOUND_ORIGINS: http://app:8080
      AGENT_ALLOW_LOCAL_OUTBOUND: "false"
```

```sh
docker compose -f docker-compose.dev.yml -f docker-compose.prod.yml up -d --build
```

The app stays bound to `127.0.0.1`; a reverse proxy on the same host serves it over HTTPS.

## Reverse Proxy Setup

The Go app serves HTTP on `PORT` (8080). Put a reverse proxy in front for TLS:

- **Proxy all traffic** to the app — it handles its own routing, including the SPA fallback
- **Re-set the forwarding headers** — `X-Forwarded-For`, `X-Forwarded-Proto` and `Host` — at the edge
- **Support SSE** — the AI chat, Coin Copilot and Deep Analysis stream Server-Sent Events; allow long-lived connections and turn buffering off

### nginx baseline for internet-facing deployments

```nginx
server {
    listen 80;
    server_name coins.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name coins.example.com;

    ssl_certificate     /etc/letsencrypt/live/coins.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/coins.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Content-Security-Policy "default-src 'self'; img-src 'self' data: blob:; media-src 'self'; connect-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; upgrade-insecure-requests" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header X-Frame-Options "DENY" always;
    add_header Permissions-Policy "camera=(self), microphone=(self), geolocation=(), payment=(), usb=(), serial=(), bluetooth=()" always;

    # Keep in step with the API's upload cap (10 MB multipart). Raise it if you
    # import full backups through the browser: a backup is one large zip.
    client_max_body_size 10m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        # Re-set forwarding headers at the trusted edge. Do not append user-supplied X-Forwarded-For.
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Port 443;
        proxy_set_header Forwarded "";

        # SSE streaming through the Go API.
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
```

### Caddy

Caddy gets certificates from Let's Encrypt by itself and handles SSE without extra settings:

```
coins.example.com {
    reverse_proxy localhost:8080
}
```

---

## Trusted Proxy and Client IP Handling

Internet-facing deployments must make the API derive the client IP only from trusted reverse proxies. Failed-sign-in events, the automatic IP block, account lockout and rate limits all depend on it.

```env
TRUSTED_PROXIES=127.0.0.1/32,172.18.0.0/16
```

Verification:

1. Browse from a device outside the server's network.
2. Open **Admin → Security → Public-Facing Readiness** and check **API sees your IP as**.
3. It must show your public IP, not the proxy's, a Docker bridge address, `127.0.0.1`, or a value you sent in `X-Forwarded-For`.
4. Send a request with a fake `X-Forwarded-For` from the internet and confirm the API still records the real IP.

Until this is verified, keep the app on the home network or behind a VPN.

---

## HTTPS

The Go app serves HTTP only; TLS ends at the reverse proxy. HTTPS is required for any deployment reached over the internet: sign-in sends passwords and tokens, and the installed app (PWA) and the camera need a secure context.

Options: **Caddy** (automatic, simplest), **nginx + certbot**, **Traefik**, or a cloud load balancer.

---

## First Launch

1. **Register the first account** from a trusted network: the first user becomes the **admin**.
2. **Registration** — after the first user, new accounts are refused while the `RegistrationMode` setting is `closed` (the default). Set it to `open` only while you are inviting someone, then close it again.
3. **AI** — **Admin → AI**: choose Anthropic (recommended) or Ollama.
4. **System** — **Admin → System**: the public app URL, the log level, the Numista API key and the collection currency.
5. **Start adding coins** — see the [Getting Started Guide](getting-started.md).

---

## Keep the Agent Service Private

The Python agent is an internal worker, not a public API. Traffic flows Browser → reverse proxy/TLS → Go API (`:8080`) → agent (`agent:8081`). In the published image the agent listens on the container's `127.0.0.1:8081` only, so it cannot be published, and its token is made on the first start; the points below are for source builds.

### Agent port privacy

- Do **not** add `ports: ["8081:8081"]` to the `agent` service: port 8081 is not published, only exposed on the Docker network.
- Set a high-entropy `AGENT_INTERNAL_SERVICE_TOKEN` in both containers; rotate it if logs or configuration leak.
- Keep `AGENT_DEBUG=false` so the FastAPI docs are not exposed.
- Firewall public hosts to `80/tcp`, `443/tcp` and restricted SSH/VPN. Block `8080/tcp` unless it is bound to `127.0.0.1`; always block `8081/tcp`.
- Keep `AGENT_TRUSTED_OUTBOUND_ORIGINS` narrow, and `AGENT_ALLOW_LOCAL_OUTBOUND=false` on an internet-facing host. (The published image needs it `true`, to reach the app on its own `127.0.0.1`; the trusted origins still limit what the agent calls.)

```sh
# From outside the host/network, these must fail or time out:
curl http://coins.example.com:8080/health
curl http://coins.example.com:8081/health
```

---

## Sign-in Protection and Audit Events

- **Account lockout** — 5 failed passwords in 15 minutes lock the account for 15 minutes. Admins unlock it in **Admin → Security → Locked Users** (`POST /admin/users/:id/unlock`).
- **IP block** — 20 failed passwords from one IP in 15 minutes block that IP until the failures are older than 15 minutes.
- **Rate limits** — register, login and refresh are rate limited.
- **Events** — `password_login_success`, `password_login_failure`, `refresh_failure`, `account_lockout`, `account_unlock` and `final_local_admin_blocked`, with username, client IP, user agent and time, in **Admin → Security → Security Events**.

Alert, at least, on failed-sign-in bursts, account lockouts, admin sign-ins from a new IP, `RegistrationMode` set to `open`, a missing nightly backup, and unexpected direct traffic to the agent.

---

## Backup & Restore

The database and the uploads must be backed up together: the database holds the records and the image references, `/app/uploads` the images they point to.

### Full backup (in the app)

**Settings → Backups → Full backup** (administrators) makes a zip of the whole database and every upload, with a `manifest.json`:

- **Every day** at the time set in **Admin → Schedules** (03:00 server time by default); the last 7 automatic backups are kept. A run missed while the computer slept is made when the app runs again.
- **Back up now**, **download**, **import** (a file that is not a backup of the app, or is damaged, is refused) and **delete**.
- **Restore** checks the backup, saves the data of now in a *pre-restore* backup, then restarts the app, which applies the backup as it starts; Docker starts the container again by itself.
- The backups are written to `BACKUP_DIR`, which `docker-compose.dev.yml` maps to a host folder, so they survive a deleted Docker volume.

A backup holds every user's data, password hashes and saved provider keys included: keep it safe.

### Backup policy

- Keep at least one copy off the host (cloud storage, a NAS, or an external disk not mounted all the time), encrypted.
- Keep enough history to recover from a deletion noticed days later.
- **Export ZIP** in **Settings → Backups** exports one user's collection and photos; it is useful for portability, but it is not a disaster-recovery backup.

### Cold backup (without the app)

```sh
docker compose -f docker-compose.dev.yml stop app
# Copy the database (ancientcoins.db plus any -wal/-shm files) and the uploads
# from the named volumes or bind mounts, together.
docker compose -f docker-compose.dev.yml start app
```

### Backup restore drill

Restore a backup on a separate stack (`-p` with another project name and `APP_PORT`) and check:

1. Sign in as admin.
2. Open several coins with images and confirm they render.
3. Run a small AI feature if the agent is configured.
4. Note the date and the result. Do this before relying on the backups, and then at least once a year.

---

## Admin-Managed Settings

These settings live in the database and are changed in the **Admin** pages — they are not environment variables.

| Setting | Default | Description |
|---|---|---|
| AIProvider | — | `anthropic` or `ollama` (required before the AI features work) |
| AnthropicAPIKey | — | API key for Claude models |
| AnthropicModel | — | Claude model |
| OllamaURL | `http://localhost:11434` | Ollama server URL |
| OllamaModel | `llava` | Vision model name |
| OllamaTimeout | `300` | Request timeout in seconds |
| SearXNGURL | — | SearXNG URL (required for Ollama web search) |
| NumistaAPIKey | — | Numista catalog API key |
| NumistaMonthlyLimit | `2000` | Numista requests a month; background requests pause from 90% |
| NumistaSearchTTLHours | `24` | Search-cache TTL, 1–720 hours |
| NumistaDetailTTLHours | `168` | Detail-cache TTL, 1–2160 hours |
| DeepIdentificationEnabled | `false` | Enables the optional Deep Analysis jobs |
| DeepIdentificationWorkerCount | `2` | Background workers, 1–32 |
| DeepIdentificationMaxActivePerUser | `1` | Concurrent jobs per user, 1–10 |
| DeepIdentificationHardTimeoutSeconds | `420` | Job hard timeout in seconds |
| DeepIdentificationNumistaCallBudget | `4` | Numista calls per job, 1–20 |
| CoinCategories / CoinEras | `Modern`, `Other` / `ancient`, `medieval`, `modern` | Options of the coin forms |
| LogLevel | — | Application log level |
| PublicAppURL | — | Public `https://` base URL used in links |
| RegistrationMode | `closed` | `open` lets new users register after the first; anything else refuses them |
| BackupEnabled / BackupStartTime / BackupKeep | `true` / `03:00` / `7` | The automatic full backup |

---

## Public Exposure Beta Acceptance Checklist

Use this gate before inviting anyone or pointing public DNS at the host.

### External nginx checks

- [ ] Only `80/tcp` and `443/tcp` are public; SSH is restricted by IP/VPN; `8080` is localhost-only; `8081` is not public.
- [ ] nginx/Caddy terminates TLS and redirects HTTP to HTTPS.
- [ ] TLS 1.2/1.3 only.
- [ ] HSTS, CSP, `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`/`frame-ancestors` and `Permissions-Policy` are present.
- [ ] `client_max_body_size` matches the API upload cap.
- [ ] SSE buffering is off for proxied app traffic.
- [ ] `X-Forwarded-For` and `X-Forwarded-Proto` are stripped and re-set at the edge; the API sees the real client IP.

### HTTPS and registration

- [ ] The first admin account was created from a trusted network.
- [ ] `RegistrationMode` is `closed`, and only opened briefly to invite someone.
- [ ] `GIN_MODE=release`, `TRUSTED_PROXIES` and `CORS_ORIGINS` are set for the real domain, served over HTTPS.
- [ ] Sign-in, sign-out and token refresh work after TLS is active.

### Private media

- [ ] **Private media**: authenticated `/uploads/*` requests require a valid token, and another user's coin images and avatar are not served.
- [ ] Signing out of the installed app clears the cached private media on shared devices.

### Operations

- [ ] **Agent port privacy**: the agent is private (`expose`, no host `ports` for 8081), not published, and requires `AGENT_INTERNAL_SERVICE_TOKEN`.
- [ ] The automatic full backup is on, and a copy goes off the host.
- [ ] A **Backup restore drill** has succeeded (SQLite plus uploads).
- [ ] Alerts exist for failed-sign-in bursts, lockouts, admin sign-ins, direct agent traffic and backup failures.
- [ ] The security events are reviewed daily for the first week.

---

## Troubleshooting

### Database locked errors
- Run only one app container against a database file
- SQLite is not designed for concurrent writes from several processes

### AI analysis not responding
- Check that the Ollama server is reachable from the container (`host.docker.internal` when Ollama runs on the same computer; `docker-compose.dev.yml` allows it)
- Verify the model is pulled (`ollama pull llava`)
- Increase `OllamaTimeout` in Admin for slower hardware

### AI agent not responding
- Published image: check that `AI_AGENT` is not `off`, and look for the agent's lines in `docker logs orlanumis`
- Source build: check that the agent container is running: `docker compose -f docker-compose.dev.yml logs agent`
- Verify `AGENT_SERVICE_URL` points to the agent container

### AI provider not configured
- Set the provider in **Admin → AI**. The agent chat shows a configuration banner while it is empty.

### Container won't start
- Check the logs: `docker logs orlanumis` (published image) or `docker compose -f docker-compose.dev.yml logs -f`
- Published image: the three folders must be writable by `PUID:PGID`; a line `could not make … belong to` names one that is not
- Source build: verify `.env` exists with `JWT_SECRET` (at least 32 characters) and `AGENT_INTERNAL_SERVICE_TOKEN`
- A failed migration step is named in the log; restore the last full backup if needed
