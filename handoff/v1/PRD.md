# PRD — karel (Personal Site & CMS)

> Status: Draft (open questions resolved 2026-08-05) · Owner: Karel · Last updated: 2026-08-05
> Slug: `karel` · Repo: `ws-tilcer-karel` · Companion spec: `openapi.yaml` (OpenAPI 3.1, v0.1.0) — drafted 2026-08-05
> Design handoff: `HANDOFF-design.md` (**wow/gamified/summer** brief) — _pending, see §7_

## 1. Overview

- **One-line summary:** Karel's bilingual (CZ/EN) personal site — a showcase of Karel and his projects (software and non-software), a linktree, and the legally required Czech business disclosure — backed by a small self-hosted CMS he edits behind auth.
- **Type:** fe/be pair (two Coolify apps, one origin — mirrors `home`/`fin`/`status`).
- **Subdomain:** **`karel.tilcer.cz`** (canonical). **`kaja.tilcer.cz` 301-redirects** to the canonical host (see FR-2 / §6).
- **Exposure:** public. The whole public site and its read APIs are public. The CMS admin area and all write endpoints sit behind the shared auth backend (**Mode B**, site id `karel`). Three public write endpoints (contact form, minigame score, link-click counter) are open by design and protected by anti-abuse guards.
- **Consumers:**
  - **Visitors** (anonymous public) — browse the site, follow links, play minigames, send a contact message.
  - **Karel** (single admin) — edits all content via the CMS.
  - No other droplet service depends on this one; this service depends only on `auth`.
- **Depends on:** `auth` (site id `karel`, **Mode B**); **Resend** for contact-form email; **self-hosted S3 (Coolify)** for media/image storage; **Cloudflare R2** for DB backup (Litestream, prefix `karel/`); **Cloudflare Turnstile** for public-form spam protection.

### Modules

Backend is a **compile-time modular monolith** (same pattern as `home`/`status`), with three functional modules plus shared infrastructure:

1. **`content`** — the CMS core: projects, project categories, links (linktree), editable page sections (About, Skills), the legal/business-info block, and media. Public read of *published* content; admin CRUD.
2. **`contact`** — public contact-form ingest → stored + emailed via Resend; admin inbox.
3. **`arcade`** — minigame high-scores: public leaderboard read + public score submit (validated, rate-limited); admin moderation. Kept deliberately small; the games themselves (and their score bounds) are defined in frontend/backend code, not the DB.

Shared: `logging`, health probes, **Mode B** auth, S3 media storage, and the bilingual (CS/EN) content contract used across all modules.

## 2. Goals & Non-Goals

**Goals**

- A **stunning, memorable** personal site that makes a strong first impression — a "showcase of Karel," not just a CV. Visual wow-effect is a first-class goal, not polish (see §7 and the design handoff).
- **Playful and gamified**, with room for **several minigames** — while remaining **fully usable by visitors who don't want the gamification** (a calm/reduced experience is a hard requirement, not an afterthought).
- A **project showcase** covering *both* software and non-software work, editable by Karel without touching code, grouped by admin-managed **categories**.
- A **linktree** — a curated, reorderable list of outbound links (socials, projects, elsewhere), with an optional click counter.
- **Bilingual CZ/EN** throughout, with an instant in-page language switch; Czech is the default/primary.
- The **legally required Czech business disclosure** (povinně zveřejněné informace o podnikateli — OSVČ: jméno+příjmení, IČO, sídlo/místo podnikání, zápis v živnostenském rejstříku), editable and always reachable.
- A **contact form** that reliably reaches Karel by email, resistant to spam.
- **Cheap to run:** single droplet, embedded SQLite, media on self-hosted S3, no heavy runtime.

**Non-Goals (v1)**

- **No blog / writing section in v1.** It is a likely *future* feature — the data model and IA must not preclude adding a `posts` module later. **SEO note:** the v1 SEO approach (client-side meta, §7.5) is deliberately minimal; **when the blog lands, revisit and move the public site to full SSR** for proper article indexing/unfurls.
- No CV-timeline / experience-timeline component (Skills yes, chronological CV no).
- No multi-user / multi-author CMS — single admin (Karel).
- No comments, no public accounts, no newsletter.
- No e-commerce, payments, or booking.
- No server-side analytics product (a lightweight per-link click counter exists; no dashboards).
- No CMS-authored *code* or arbitrary HTML injection — content is structured fields + constrained Markdown, sanitized.

## 3. Users & Roles

