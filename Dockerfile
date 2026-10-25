# Orla Numis images. The last stage, "release", is the published image: the
# app and its AI agent in one container (docker build . builds it). The
# "app" stage is the app alone, for the two-container stacks of
# docker-compose.dev.yml and docker-compose.exploration.yml.

# Build Vue frontend (the same files on every architecture)
FROM --platform=$BUILDPLATFORM node:24-alpine@sha256:156b55f92e98ccd5ef49578a8cea0df4679826564bad1c9d4ef04462b9f0ded6 AS web-build
ARG APP_VERSION=dev
ARG BUILD_DATE=""
WORKDIR /web
COPY src/web/package.json src/web/package-lock.json* ./
RUN npm ci
COPY VERSION /VERSION
COPY src/web/ .
ENV VITE_API_BASE_URL=""
ENV VITE_APP_VERSION=${APP_VERSION}
ENV VITE_BUILD_DATE=${BUILD_DATE}
RUN npm run prepare-background-removal-assets
RUN npm run build

# Build Go API, cross-compiled for the target architecture
FROM --platform=$BUILDPLATFORM golang:1.26.6-alpine@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df AS api-build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY src/api/go.mod src/api/go.sum ./
RUN go mod download
COPY src/api/ .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /app/ancient-coins-api .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /app/exploration-seed ./cmd/exploration-seed

# Test-only Feature 360 seed image. The seed binary is intentionally absent
# from the normal application image below.
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d AS exploration-seed
RUN addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app
WORKDIR /app
COPY --from=api-build --chown=app:app /app/exploration-seed .
RUN mkdir -p /app/data && chown -R app:app /app
USER 10001:10001
ENTRYPOINT ["./exploration-seed"]

# The app alone
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d AS app
RUN apk upgrade --no-cache \
    && apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app
WORKDIR /app
COPY --from=api-build --chown=app:app /app/ancient-coins-api .
COPY --from=web-build --chown=app:app /web/dist ./wwwroot
COPY --chown=app:app VERSION ./VERSION
RUN mkdir -p /app/uploads /app/data \
    && chown -R app:app /app
VOLUME ["/app/uploads", "/app/data"]
ENV PORT=8080
ENV DB_PATH=/app/data/ancientcoins.db
EXPOSE 8080
USER 10001:10001
ENTRYPOINT ["./ancient-coins-api"]

# The AI agent's locked dependencies, for the target architecture
# (built where it will run, as a venv's scripts name its path).
FROM python:3.12-slim@sha256:d764629ce0ddd8c71fd371e9901efb324a95789d2315a47db7e4d27e78f1b0e9 AS agent-build
WORKDIR /app/agent
COPY src/agent/pyproject.toml src/agent/uv.lock ./
RUN pip install --no-cache-dir uv==0.11.22 \
    && uv sync --locked --no-dev --no-install-project

# The published image: the app and its AI agent in one container. It starts
# as root only to hand the data folders to PUID:PGID (see docker/entrypoint.sh).
FROM python:3.12-slim@sha256:d764629ce0ddd8c71fd371e9901efb324a95789d2315a47db7e4d27e78f1b0e9 AS release
ARG APP_VERSION=dev
# Strip system pip (PYSEC-2026-3721): the agent's venv never needs it at runtime.
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends ca-certificates tini \
    && rm -rf /var/lib/apt/lists/* \
    && python -m pip uninstall -y pip
WORKDIR /app
COPY --from=agent-build /app/agent/.venv /app/agent/.venv
COPY src/agent/app/ /app/agent/app/
COPY --from=api-build /app/ancient-coins-api ./
COPY --from=web-build /web/dist ./wwwroot
COPY VERSION ./VERSION
COPY docker/entrypoint.sh /usr/local/bin/orlanumis
RUN chmod 0755 /usr/local/bin/orlanumis \
    && mkdir -p /app/data /app/uploads /app/backups
ENV PORT=8080 \
    DB_PATH=/app/data/ancientcoins.db \
    UPLOAD_DIR=/app/uploads \
    BACKUP_DIR=/app/backups \
    GIN_MODE=release \
    TRUSTED_PROXIES=none \
    AGENT_SERVICE_URL=http://127.0.0.1:8081 \
    PUID=10001 \
    PGID=10001 \
    PYTHONUNBUFFERED=1
VOLUME ["/app/data", "/app/uploads", "/app/backups"]
EXPOSE 8080
# The health check asks /healthz with bash alone: starting Python for it
# every 30 seconds cost a small server more than the idle app itself.
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD ["bash", "-c", "exec 3<>/dev/tcp/127.0.0.1/${PORT:-8080} && printf 'GET /healthz HTTP/1.0\\r\\n\\r\\n' >&3 && head -n 1 <&3 | grep -q ' 200 '"]
LABEL org.opencontainers.image.title="Orla Numis" \
      org.opencontainers.image.description="A self-hosted app for cataloguing a personal coin and banknote collection" \
      org.opencontainers.image.url="https://github.com/AnaGermano6/OrlaNumis" \
      org.opencontainers.image.source="https://github.com/AnaGermano6/OrlaNumis" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${APP_VERSION}"
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/orlanumis"]
