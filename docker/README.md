# Orla Numis

A self-hosted catalogue for a personal collection of coins and banknotes: records with photos and catalogue references (KM, Pick, Numista), Numista lookups and values, sets and challenges such as the euro commemoratives, statistics, a wish list, duplicates to trade, PDF and Excel exports, a full backup every day, and optional AI tools (Anthropic Claude or a self-hosted Ollama). In English and European Portuguese.

One image with everything: the app and its AI agent, for `linux/amd64` and `linux/arm64`.

Source, documentation and issues: [github.com/AnaGermano6/OrlaNumis](https://github.com/AnaGermano6/OrlaNumis).

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/C0H729CST9)

## Start

```sh
docker run -d --name orlanumis --restart unless-stopped \
  -p 8080:8080 \
  -e PUID=1000 -e PGID=1000 -e TZ=Europe/Lisbon \
  -v ./data:/app/data \
  -v ./uploads:/app/uploads \
  -v ./backups:/app/backups \
  anagermano/orlanumis:latest
```

Open `http://<this computer>:8080` and register: the first account becomes the administrator. Numista and the AI providers are set up in **Admin**.

With Docker Compose, use the project's [`docker-compose.yml`](https://github.com/AnaGermano6/OrlaNumis/blob/main/docker-compose.yml): `docker compose up -d` in its folder.

On **Unraid**, install **OrlaNumis** from Community Applications.

## Folders

| Folder | What it holds |
|---|---|
| `/app/data` | The database, the settings and the secrets |
| `/app/uploads` | The photos of the collection |
| `/app/backups` | The full backups, made every day (Settings → Backups) |

Keep the first two together: the database points to the photos.

## Environment

| Variable | Default | What it does |
|---|---|---|
| `PUID`, `PGID` | `10001` | User and group that run the app and own the folders |
| `TZ` | UTC | Time zone, for the daily backup and the schedules |
| `AI_AGENT` | `on` | `off` leaves the AI agent out, and with it the AI features |
| `TRUSTED_PROXIES` | `none` | Behind a reverse proxy, its address (`172.18.0.1/32`), so sign-in protection sees the real client |
| `CORS_ORIGINS` | — | Other browser origins allowed to call the API, if any |
| `JWT_SECRET`, `AGENT_INTERNAL_SERVICE_TOKEN` | made on the first start | Kept in `/app/data/.orlanumis-secrets`; set them only to choose your own |

Before reaching it from the internet, read the [deployment guide](https://github.com/AnaGermano6/OrlaNumis/blob/main/docs/deployment.md): HTTPS through a reverse proxy, trusted proxies, closed registration and backups.

## Updating

**Admin → Updates** says when a new version is out and what it brings. Updating keeps the data: the app migrates it when it starts on the new version.

- Docker Compose: `docker compose pull && docker compose up -d`
- Unraid: Docker tab, **Check for Updates**, then **Apply Update**
- `docker run`: pull the new image, then remove the container and create it again with the same options

With [Watchtower](https://github.com/nicholas-fedor/watchtower) set up (its HTTP API on), the app updates itself with one click, after a backup. The `docker-compose.yml` has it ready.

## Tags

| Tag | |
|---|---|
| `latest` | The latest version |
| `1.2.3` | That version |
| `1.2` | The latest `1.2.x` |

## License

MIT. Based on [Aurearia](https://github.com/briandenicola/Aurearia) by Brian DeNicola.