- **Visitor (anonymous)** — the primary audience. No account. Can read all *published* content, switch language, follow links, play minigames and submit a score, and send a contact message. Cannot see drafts or the admin area.
- **Admin (Karel)** — the only authenticated user. Full CRUD over projects, categories, links, page sections, skills, business info, and media; reads the contact inbox; moderates/resets arcade scores. Authenticated via the shared `auth` backend as site **`karel`** using **Mode B** (self-hosted login + own session cookie for long-term, JWT 15-min for API calls) — the same pattern as `home`/`fin`/`status`.
- **Machine callers** — none inbound beyond browsers. Outbound only: the backend calls Resend (email) and the S3 store (media). No BE→BE `X-Service-Secret` surface in v1.

## 4. Functional Requirements

### FR-1: Serve public site content (read)
- **Description:** Expose all *published* content for the public SPA.
- **Trigger:** `GET` on public content endpoints (projects, categories, links, sections, skills, business info).
- **Inputs:** optional `?category=<slug>`, `?featured=true` (projects); no auth.
- **Behaviour:** Return only rows with `status = published` / `visible = true`, ordered by `sort_order` then date. Every translatable field is returned in **both** `cs` and `en` so the frontend can switch language instantly without refetching. Unpublished/hidden items are never exposed on public endpoints.
- **Outputs:** `200` JSON.
- **Errors:** `404` for an unknown project slug; otherwise `200` with possibly-empty arrays.

### FR-2: Canonical-domain redirect (kaja → karel)
- **Description:** `kaja.tilcer.cz` must not serve a duplicate site.
- **Trigger:** Any request whose `Host` is `kaja.tilcer.cz`.
- **Behaviour:** Respond **301** to the same path/query on `https://karel.tilcer.cz`. Implemented at the edge (Coolify/Traefik host rule) — **preferred over app-level** — so the SPA and API only ever run on the canonical host. `<link rel="canonical">` and `og:url` always point at `karel.tilcer.cz`.
- **Outputs:** `301` with `Location`.
- **Errors:** n/a.

### FR-3: Manage projects (admin)
- **Description:** Full CRUD + ordering + publish for showcase projects (software *and* non-software).
- **Trigger:** `POST/PATCH/DELETE /api/admin/projects[/{id}]`, `POST /api/admin/projects/reorder`.
- **Inputs:**
  - `slug` (string, `^[a-z0-9][a-z0-9-]{0,80}$`, unique).
  - `category_id` (FK → `category`, required).
  - `title_cs`, `title_en` (required); `summary_cs`, `summary_en` (short); `body_cs`, `body_en` (Markdown, sanitized).
  - `cover_media_id` (FK → media, optional); `gallery` (ordered list of media ids, optional).
  - `links` (array of `{ label, url, type }`, e.g. repo/live/demo).
  - `project_date` (optional), `featured` (bool), `status` (`draft|published`), `sort_order` (int).
- **Behaviour:** Validate slug format + uniqueness, that `category_id` and referenced media exist. Sanitize Markdown → safe HTML (no scripts). Reorder accepts an ordered id list and rewrites `sort_order` transactionally.
- **Outputs:** `201`/`200` project; `204` on delete (also detaches gallery joins).
- **Errors:** `409` dup slug; `422` validation / unknown category; `401` unauthenticated.

### FR-4: Manage project categories (admin)
- **Description:** The admin-managed category list projects are grouped and filtered by.
- **Trigger:** `GET /api/categories` (public), `POST/PATCH/DELETE /api/admin/categories[/{id}]`, `POST /api/admin/categories/reorder`.
- **Inputs:** `slug` (`^[a-z0-9][a-z0-9-]{0,40}$`, unique), `name_cs`, `name_en`, `sort_order`, `visible` (bool).
- **Behaviour:** CRUD + reorder. **Delete is blocked (`409`) while any project references the category** — the admin must reassign or remove those projects first (`ON DELETE RESTRICT`). Initial categories are created by Karel in the CMS (none hardcoded).
- **Outputs:** `201`/`200`/`204`.
- **Errors:** `409` dup slug / category in use; `422`; `401`.

### FR-5: Manage linktree links (admin)
- **Description:** CRUD + reorder + show/hide for the curated link list.
- **Trigger:** `POST/PATCH/DELETE /api/admin/links[/{id}]`, `POST /api/admin/links/reorder`.
- **Inputs:** `label_cs`, `label_en`, `url` (validated http/https), `icon` (string key, optional), `visible` (bool, default true), `sort_order`.
- **Behaviour:** Validate URL scheme. Click counting is FR-6.
- **Outputs:** `201`/`200`/`204`.
- **Errors:** `422` invalid URL; `401`.

