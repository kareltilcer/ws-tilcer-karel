# karel — Personal Site & CMS

Karel's bilingual (CZ/EN) personal-brand showcase + a small self-hosted CMS, on
**`karel.tilcer.cz`** (`kaja.tilcer.cz` 301-redirects to it). A wow/summer/
ice-cream public site (projects, linktree, about, skills, arcade, contact, legal)
plus a quiet, function-first admin CMS at `/admin` behind Mode B auth.

Two Coolify apps, one origin — mirrors the `home`/`fin`/`status` fleet.

- **Backend** — Go modular monolith (chi + embedded SQLite + Goose), port **2002**.
- **Frontend** — React + TypeScript + Vite SPA (static Nginx image), port **2001**.

## Layout

```
backend/    Go API (cmd/karel, internal/{platform,modules,bootstrap})
frontend/   React SPA (src/{api,i18n,lib,components,pages,admin})
docker-compose.yml  docker-entrypoint.sh  litestream.yml
handoff/ design/     specs + delivered design (frontend/design-ref mirrors design/v2)
```

Backend modules over shared platform infra (`config`, `db`, `httpx`, `auth`,
`registry`, `spam`, `ratelimit`, `email`, `media`):

- **content** — projects, categories, links (+click), page sections, skills,
  business-info; owns the shared `media` table.
- **media** — S3 upload/list/delete with mime+size validation, **SVG sanitize**,
  and a reference-guarded delete.
- **contact** — public submit (honeypot + Turnstile + rate-limit + body cap →
  store-first then Resend email; email failure is non-fatal) + admin inbox.
- **arcade** — code-defined game registry; public leaderboard + spam-guarded
  score submit; admin moderation.

## Run locally

### Full stack via Docker (offline harness)

```bash
docker compose up --build
```

Opens the SPA on <http://localhost:2001> (Nginx proxies `/api` to the backend),
backend on <http://localhost:2002>. Runs with the **dev auth bypass** (you are a
fake admin — no login) and **no R2 replication**. This is a smoke test of the
images + entrypoint, not a real deploy.

### Dev servers (hot reload)

```bash
# backend (terminal 1) — dev bypass, local SQLite file
cd backend
KAREL_ENV=development KAREL_DEV_AUTH_BYPASS=true KAREL_ADDR=:2002 KAREL_DB_PATH=./karel.db \
  go run ./cmd/karel

# frontend (terminal 2) — Vite proxies /api to :2002
cd frontend
npm install
npm run dev
```

Open the printed Vite URL. The public site is at `/`, the CMS at `/admin`.

### Tests

```bash
cd backend && go test ./...          # unit + integration (in-memory SQLite)
cd frontend && npm run lint          # tsc typecheck
cd frontend && npm run build         # production build (admin is a lazy chunk)
```

## Configuration (env — set in Coolify; never commit secrets)

**Core / auth (Mode B, mirror `home`):** `KAREL_ENV`, `KAREL_ADDR=:2002`,
`KAREL_DB_PATH`, `KAREL_SITE_KEY=karel`, `AUTH_BASE_URL`,
`KAREL_AUTH_SERVICE_SECRET`, `KAREL_AUTH_JWT_SECRET`, `KAREL_AUTH_JWT_ISSUER`
(optional), `KAREL_ALLOWED_ORIGINS`, `KAREL_SESSION_TTL_DAYS` (90),
`KAREL_ROLE_REFRESH_MINUTES` (15). Dev only: `KAREL_DEV_AUTH_BYPASS=true`
(refused when `KAREL_ENV=production`).

**Contact email (Resend):** `RESEND_API_KEY`, `CONTACT_TO`, `CONTACT_FROM`.

**Spam / anti-abuse:** `TURNSTILE_SECRET` (empty = verification disabled, honeypot
still enforced) + frontend build arg `VITE_TURNSTILE_SITE_KEY`; `CONTACT_RATE`
(`5/min`), `SCORE_RATE` (`20/min`), `CLICK_RATE` (`60/min`), `MAX_CONTACT_BYTES`,
`MAX_MEDIA_BYTES`.

**Media S3 (Cloudflare R2):** `S3_ENDPOINT` (R2 API host
`https://<account-id>.r2.cloudflarestorage.com`), `S3_REGION=auto`, `S3_BUCKET`,
`S3_ACCESS_KEY`, `S3_SECRET_KEY` (an R2 API token), `S3_FORCE_PATH_STYLE=true`,
`S3_MEDIA_PREFIX` (`media/`), `MEDIA_PUBLIC_BASE_URL` (the bucket's **public** read
host — a custom domain or `pub-*.r2.dev`, *not* the R2 API host, which needs signed
requests). Uploads return **503** until configured.

**DB backup (Litestream → R2):** `LITESTREAM_ENABLED=true`,
`LITESTREAM_R2_ENDPOINT`, `LITESTREAM_R2_BUCKET`, `LITESTREAM_ACCESS_KEY_ID`,
`LITESTREAM_SECRET_ACCESS_KEY` (DB stored under prefix `karel`).

## Deploy (Coolify, two apps)

1. **Backend app** — Build Pack Dockerfile, Base Directory `/`, Dockerfile
   `/backend/Dockerfile`; port **2002**; persistent volume at `/data`; set all
   `KAREL_*` + secret env. Route `karel.tilcer.cz/api`, `/healthz`, `/readyz` →
   this app. **Do NOT enable Strip Prefix** (routes are served under `/api`).
2. **Frontend app** — Build Pack Dockerfile, Base Directory `/frontend`,
   Dockerfile `/frontend/Dockerfile`; port **2001**; build arg
   `VITE_TURNSTILE_SITE_KEY`. Route everything else on `karel.tilcer.cz` → this app.
3. **Edge 301** — add a Traefik host rule redirecting `kaja.tilcer.cz` → the same
   path on `https://karel.tilcer.cz` (preferred over app-level).
4. `/readyz` is the recommended `monitor_url` for the `status` service.

## Backups

**Database (automatic).** The backend runs under `litestream replicate -exec` (see
`docker-entrypoint.sh`); every write streams to R2 under prefix `karel`. On a fresh
volume the entrypoint restores the DB from R2 first (`-if-db-not-exists
-if-replica-exists`), so a rebuild comes back with its data and the initial
migration never double-seeds. **Fresh-build restore drill:** delete the `/data`
volume and redeploy — the DB should restore from R2 and the site serve prior data.

**Media (not backed up).** Media lives in **Cloudflare R2** and relies on R2's own
durability — there is **no separate media backup**, and it is not in the DB backup
(the DB holds only object keys/URLs). Consequence: if the R2 media bucket is ever
lost, the DB's stored `public_url`s stop resolving and the images must be
re-uploaded. This is an accepted risk — media is non-critical relative to the DB.

## Ops prerequisites (before/with deploy)

Provision the Cloudflare R2 media bucket (+ public access via a custom domain);
create the
Cloudflare Turnstile site + secret keys; set up the Resend sender domain for
`CONTACT_FROM`; enter the real OSVČ business-disclosure values in the CMS after
first deploy (the seed leaves them blank).

## Deferred (v1 non-goals)

No blog (when it lands, migrate the public site to full SSR for article SEO); the
two remaining arcade games (registry is extensible — add keys + score bounds in
`backend/internal/modules/arcade/games.go`); no BE→BE inbound surface.
