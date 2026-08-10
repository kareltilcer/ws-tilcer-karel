# Engineering Handoff — karel (Personal Site & CMS)

> For: **Claude Code** · Owner: Karel · Last updated: 2026-08-08
> **Source of truth:** `PRD.md` (behaviour, data model, acceptance criteria) and `openapi.yaml` (the API contract — implement it exactly). `HANDOFF-design.md` + the delivered `*.dc.html` prototypes drive the frontend look. `../../CLAUDE.md` conventions are inherited and authoritative. This doc is the build plan, not a re-spec.

## 1. Scope

Build the `karel` service: a Go **modular monolith** backend (three modules — `content`, `contact`, `arcade` — over shared infra) plus a bilingual (CZ/EN) React SPA, deployed as a two-app pair on `karel.tilcer.cz`. The public site is a wow/summer/ice-cream showcase (projects, linktree, About, Skills, arcade, contact, legal); the CMS is a quiet, function-first admin area at `/admin` behind Mode B auth. `kaja.tilcer.cz` 301-redirects to the canonical host.

## 2. Non-negotiable conventions (from CLAUDE.md + REGISTRY)

- **Deploy:** two Coolify apps, one origin (mirror `home`/`fin`/`status`). Backend = API-only Go image mapped to `karel.tilcer.cz/api`; frontend = static Nginx SPA on the catch-all domain. Traefik path-routes. **Do not enable Strip Prefix** on the backend — routes are served under `/api` or they 404.
- **Ports:** backend **2002**, frontend **2001**.
- **Canonical host:** `karel.tilcer.cz`. **`kaja.tilcer.cz` → 301** to the same path on the canonical host, **at the edge** (Traefik/Coolify host rule) — preferred over app-level so the SPA/API only ever run on one host. `<link rel="canonical">` + `og:url` always point at `karel.tilcer.cz`.
- **Backend stack:** Go **1.26**, `chi` router, `modernc.org/sqlite` (embedded, cgo-free), **Goose** migrations.
- **Backups:** **Litestream → Cloudflare R2**, prefix **`karel/`** (DB only). **Media is not backed up** — it relies on R2 durability (see §5 / §6). Fresh builds restore the DB from R2.
- **Frontend stack:** React + TypeScript + **Vite**, **TanStack Query** (`useQuery`), static Nginx image. Bilingual CZ/EN.
- **Auth:** shared `auth` backend, site id **`karel`**, **Mode B** (self-hosted login + own session cookie), mirroring `home`/`status`. JWT 15-min for API + session cookie for long-term.
- **Secrets/config:** Coolify env vars only. No secrets — and no real business-disclosure PII — in the repo.
- **Observability baseline:** `GET /healthz`, `GET /readyz` (with SQLite check), structured JSON logs to stdout, per-request logging (method, path, status, latency).
- **Repo:** `ws-tilcer-karel`.

## 3. Suggested layout

```
cmd/server/            main: config, wiring, graceful shutdown
internal/
  httpx/               chi middleware: request log, auth (JWT+session), rate limit, body-cap, recover, CORS(none-needed/same-origin)
  auth/                Mode B: self-hosted login, own session, auth-backend introspection + cache (mirror home)
  store/               sqlite open, PRAGMA foreign_keys=ON, Goose migration runner
  content/             projects, categories, links(+click), page_sections, skills, business_info; public read + admin CRUD
  media/               S3-compatible client (upload/delete), mime+size validation, SVG sanitize, reference-guard
  contact/             public submit (honeypot+Turnstile+rate+cap) → store + Resend email; admin inbox
  arcade/              code-defined game registry (keys + score bounds); public leaderboard + submit; admin moderation
  spam/                Turnstile server-side verify + honeypot check (shared by contact + arcade)
  i18n/                paired *_cs/*_en helpers, missing-translation flags
migrations/            Goose SQL (initial: all tables + indexes + seeds)
openapi.yaml           the contract (kept in sync with the build)
web/                   React + TS + Vite SPA (public site + lazy /admin)
  design-ref/          the delivered *.dc.html prototypes (visual reference; see §7)
README.md              run/deploy notes
```

## 4. Build order (milestones)