### FR-6: Link click counter (public) — **v1**
- **Description:** Count outbound link clicks without a heavy analytics stack.
- **Trigger:** `POST /api/links/{id}/click` (fire-and-forget from the frontend).
- **Behaviour:** Increment `link.click_count` (best-effort, per-IP rate-limited). Never blocks navigation; the frontend navigates regardless of the call's outcome.
- **Outputs:** `204`.
- **Errors:** silently no-ops on abuse/rate-limit; never surfaces to the user. `404` if the link id is unknown (ignored client-side).

### FR-7: Edit page sections — About & Skills (admin)
- **Description:** Editable sections that make up the "showcase of me": an **About/bio** block and a **Skills** collection. (No CV/experience timeline in v1.)
- **Trigger:** `GET /api/sections/{key}`, `GET /api/skills` (public); `PUT /api/admin/sections/{key}`, `POST/PATCH/DELETE /api/admin/skills[/{id}]` + reorder (admin).
- **Inputs:**
  - **About** (`key=about`): `heading_cs/en`, `body_cs/en` (Markdown), `photo_media_id` (optional).
  - **Hero** (`key=hero`, optional): headline/tagline `cs/en`, optional media.
  - **Skill** items: `name`, `category` (grouping, e.g. `languages`, `tools`, `craft`), `icon` (optional), `level` (optional int 1–5), `visible`, `sort_order`.
- **Behaviour:** Sections upserted by `key`; skills ordered within their group. All bilingual fields sanitized.
- **Outputs:** `200` section / skill objects.
- **Errors:** `422`; `401` on admin routes.

### FR-8: Legal / business disclosure (Impressum) — OSVČ
- **Description:** The **povinně zveřejněné informace o podnikateli** required of a Czech sole trader (OSVČ) — always reachable from the footer, editable, served in Czech with an English translation alongside.
- **Trigger:** `GET /api/business-info` (public), `PUT /api/admin/business-info` (admin).
- **Fields (singleton):** `legal_name` (jméno a příjmení), `ico` (IČO), `dic` (DIČ — **hidden/empty in v1; Karel is a neplátce DPH**, field kept for future), `registered_seat` (sídlo / místo podnikání — address lines, city, ZIP, country), `register_note_cs/en` ("Zapsáno v živnostenském rejstříku" + issuing živnostenský úřad), `contact_email`, `contact_phone` (optional), `data_note_cs/en` (optional).
- **Behaviour:** Simple upsert; no publish flag (always live). **Presentation:** the full **sídlo (a home address) is shown only on the legal/Impressum page** (footer-linked) — never on the homepage or contact hero. It is public in ARES regardless; the site keeps it low-key. `dic` renders only when non-empty. Values entered by Karel in the CMS (not committed to the repo).
- **Outputs:** `200` object with all fields (translatable ones in both languages).
- **Errors:** `422`; `401` on admin route.

### FR-9: Media upload & management (admin)
- **Description:** Upload images for project covers/galleries and the About photo; store on the **self-hosted S3 (Coolify)** store, keep only metadata in SQLite.
- **Trigger:** `POST /api/admin/media` (upload), `GET /api/admin/media`, `DELETE /api/admin/media/{id}`.
- **Inputs:** image file (`image/png|jpeg|webp|gif`, and **`image/svg+xml` accepted but sanitized** — see below), size-capped (`MAX_MEDIA_BYTES`), optional `alt_cs/en`.
- **Behaviour:** Validate mime + size. **SVG uploads are run through a sanitizer** (strip `<script>`, event handlers, external refs) before storage, and served with a restrictive `Content-Security-Policy` / `Content-Disposition` so they can't execute inline. Store the object in the S3 bucket under the media prefix; persist `{ s3_key, public_url, mime, width, height, size_bytes, alt_cs/en }`. Public URLs are served from `MEDIA_PUBLIC_BASE_URL` (the S3 store's public base), not proxied through the API. Delete removes the S3 object and the row; deleting media still referenced by a project is blocked (`409`) — reassign first.
- **Outputs:** `201` media object with `public_url`.
- **Errors:** `413` too large; `415` unsupported type; `409` still referenced; `422` failed SVG sanitization; `401`.

