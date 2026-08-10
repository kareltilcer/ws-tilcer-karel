# syntax=docker/dockerfile:1
#
# API-only image for the `karel` backend: the Go JSON API, with Litestream
# streaming the SQLite database to Cloudflare R2. The public SPA ships as its own
# Nginx image (see frontend/Dockerfile); this backend serves no static assets.
#
# The build context is the REPOSITORY ROOT (not backend/), because the image
# needs backend/, litestream.yml, and docker-entrypoint.sh. In Coolify set:
#   Build Pack          = Dockerfile
#   Base Directory      = /            (context root)
#   Dockerfile Location = /backend/Dockerfile
# The backend listens on :2002 (KAREL_ADDR) — point Coolify's port at 2002.

# ── Stage 1: build the Go binary ───────────────────────────────────────
FROM golang:1.26-alpine AS backend
RUN apk add --no-cache git
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=readonly
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Pure-Go SQLite (modernc, CGO off). Cap compiler parallelism (-p 1 / single
# GOMAXPROCS) so the memory-hungry modernc.org/libc + sqlite compile doesn't get
# OOM-killed on small build hosts, especially since Coolify builds --no-cache
# with the frontend building in parallel. Trades a little build time for a much
# lower peak RSS.
RUN GOMAXPROCS=1 go build -p 1 -trimpath -ldflags="-s -w" -o /out/karel ./cmd/karel

# ── Stage 2: pull the Litestream binary ────────────────────────────────
FROM litestream/litestream:0.3.13 AS litestream

# ── Stage 3: runtime ───────────────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream
COPY --from=backend    /out/karel                /usr/local/bin/karel
COPY litestream.yml       /etc/litestream.yml
COPY docker-entrypoint.sh /usr/local/bin/entrypoint
RUN chmod +x /usr/local/bin/entrypoint

# Defaults; secrets and per-service config come from Coolify env at runtime.
# API-only: KAREL_STATIC_DIR is intentionally unset (the SPA is the frontend image).
ENV KAREL_ADDR=:2002 \
    KAREL_DB_PATH=/data/karel.db
EXPOSE 2002
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/entrypoint"]