- **M0 — Scaffolding.** Repo, Go module, chi, sqlite+Goose, Litestream config (R2 `karel/`), health probes, structured logging + request-log middleware, Coolify two-app deploy skeleton on ports 2002/2001, edge 301 `kaja→karel`. Mode B auth wired (login, own session, introspection cache — mirror `home`). Clients configured: S3 (media), Resend (email), Turnstile (verify).
- **M1 — Content module.** Migrations for all tables (§5 of PRD). Public reads (`GET /api/projects`, `/projects/{slug}`, `/categories`, `/links`, `/sections/{key}`, `/skills`, `/business-info`) returning **published/visible only**, both languages. Admin CRUD + reorder: projects (draft/publish, category FK, gallery, out-links), categories (**RESTRICT → 409 delete-in-use**), links (+ public click counter FR-6), sections (about/hero), skills, business-info upsert.
- **M2 — Media module.** S3 upload (mime/size caps; **SVG sanitized** before store), list, delete (**409 if referenced**); wire `cover_media_id`, `project_media` gallery join, `page_section.media_id`; serve public URLs via `MEDIA_PUBLIC_BASE_URL` (never proxy bytes through the API).
- **M3 — Contact module.** `POST /api/contact`: honeypot + **Turnstile verify** + per-IP rate limit + body cap → store `contact_message` → email `CONTACT_TO` via Resend (reply-to = sender). **Email failure is non-fatal** (still `202`, logged). Admin inbox: list (cursor + status filter), triage `PATCH`, delete.
- **M4 — Arcade module.** Code **game registry** (each game: `key`, human name, `max_score` bound, optional per-game validation). `POST /api/arcade/{gameKey}/scores`: validate key + score bound + honeypot + Turnstile + rate limit → insert → return `{rank, top[]}`. `GET …/scores` top-N. Admin: delete score, reset board.
- **M5 — Frontend.** Build the SPA from the delivered design (`HANDOFF-design.md` + `*.dc.html`): public screens (hero, projects grid+filter+detail, linktree, about, skills, arcade+minigames, contact, legal, 404) + **lazy `/admin`** CMS + Mode B login. Bilingual instant switch, `prefers-reduced-motion` honored, per-route meta/OG + `sitemap.xml` + `robots.txt` + canonical. TanStack Query keys/invalidation per PRD §7.4.
- **M6 — Backups & ops hardening.** Verify Litestream fresh-build restore (DB). **Media is not backed up** — it relies on R2 durability; no media mirror or restore drill (PRD §8/§11).

## 5. Implementation notes & gotchas