### FR-10: Submit contact message (public)
- **Description:** A visitor sends Karel a message; it is stored and emailed.
- **Trigger:** `POST /api/contact`.
- **Inputs:** `name`, `email` (validated), `message` (required, length-capped), optional `subject`, `locale` (`cs|en`), a **honeypot** field (must be empty) and a **Turnstile token**.
- **Behaviour:** Reject if honeypot filled or **Turnstile** verification fails. Enforce per-IP rate limit (`CONTACT_RATE`) and body-size cap. Persist a `contact_message` row (`status=new`, with `ip`/`user_agent` for triage). Send an email to `CONTACT_TO` via **Resend** (reply-to = sender); email failure does **not** lose the message (already stored) and is logged. Localize on-page success/error copy by `locale`.
- **Outputs:** `202` `{ ok: true }`.
- **Errors:** `422` invalid input; `429` rate-limited; `400` failed spam check (honeypot/Turnstile). Email-send failures still return `202` (stored) but are logged/flagged.

### FR-11: Contact inbox (admin)
- **Description:** Read and triage received messages.
- **Trigger:** `GET /api/admin/contact`, `PATCH /api/admin/contact/{id}` `{ status }`, `DELETE /api/admin/contact/{id}`.
- **Behaviour:** List newest-first with cursor pagination and `?status=new|read|archived|spam` filter; PATCH updates status; DELETE removes.
- **Outputs:** `200`/`204`.
- **Errors:** `404`; `401`.

### FR-12: Submit minigame score (public)
- **Description:** Record a score for one of the site's minigames.
- **Trigger:** `POST /api/arcade/{gameKey}/scores`.
- **Inputs:** `player_name` (short, sanitized, profanity-guarded), `score` (int), optional `locale`, **honeypot + Turnstile token**.
- **Behaviour:** Validate `gameKey` against the **code-defined game registry** and that `score` is within that game's **hardcoded plausible bound** (reject absurd values). Honeypot + Turnstile + per-IP rate limit. Insert a `game_score` row. Return the submitter's rank.
- **Outputs:** `201` `{ rank, top: [...] }`.
- **Errors:** `404` unknown game; `422` invalid score/name; `400` failed spam check; `429` rate-limited.

### FR-13: Get leaderboard (public)
- **Description:** Top scores for a game.
- **Trigger:** `GET /api/arcade/{gameKey}/scores?limit=`.
- **Behaviour:** Return top-N by score desc (default 10). Cached briefly.
- **Outputs:** `200` array `{ player_name, score, created_at }`.
- **Errors:** `404` unknown game.

### FR-14: Moderate arcade scores (admin)
- **Description:** Remove abusive entries or reset a board.
- **Trigger:** `DELETE /api/admin/arcade/scores/{id}`, `POST /api/admin/arcade/{gameKey}/reset`.
- **Behaviour:** Delete a single score or clear a game's board.
- **Outputs:** `204`.
- **Errors:** `404`; `401`.

### FR-15: Language handling (cross-cutting)
- **Description:** Bilingual CZ/EN behaviour, consistent across all content.
- **Behaviour:** Every translatable field is stored and served in both `cs` and `en`. Czech is the default; the frontend detects `Accept-Language` for first paint but honors an explicit toggle (persisted client-side). Public endpoints return both languages (instant switch, no refetch). If one language is empty for a field, the frontend falls back to the other and the admin UI flags the gap.
- **Outputs:** paired `_cs`/`_en` fields throughout.
- **Errors:** n/a.

### FR-16: Health probes
- **Description:** Baseline observability.
- **Trigger:** `GET /healthz`, `GET /readyz`.
- **Behaviour:** `/healthz` = process up; `/readyz` additionally checks SQLite connectivity. Both public, unauthenticated. (`/readyz` is the recommended `monitor_url` target for the `status` service.)

## 5. Data Model

SQLite (embedded, `modernc.org/sqlite`), migrations via Goose. Every translatable text is stored as paired `*_cs` / `*_en` columns (or JSON `{cs,en}` where optional/rich).

**`category`** — unique `slug`, index `(visible, sort_order)`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `slug` | TEXT NOT NULL UNIQUE | filter key in URLs |
| `name_cs` / `name_en` | TEXT NOT NULL | |
| `sort_order` | INTEGER NOT NULL DEFAULT 0 | |
| `visible` | INTEGER NOT NULL DEFAULT 1 | |
| `created_at` | TEXT NOT NULL | |

**`project`** — index `(status, sort_order)`, `(category_id)`, unique `slug`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `slug` | TEXT NOT NULL UNIQUE | url slug |
| `category_id` | INTEGER NOT NULL REFERENCES category(id) ON DELETE RESTRICT | admin-managed |
| `title_cs` / `title_en` | TEXT NOT NULL | |
| `summary_cs` / `summary_en` | TEXT NULL | short card text |
| `body_cs` / `body_en` | TEXT NULL | Markdown (sanitized) |
| `cover_media_id` | INTEGER NULL REFERENCES media(id) ON DELETE SET NULL | |
| `links` | TEXT NULL | JSON array `{label,url,type}` |
| `project_date` | TEXT NULL | RFC3339/date |
| `featured` | INTEGER NOT NULL DEFAULT 0 | |
| `status` | TEXT NOT NULL DEFAULT 'draft' | draft/published |
| `sort_order` | INTEGER NOT NULL DEFAULT 0 | |
| `created_at` / `updated_at` | TEXT NOT NULL | |

**`project_media`** — gallery join, PK `(project_id, media_id)`, index `(project_id, sort_order)`
| col | type | notes |
|---|---|---|
| `project_id` | INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE | |
| `media_id` | INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE | |
| `sort_order` | INTEGER NOT NULL DEFAULT 0 | |
| `caption_cs` / `caption_en` | TEXT NULL | |

**`media`**
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `s3_key` | TEXT NOT NULL | object key in the S3 store |
| `public_url` | TEXT NOT NULL | from `MEDIA_PUBLIC_BASE_URL` |
| `mime` | TEXT NOT NULL | incl. sanitized `image/svg+xml` |
| `width` / `height` | INTEGER NULL | |
| `size_bytes` | INTEGER NOT NULL | |
| `alt_cs` / `alt_en` | TEXT NULL | accessibility |
| `created_at` | TEXT NOT NULL | |

**`link`** — index `(visible, sort_order)`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `label_cs` / `label_en` | TEXT NOT NULL | |
| `url` | TEXT NOT NULL | http/https |
| `icon` | TEXT NULL | icon key |
| `visible` | INTEGER NOT NULL DEFAULT 1 | |
| `sort_order` | INTEGER NOT NULL DEFAULT 0 | |
| `click_count` | INTEGER NOT NULL DEFAULT 0 | FR-6 counter |
| `created_at` | TEXT NOT NULL | |

**`page_section`** — singleton-per-key
| col | type | notes |
|---|---|---|
| `key` | TEXT PK | e.g. `about`, `hero` |
| `heading_cs` / `heading_en` | TEXT NULL | |
| `body_cs` / `body_en` | TEXT NULL | Markdown |
| `media_id` | INTEGER NULL REFERENCES media(id) ON DELETE SET NULL | e.g. About photo |
| `updated_at` | TEXT NOT NULL | |

**`skill`** — index `(category, sort_order)`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `name` | TEXT NOT NULL | usually language-neutral |
| `category` | TEXT NOT NULL | grouping |
| `icon` | TEXT NULL | |
| `level` | INTEGER NULL | 1–5, optional |
| `visible` | INTEGER NOT NULL DEFAULT 1 | |
| `sort_order` | INTEGER NOT NULL DEFAULT 0 | |

**`business_info`** — singleton (`id=1`)
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK CHECK (id = 1) | one row |
| `legal_name` | TEXT NOT NULL | jméno a příjmení |
| `ico` | TEXT NULL | IČO |
| `dic` | TEXT NULL | DIČ — empty (neplátce); renders only if set |
| `address_lines` / `city` / `zip` / `country` | TEXT NULL | sídlo (home; legal-page only) |
| `register_note_cs` / `register_note_en` | TEXT NULL | živnostenský-rejstřík statement + úřad |
| `contact_email` | TEXT NULL | |
| `contact_phone` | TEXT NULL | |
| `data_note_cs` / `data_note_en` | TEXT NULL | optional |
| `updated_at` | TEXT NOT NULL | |

**`contact_message`** — index `(status, created_at DESC)`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `name` | TEXT NOT NULL | |
| `email` | TEXT NOT NULL | reply-to |
| `subject` | TEXT NULL | |
| `message` | TEXT NOT NULL | |
| `locale` | TEXT NULL | cs/en |
| `ip` | TEXT NULL | triage |
| `user_agent` | TEXT NULL | |
| `status` | TEXT NOT NULL DEFAULT 'new' | new/read/archived/spam |
| `created_at` | TEXT NOT NULL | |