- **Auth (Mode B):** admin endpoints require `bearerAuth` (JWT) with `sessionCookie` fallback; public read + the three public writes are open. Reuse `home`'s Mode B session handling + auth-backend introspection cache; site id `karel`. Login is self-hosted (email+password → own session); MFA is delegated (redirect/notice to `auth.tilcer.cz`). Do **not** build signup/reset/TOTP/Google.
- **DB:** `PRAGMA foreign_keys=ON`. FK behaviour matters: `project.category_id` **`ON DELETE RESTRICT`** (block category delete while referenced → **409**, don't cascade); `project_media` **CASCADE**; `cover_media_id` / `page_section.media_id` **`ON DELETE SET NULL`**. All timestamps RFC3339 **UTC**. Initial migration seeds the `business_info` singleton and empty `page_section` rows (`about`, `hero`); **no categories seeded** (Karel adds them).
- **Bilingual contract:** every translatable field is a `*_cs`/`*_en` pair, **both returned on public reads** (frontend switches with no refetch). Admin PATCH/PUT accept either/both; surface a **missing-translation flag** to the CMS (e.g. computed booleans) — the public site falls back to the other language.
- **Media / S3:** use an S3-compatible client (`aws-sdk-go-v2`) against **Cloudflare R2** with **path-style** addressing (`S3_FORCE_PATH_STYLE=true`; `S3_REGION=auto`). Validate mime + size (`MAX_MEDIA_BYTES`) → 413/415. **SVG must be sanitized** (strip `<script>`, `on*` handlers, external refs / `xlink:href` to remote, `<foreignObject>`) before storage → 422 on failure; serve SVG with restrictive `Content-Security-Policy` + `Content-Disposition`. Store only metadata (`s3_key`, `public_url`, mime, dims, size, alt). Delete removes the S3 object + row, but **block if referenced** (project cover/gallery or section) → 409.
- **Markdown:** project/section bodies are Markdown; **sanitize rendered HTML** (bluemonday-style allowlist; no scripts/iframes) — either render+sanitize server-side or sanitize on the client with a vetted lib. Never trust stored Markdown as safe HTML.
- **Spam guard (shared):** both `POST /api/contact` and `POST /api/arcade/{gameKey}/scores` require an empty **honeypot** field + a valid **Turnstile** token (server-side `siteverify` with `TURNSTILE_SECRET`) + per-IP rate limit. Honeypot filled or Turnstile invalid → **400**; over-rate → **429** (with `Retry-After`). Enforce body caps before decode.
- **Contact email:** store first, email second. A Resend failure must **not** lose the message or fail the request — return `202`, log the failure, flag the row for retry/attention. Reply-to = submitter's email.
- **Arcade games in code:** the game roster + score bounds live in a Go registry, not the DB. An unknown `gameKey` → 404; an out-of-bound/implausible score → 422. (The delivered design ships a 4-game roster with 2 playable — Catch-the-scoop, Scoop Match; implement those two's keys/bounds now, leave registry extensible.)
- **Link click counter:** `POST /api/links/{id}/click` is best-effort — increment `click_count`, per-IP rate-limited, **never blocks** the client's navigation; unknown id → 404 (ignored client-side). No analytics dashboard.
- **Reorder endpoints:** accept an ordered id list and rewrite `sort_order` **transactionally** (projects, categories, links, skills).
- **Public visibility:** public reads must filter `status='published'` / `visible=1`; drafts/hidden never leak. Admin list endpoints return everything (incl. drafts/hidden).
- **Pagination:** opaque cursor (`limit` 1–200, default 50) on admin `projects`, `contact`, `media` lists per `openapi.yaml`.
- **Errors:** uniform `{ "error": string, "detail"?: string }` (`Error` schema) across all failures.
- **Backups — media:** DB is Litestream→R2 (`karel/`). **Media is not backed up** — it lives in Cloudflare R2 and relies on R2's own durability (the DB holds only keys/URLs). If the R2 media bucket is lost, stored `public_url`s stop resolving and images must be re-uploaded — an accepted risk (media is non-critical relative to the DB).
- **SEO (v1):** client-side per-route `<title>`/OG/Twitter meta + `sitemap.xml` + `robots.txt` + canonical to `karel.tilcer.cz`. No SSR in v1 — **flagged to revisit as full SSR when the blog module lands** (PRD §2/§7.5). Ship the OG social-card per the design's `OG-Card.dc.html` direction.
- **Motion safety:** single full-experience register (calm mode was removed). Honor `prefers-reduced-motion`: still ambient CSS loops + the Home hero physics/sprinkles; keep functional motion (loaders, active gameplay). No content may be gated behind a game/animation.

## 6. Config (env — values in Coolify)

`AUTH_BASE_URL`, `AUTH_SITE_ID=karel`, `SESSION_SECRET`; `RESEND_API_KEY`, `CONTACT_TO`, `CONTACT_FROM`; `TURNSTILE_SECRET` (+ frontend Turnstile site key); `CONTACT_RATE` (e.g. `5/min`/IP), `SCORE_RATE`, `CLICK_RATE`, `MAX_CONTACT_BYTES`, `MAX_MEDIA_BYTES`; media S3 — `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_FORCE_PATH_STYLE=true`, `MEDIA_PUBLIC_BASE_URL`; DB backup — `LITESTREAM_*` + R2 creds (prefix `karel/`); no media backup (media relies on R2 durability); `PORT=2002` (frontend app `2001`). No BE→BE `X-Service-Secret` in v1.

## 7. Frontend integration (from the delivered design)

- The delivered prototypes are **visual references**, not production code — the DC-runtime has quirks (e.g. no `componentDidUpdate`) that don't apply to the real Vite/React build. Rebuild each screen as real React + TS + TanStack Query against `openapi.yaml` shapes.
- Screen → file map (delivered): `Home.dc.html` (hero), `Projects.dc.html` (grid+filter+detail), `Pages.dc.html` (About/Skills/Linktree/Contact/Legal), `Arcade.dc.html` (hub + 2 playable games + leaderboard + submit), `Admin.dc.html` (CMS + Mode B login), `NotFound.dc.html` (404), `OG-Card.dc.html` (social-card direction), `DesignSystem.dc.html` (tokens, type, motion, component inventory).
- **The design files currently live in `Downloads/Empty repo clarification needed/` — commit them into the repo** (suggest `web/design-ref/`) so the build has them.
- Extract tokens from `DesignSystem.dc.html` (day + night themes, Czech-safe type scale, motion vocab). Keep user-facing strings centralized (CZ + EN) in one module. Verify diacritics at display size and the three Czech plural forms for count labels.
- TanStack Query keys/invalidation per PRD §7.4; `401` on admin routes → redirect to login. On-theme empty/loading/error states.

## 8. Testing & acceptance

Map tests to **PRD §11 acceptance criteria**. Minimum:

- **Unit:** Turnstile verify + honeypot gate; rate-limit buckets (contact/score/click); body-cap enforcement; SVG sanitizer (rejects script/handler/remote-ref payloads); Markdown sanitizer; reorder transaction; missing-translation flagging; arcade score-bound validation.
- **Integration:** project/category/link/skill/section CRUD incl. duplicate slug (409) and validation (422); **category delete-in-use → 409**; **media delete-referenced → 409**; media upload happy path + 413/415/422(SVG); public reads hide drafts/hidden and return both languages; contact submit happy path (202 + stored + email attempted), honeypot/Turnstile (400), over-rate (429); email-failure still 202 + logged; arcade submit (201 + rank), unknown game (404), bad score (422); link-click increments and never blocks.
- **Auth:** admin endpoints reject unauthenticated (401); Mode B login/session/introspection cache behaves like `home`.
- **Contract:** responses conform to `openapi.yaml` (generate types / validate against it).
- **Backups:** DB fresh-build restore from R2 (`karel/`); **media is not backed up** (relies on R2 durability) — no media restore drill.
- **Frontend:** bilingual instant switch (no refetch); `prefers-reduced-motion` stills ambient motion/physics; per-route meta/OG + sitemap/robots present; 401→login.

## 9. Definition of done

All PRD §11 acceptance criteria pass; deployed as two Coolify apps (BE **2002** / FE **2001**) on `karel.tilcer.cz` with `/api` routing and **no strip-prefix**; `kaja.tilcer.cz` 301s to canonical; Mode B auth against `auth` site `karel`; Litestream replicating to R2 `karel/` with a verified fresh-build restore; **media stored in Cloudflare R2 with no separate backup** (relies on R2 durability); media served from `MEDIA_PUBLIC_BASE_URL`; contact emails via Resend with Turnstile+honeypot+rate-limit; arcade leaderboards with code-defined games; `/healthz` + `/readyz` per baseline; bilingual CZ/EN throughout with `prefers-reduced-motion` honored; `openapi.yaml` matches the built surface; REGISTRY row moved from "spec + design done, pre-build" to implemented.

## 10. Deferred / notes (defaults set — safe to proceed)

- **Blog** deferred to a future version; **when it lands, migrate the public site to full SSR** for article SEO/unfurls (PRD §2/§7.5). Leave IA room; build nothing now.
- **Remaining minigames** beyond the 2 playable — fill the code registry later; the API contract is unchanged.
- **Real business-disclosure values** (IČO, sídlo, živnostenský-úřad wording) entered by Karel in the CMS post-deploy; the seed leaves them blank.
- **Ops prerequisites** (do before/with M0/M6): provision the Cloudflare R2 media bucket (+ public access via a custom domain; no media backup), create the Turnstile keys, set up the Resend sender domain for `CONTACT_FROM`.
- **`prefers-reduced-motion` is the only motion-safety path** now (calm mode removed, PRD §10 #12) — WCAG motion-safety depends on it, so treat it as required, not optional.