**`game_score`** — index `(game_key, score DESC)`
| col | type | notes |
|---|---|---|
| `id` | INTEGER PK AUTOINCREMENT | |
| `game_key` | TEXT NOT NULL | must match a code-defined game |
| `player_name` | TEXT NOT NULL | short, sanitized |
| `score` | INTEGER NOT NULL | code-bounded per game |
| `locale` | TEXT NULL | |
| `ip` | TEXT NULL | anti-abuse |
| `created_at` | TEXT NOT NULL | |

**Goose notes:** initial migration creates all tables + indexes, enables `PRAGMA foreign_keys=ON`, and seeds the `business_info` singleton and empty `page_section` rows for `about`/`hero`. No categories are seeded (Karel creates them in the CMS). Expected future changes: a `post`/blog module (title/slug/body_cs/en, published_at, tags) when blogging lands — **paired with the SSR migration** noted in §2/§7.5; a `project_tag` table if tags are added alongside categories; a `game` registry table only if minigame config needs to move from code into the DB.

## 6. API Surface

Full detail in `openapi.yaml` (pending). Backend served at `karel.tilcer.cz/api` (no strip-prefix, per the two-app deploy pattern). `kaja.tilcer.cz` 301s to `karel.tilcer.cz` at the edge (FR-2).

**Auth per group**
- `/healthz`, `/readyz` — public.
- **Public read:** `GET /api/projects`, `GET /api/projects/{slug}`, `GET /api/categories`, `GET /api/links`, `GET /api/sections/{key}`, `GET /api/skills`, `GET /api/business-info`, `GET /api/arcade/{gameKey}/scores`.
- **Public write (anti-abuse guarded, no auth):** `POST /api/contact`, `POST /api/arcade/{gameKey}/scores` (both honeypot + **Turnstile** + rate limit), `POST /api/links/{id}/click` (rate limit only).
- **Admin (Mode B — `bearerAuth` JWT with session-cookie fallback, site `karel`):** everything under `/api/admin/**` — projects, categories, links, sections, skills, business-info, media, contact inbox, arcade moderation.

**Conventions:** cursor pagination (`limit` + opaque `cursor`) on inbox/score lists; query-param filters; RFC3339 timestamps; JSON everywhere; bilingual fields returned as `*_cs`/`*_en`.

## 7. Frontend & Design Direction

React + TypeScript + Vite SPA (static Nginx image), TanStack Query. **Bilingual CZ/EN** with an instant in-page switch (Czech default). Served only on the canonical `karel.tilcer.cz`. The CMS **admin area is the same SPA with a lazy-loaded authed `/admin`** bundle (kept out of the public bundle).

### 7.1 Design direction — READ THIS FIRST (drives the design handoff)

> **The visual wow-effect is the single most important success factor of this project.** This site is Karel's personal brand front door and must be **genuinely stunning**, not a template. The following brief must be carried, loud and unabbreviated, into the dedicated **Claude Design** handoff (`HANDOFF-design.md`):
>
> - **Summer vibes** — a bright, joyful, warm palette and mood; **ice-cream** motifs (scoops, cones, sprinkles, melt/drip, popsicles), sun, beach/pool energy. Playful, tactile, delicious.
> - **Gamified & interactive** — go **really far**. Physics, hover/scroll surprises, easter eggs, a playful cursor, delightful micro-interactions, and **several minigames** (e.g. catch-the-scoop, a melt-timer) woven into the experience. The `arcade` backend persists their scores; **game keys and score bounds are defined in code**.
> - **Wow on first paint** — a hero moment that makes visitors smile within the first second.
>
> **Hard constraint — calm mode:** the site must be **fully usable and pleasant for visitors who do not want the gamification.** Provide an explicit **"calm / reduced" toggle** and **respect `prefers-reduced-motion`**: in that mode, heavy animation, physics, and games step aside for a clean, fast, content-first layout. All *content* (projects, about, skills, links, legal info, contact) must be reachable and legible without any game or animation. Gamification enhances; it never gates content.

### 7.2 Public screens

- **Home / hero** — the wow moment; entry to everything; language + calm-mode toggles.
- **Projects** — grid filterable by **category** (software / non-software); **Project detail** with gallery, Markdown body, and out-links.
- **Linktree** — the curated link list, prominent and quick (click-counted, FR-6).
- **About** — bio + photo ("showcase of me").
- **Skills** — grouped skills (no CV timeline).
- **Arcade** — the minigame(s) and their leaderboards.
- **Contact** — the form (name/email/message), localized copy, Turnstile-guarded, success/error states.
- **Legal / Impressum** — the required Czech OSVČ disclosure, always reachable (footer link); the **sídlo/home address appears only here**.

### 7.3 Admin (CMS) area

Behind Mode B auth at `/admin` (lazy-loaded within the same SPA). Screens: projects (list/editor with bilingual fields + media picker + reorder), categories (CRUD + reorder), links (editor + reorder), sections (About/Hero), skills, business info, media library (upload/manage), contact inbox, arcade moderation. Bilingual editors flag missing translations (FR-15).

### 7.4 Data fetching (TanStack Query)

- Query keys: `['projects', filters]`, `['project', slug]`, `['categories']`, `['links']`, `['section', key]`, `['skills']`, `['business-info']`, `['leaderboard', gameKey]`; admin: `['admin','projects']`, `['admin','contact', filters]`, `['admin','media']`, etc.
- Invalidation: each admin mutation invalidates its public + admin keys. Public reads are cache-friendly (long stale time).
- **Empty:** friendly, on-theme empty states. **Loading:** on-brand skeletons. **Error:** inline retry; `401` on admin routes → redirect to auth login.

### 7.5 SEO & sharing

**v1:** per-route `<title>`/meta + Open Graph/Twitter tags set client-side, plus `sitemap.xml`, `robots.txt`, and a canonical URL to `karel.tilcer.cz`. JS-running crawlers (e.g. Google) index fine; some social unfurlers will see the shell — an accepted v1 trade-off for a small personal site. **Documented future rework:** when the **blog** module is implemented, move the public site to **full SSR** so articles index and unfurl correctly (see §2 Non-Goals).

## 8. Non-Functional Requirements

- **Observability:** baseline — `GET /healthz`, `GET /readyz` (SQLite check), structured JSON logs to stdout, per-request logging (method, path, status, latency). Contact-email sends and any S3/Resend/Turnstile failures are logged.
- **Performance:** read-heavy and tiny. Public reads are cacheable and served from SQLite in one query each; images come from the S3 store's public URL, not the API. The only bursty paths are the public POSTs, guarded by rate limits + size caps + Turnstile.
- **Security:** all admin routes require Mode B auth; strict input validation (slugs, URLs, emails, score bounds); Markdown sanitized (no script injection); **SVG uploads sanitized** and served with restrictive CSP/`Content-Disposition`; public forms protected by honeypot + Turnstile + rate limit; media uploads mime/size-checked; secrets only via Coolify env; no secrets or real business-disclosure PII committed to the repo.
- **Accessibility:** WCAG-minded; `prefers-reduced-motion` honored; **calm mode** is a first-class path; all content reachable without JS-driven games; alt text on media (`alt_cs/en`); keyboard-navigable.
- **Backup:** DB via Litestream → **Cloudflare R2** under prefix `karel/`; fresh build restores the DB from R2. **Media lives in the self-hosted S3 (Coolify) store and is NOT covered by the DB backup** — the DB stores only keys/URLs; the S3 store needs its **own persistence/backup** (flagged as an ops task, not this service's responsibility). A fresh app build restores the DB from R2; media persists in the S3 store independently.

## 9. Configuration

All via Coolify env vars (no secrets in repo):

- `AUTH_BASE_URL`, `AUTH_SITE_ID=karel`, plus shared auth config — **Mode B** admin auth (mirror `home`/`fin`/`status`).
- `SESSION_SECRET` — own session-cookie signing.
- `RESEND_API_KEY`, `CONTACT_TO`, `CONTACT_FROM` — contact-form email.
- `TURNSTILE_SECRET` (+ frontend Turnstile site key) — public-form spam protection (contact + score submit).
- `CONTACT_RATE` (e.g. `5/min`/IP), `SCORE_RATE`, `CLICK_RATE`, `MAX_CONTACT_BYTES`, `MAX_MEDIA_BYTES` — anti-abuse/size caps.
- **Media S3 (self-hosted, Coolify):** `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_FORCE_PATH_STYLE` (true for MinIO-style), `MEDIA_PUBLIC_BASE_URL`.
- **DB backup:** `LITESTREAM_*` / Cloudflare R2 creds — prefix `karel/`.
- `PORT` — backend listen port **`2002`** (frontend static app **`2001`**). (home `7999`, fin `8999`, status `112`/`155`.)
- No BE→BE `X-Service-Secret` in v1.

## 10. Decisions & Remaining Items

**Resolved at interview (2026-08-05):**

1. **Slug/repo/auth id** → `karel` (repo `ws-tilcer-karel`, folder `services/karel/`, auth site `karel`); public host stays `karel.tilcer.cz` (canonical) with `kaja.tilcer.cz` 301.
2. **Ports** → backend **2002**, frontend **2001**.
3. **Media storage** → self-hosted **S3 on Coolify** (S3-compatible), public via `MEDIA_PUBLIC_BASE_URL`; DB backup stays Litestream→Cloudflare R2 (`karel/`).
4. **Spam protection** → **Turnstile + honeypot + rate limit** on both public POSTs.
5. **Minigames** → persistent leaderboards; **games defined in code** (keys + score bounds hardcoded).
6. **Project taxonomy** → **admin-managed category list** (`category` table, CMS CRUD).
7. **Business disclosure** → **OSVČ (živnostník)**; **neplátce DPH** (DIČ hidden); **sídlo is a home address, shown on the legal page only**.
8. **Admin UI** → same SPA, lazy-loaded `/admin`.
9. **SEO** → client-side meta + sitemap/robots/canonical for v1; **rework to full SSR when the blog is implemented**.
10. **SVG uploads** → **allowed, sanitized** on upload + safe serving headers.
11. **Link click counter** → **included in v1** (FR-6).

**Remaining (not blockers for the spec):**

- **Specific minigame designs** (which games, mechanics, theming) — to be fleshed out in the **design handoff**; the backend registry just needs their keys + score bounds once chosen.
- **Real business-disclosure values** (IČO, sídlo, živnostenský-úřad wording) — entered by Karel in the CMS; not needed to build.
- **Ops:** provision the Coolify S3 bucket + its own backup; create the Turnstile site/secret keys; Resend domain/sender for `CONTACT_FROM`.

## 11. Acceptance Criteria

- [ ] Public visitor can browse published projects (filter by **category**), open a project detail with gallery + Markdown, see the linktree, About, Skills, and the legal/Impressum page — all in **both CZ and EN with an instant switch**.
- [ ] Drafts and hidden items never appear on public endpoints.
- [ ] `kaja.tilcer.cz` **301-redirects** to `karel.tilcer.cz` (same path); canonical/OG URLs point at `karel.tilcer.cz`.
- [ ] Admin (Mode B, site `karel`) can CRUD + reorder + publish projects, manage **categories**, links, sections, and skills, and edit business info — all with bilingual fields; missing translations are flagged.
- [ ] A category in use cannot be deleted (`409`) until its projects are reassigned/removed.
- [ ] Admin can upload images to the **S3 store**; media serve from `MEDIA_PUBLIC_BASE_URL`; **SVG uploads are sanitized**; deleting referenced media is blocked; SQLite stores only metadata.
- [ ] The OSVČ disclosure (jméno, IČO, sídlo, živnostenský-rejstřík statement) is editable and reachable from the footer; **DIČ is hidden while empty**; the **home sídlo appears only on the legal page**.
- [ ] Contact form: valid submit returns `202`, stores the message, and emails `CONTACT_TO` via Resend (reply-to = sender); honeypot/**Turnstile**/rate-limit reject spam; email-send failure still preserves the stored message and is logged.
- [ ] Admin can read and triage the contact inbox.
- [ ] A minigame can submit a score with server-side **code-defined** score-bound + Turnstile + rate-limit validation; leaderboard reads return top-N; admin can moderate/reset.
- [ ] Link clicks increment `click_count` via `POST /api/links/{id}/click` without blocking navigation.
- [ ] **Calm mode:** an explicit toggle and `prefers-reduced-motion` both yield a fully usable, content-first site with no games/heavy animation gating any content.
- [ ] Design delivers a genuine **wow-effect** with summer/ice-cream theme and gamified interactions (verified against the design handoff).
- [ ] Per-route meta/OG tags, canonical URL, `sitemap.xml`, and `robots.txt` are present (v1 client-side SEO).
- [ ] Observability baseline works (`/healthz`, `/readyz` w/ SQLite check, structured + request logs).
- [ ] Litestream replicates the DB to R2 prefix `karel/`; a fresh build restores from R2; media persists in the S3 store independently.
- [ ] Backend on port **2002**, frontend on **2001**; admin uses **Mode B** against `auth` site `karel`.
- [ ] OpenAPI 3.1 spec matches the implemented surface.
