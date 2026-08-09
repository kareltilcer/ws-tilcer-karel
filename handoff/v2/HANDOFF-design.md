# Design Handoff — karel (Personal Site)

> For: **Claude Design** · Owner: Karel · Last updated: 2026-08-05 · Status: v1 brief (design delivered 2026-08-07)
> **Read first:** root `CLAUDE.md` (fleet conventions), then `PRD.md` (source of truth for behaviour + data) and `openapi.yaml` (exact response shapes you'll render). This brief covers **what to design and why**; those govern **what it does** and **what data exists**.
>
> This is unlike the other briefs in the fleet. `home`, `fin`, and `status` are function-first internal tools ("make it pretty *after* it works"). **`karel` is the opposite: it is a public personal-brand showcase where the visual experience *is* the product.** Design accordingly.

> **⚠ ADDENDUM 2026-08-07 — CALM MODE REMOVED (supersedes the hard constraint below).** During design, Karel decided to **drop the "calm mode" dual-register** — the site ships a **single full-experience register**. The removed items: the calm/full toggle, the `[data-calm]` machinery, the Arcade calm opt-in gate, and the `prefers-reduced-motion`→calm auto-switch. **Motion-safety is instead met by honoring `prefers-reduced-motion` directly** (ambient CSS loops + the Home hero physics/sprinkles are stilled under the query; functional motion — loaders, active gameplay — is preserved). Wherever this brief says "calm mode" (esp. the Prime-directive hard constraint, §"the one hard constraint", the screen calm-registers, and the DoD), read it as **superseded**; the surviving principle is only that *content is never gated behind a game or animation, and `prefers-reduced-motion` is respected.* See `PRD.md` §10 item 12.

## Prime directive — WOW FIRST

**The visual wow-effect is the single most important success factor of this project.** Quoted from Karel's brief, unabbreviated:

> *"The site needs to have visual wow-effect. It needs to be stunning, preferably a little bit gamified, with summer vibes (ice-cream and stuff). This is absolutely key."*

This is Karel's front door to the world — the thing a recruiter, client, collaborator, or curious stranger sees first. A competent, tasteful, *safe* portfolio is a **failure** of this brief. The bar is: a visitor smiles or says "whoa" within the first second, and remembers the site a week later. Be ambitious, playful, and specific. Bring a real point of view, not a template.

**Go far.** Physics, motion, sound (opt-in), an ice-cream world, easter eggs, and **several minigames** (see §5) are all in scope and encouraged. This is the one project in the fleet where "too much" is the right starting instinct — we can dial back, but start bold.

### The one hard constraint — calm mode (non-negotiable) — **[SUPERSEDED 2026-08-06, see top addendum]**

> _Historical brief (kept for context; calm mode was removed during design):_ The site must be **fully usable and pleasant for visitors who do not want the gamification.** Confirmed with Karel:

> *"The frontend should be usable for people who don't want the gamification, but other than that go really far."*

So design **two coherent registers of the same site**, not one experience with an accessibility apology bolted on:

- **Full experience (default):** the whole wow — motion, physics, interactive ice-cream, games, surprises.
- **Calm mode:** a clean, fast, content-first version. All content (projects, About, Skills, linktree, legal info, contact) is fully reachable and legible with no games, no heavy animation, no motion assault. This is not a stripped page — it should look *intentional and lovely*, just quiet.

Calm mode is triggered by **an explicit toggle** (persistent, obvious, reachable immediately — a visitor who feels motion-sick must find it fast) **and** automatically when the browser reports `prefers-reduced-motion`. **Gamification enhances; it never gates content.** No project, no piece of information, and no navigation may live *only* inside a game or an animation. Treat the calm/full duality as the central design problem of the whole project (§7, Hard Problem 1).

**→ As built:** calm mode was dropped; the surviving requirement is that `prefers-reduced-motion` stills ambient motion/physics and that no content is gated behind a game/animation.

## What `karel` is

A bilingual (Czech/English) personal site + a small self-hosted CMS. Public host **`karel.tilcer.cz`** (canonical; `kaja.tilcer.cz` 301-redirects to it). One SPA (React + TypeScript + Vite + TanStack Query, static Nginx), part of the same droplet fleet as `home`/`fin`/`status`, but with its own distinct, bold identity — this one is *not* meant to look like the others.

Two very different faces live in one app, and they should feel deliberately different (§6):

- **The public site** — the wow surface. Showcase of Karel and his work.
- **The admin CMS** (`/admin`, behind login) — a quiet, efficient, function-first editor. Here the fleet's usual "clarity over decoration" rule *does* apply.

Content Karel manages via the CMS: **projects** (software *and* non-software — grouped by admin-managed **categories**), a **linktree**, an **About** section, a **Skills** section, the **legal/business disclosure**, and **media**. Plus the **arcade** leaderboards and a **contact** inbox.

## Audience & usage context

- **Primary:** the public — potential employers/clients/collaborators, peers, and curious visitors. Mixed technical level. Many arrive from a shared link (so first paint + social preview matter) or from the linktree.
- **Secondary:** Karel, as the sole admin, editing content in the CMS.
- **Devices:** true dual-priority — a large share of traffic is **mobile** (links shared in chats/socials), and the wow must land on a phone, not just a 27" display. Design mobile and desktop as equals. The full-experience motion/games must degrade gracefully on low-end phones (see performance bar, §9).
- **Two visitor intents:** (a) *"who is this person?"* — a delightful wander; (b) *"I need X"* — find a project, grab a link, or contact Karel fast. The design must reward the wanderer **and** get the goal-directed visitor to content in seconds (the linktree and a clear nav are the fast paths).

## Visual direction — summer & ice-cream

The mood: **bright, warm, joyful, a little playful and surreal.** Think a perfect summer afternoon — sun, melting ice-cream, a pool, sprinkles. This is the emotional target; the exact execution is yours to propose, but it must be unmistakably *summer* and unmistakably *fun*.

Raw material to draw on (use freely, don't feel limited to it): **ice-cream** (scoops, cones, popsicles, sprinkles, drips/melt, a bite taken out), sun and sunbeams, pastel-meets-saturated summer palettes (creams, sorbet pinks, pistachio, mango, sky blue, cherry red), soft gradients, grain/paper texture, chunky rounded type, sticker/badge aesthetics, gentle physics (things that wobble, bounce, melt, or can be flung). Sound is welcome but **opt-in and off by default** (a mute/unmute control; never autoplay audio).

Establish, in the design-system doc:

- **Color tokens** for a light "daylight" theme as the canonical look. A dark/"evening" or "night at the gelateria" variant is optional — propose it if it strengthens the concept, but daylight is the hero. Whatever you pick must pass **AA contrast** for all text (see §9) — saturated summer colors on cream backgrounds are a classic contrast trap.
- **Typography** with personality (a characterful display face for headlines, a highly legible face for body/CMS content) — and it must **not clip Czech diacritics** (see §8). Pick faces that actually ship the full Czech glyph set.
- **A motion language** — the vocabulary of how things wobble/melt/bounce/drip — defined as reusable tokens (durations, easings, spring configs), *and* its calm-mode counterpart (what each motion collapses to when reduced). _[As built: the "calm collapse" became the `prefers-reduced-motion` behaviour — ambient loops stilled — since calm mode was removed.]_
- **An illustration/asset direction** for the ice-cream world (2D sticker style? soft-3D? hand-drawn?) — propose one and commit; consistency beats variety here.

You have real creative latitude on the concept. What's fixed: summer + ice-cream + fun + wow, bilingual, and (originally) the calm-mode duality — now the single-register + reduced-motion path.

## The gamification layer & minigames (§ the fun)

Gamification spans two tiers — design both:

**1. Ambient playfulness (everywhere).** Micro-interactions and surprises woven through the normal site: a playful cursor, elements that react to hover/scroll/drag, a hero moment that delights on first paint, physics on some objects, and **easter eggs** (reward the curious — a Konami-style secret, a clickable hidden scoop, etc.). This is what makes the site feel alive even for someone who never opens a "game." All of it must be stilled cleanly under `prefers-reduced-motion`.

**2. The arcade — actual minigames with leaderboards.** There is an **Arcade** area (its own section/route) hosting **several small games** (summer/ice-cream themed — e.g. catch-the-falling-scoop, a melt-timer/reaction game, stack-the-cones — *propose the set and their mechanics; you have latitude*). Backend contract you're designing against (from `PRD.md` FR-12/13 and `openapi.yaml`):

- Games and their score bounds are **defined in code** (not admin-editable) — so the set you design is a fixed roster per release.
- Each game can have a **persistent leaderboard**: `GET /api/arcade/{gameKey}/scores` → top-N `ScoreEntry` `{ player_name, score, created_at }`.
- Submitting a score: `POST /api/arcade/{gameKey}/scores` with `{ player_name, score, … }` → `ScoreSubmitResult { rank, top[] }`. The submit form is **spam-guarded** (a Turnstile challenge + a honeypot field) — design the score-submit moment to include a Turnstile widget and a name entry, with a graceful "you placed #N" result and the updated board.

Design for each game: an **attract/idle** state (invites play), **playing**, **game-over + score submit**, and the **leaderboard** view. And design the **arcade hub** that houses them. Keep entry frictionless — a visitor should be playing within a tap or two, no account.

_[As built: 4-game roster with 2 games fully playable (Catch-the-scoop, Scoop Match). The calm-mode opt-in gate was removed with calm mode.]_

## Language, locale & theme (fixed)

### Bilingual — Czech & English, Czech default

The whole public site and CMS are **CZ/EN**. Czech is the default/primary; an **instant in-page language switch** (no reload, no refetch — both languages arrive in the payload as `*_cs`/`*_en`, see `openapi.yaml`). Design:

- A **language toggle** that's always reachable (pairs naturally with the sound-mute control in a small persistent control cluster — propose the placement). _[Originally also housed the calm toggle; that's removed.]_
- Persisted choice (localStorage); first visit may follow `Accept-Language` but the toggle always wins.
- A **fallback** treatment for when one language of a field is empty (the CMS flags gaps, but the public site must still render — show the other language rather than a blank).

Write real UX copy **in both languages** (use `design:ux-copy`, once per language — don't ship English strings machine-translated to Czech or vice-versa). Playful microcopy is part of the wow: the empty states, the 404, the game over screens, the easter-egg payoffs should all have charm — in both languages.

### Czech has design consequences (same as the `home` brief — respect them)

- **Strings run longer than English**, and compound labels grow. Nav items, buttons, category chips, and game UI must survive the longer Czech string without clipping or reflowing badly. Verify with real Czech, never ASCII placeholders.
- **Diacritics need vertical room:** ě š č ř ž ý á í é ú ů ď ť ň and capitals (Č Ř Ů) sit taller — a characterful display face with tight leading will clip them. **Test headlines with real accented Czech**, because the wow type is exactly where this breaks.
- **Three plural forms** (1 / 2–4 / 5+): *1 projekt · 2 projekty · 5 projektů*, *1 skóre*… any count label needs all three; design the copy and flag for implementation.
- **Czech collation** (č, ř, š, ž sort *after* c, r, s, z) wherever anything sorts alphabetically.
- **Formats:** dates `d. M. yyyy` (*5. 8. 2026*), 24-hour time, space thousands separator, comma decimal — for project dates, leaderboard timestamps, etc.

### The legal/Impressum text is Czech-primary

The business disclosure is a Czech legal obligation. Czech is authoritative; English is a courtesy translation alongside. (Content details in §6.9.)

## Screen inventory — public site

Design each screen in its **default, empty, loading, and error** states, at **mobile (375 px) and desktop (1440 px)**, in **both languages** (verify the Czech length). _(Originally: also in a calm register — now a single register with `prefers-reduced-motion` handling.)_ Data field names come from `openapi.yaml`.

### 6.1 Home / hero — the wow moment
The make-or-break screen. This is where "whoa" happens. An immersive summer/ice-cream hero with the strongest interaction on the site, plus a clear path onward to everything (projects, about, linktree, arcade, contact). Must still communicate *who Karel is* in one glance for the goal-directed visitor. Under `prefers-reduced-motion`, the hero physics/sprinkles are stilled to a beautiful resting state.

### 6.2 Projects — grid + filter
Grid/gallery of **published** projects (`GET /api/projects` → `ProjectSummary[]`), each card: cover image, bilingual title, short summary, category, a "featured" treatment. **Filter by category** (`GET /api/categories`; categories are admin-managed, so the filter set is dynamic — design for 2 to ~12 categories, some with long Czech names). Non-software work (crafts, music, etc.) must showcase as beautifully as software — imagery-led. States: empty ("no projects yet"), loading (on-theme skeletons), error.

### 6.3 Project detail
`GET /api/projects/{slug}` → `ProjectDetail`: bilingual title + body (**Markdown**, rendered), a **gallery** of media, and **out-links** (`links[]`: repo/live/demo). Design a rich, media-forward layout that works for both a code project (repo/live links, screenshots) and a physical/creative one (photo gallery). Handle a project with no cover, a long body, and a many-image gallery.

### 6.4 Linktree
The fast path — a prominent, quick, scannable list of outbound links (`GET /api/links` → `Link[]`: bilingual label, url, icon, order). This is often *the* thing a visitor came for; make it delightful but instant. Clicks fire a best-effort counter (`POST /api/links/{id}/click`) — no UI needed for the count, it's admin-side. Design this so it also works as a near-standalone "link-in-bio" moment (it may be the first screen someone lands on from social).

### 6.5 About
`GET /api/sections/about`: bilingual heading + Markdown body + a photo. The "showcase of me" — personality-forward, on-theme.

### 6.6 Skills
`GET /api/skills` → `Skill[]`: grouped by `category` (e.g. languages/tools/craft), each with optional icon and optional level (1–5). Design a playful but honest skills display — **no CV/experience timeline** (out of scope). Handle skills with and without a level.

### 6.7 Arcade
The minigame hub + the games + leaderboards + score submit (see §5).

### 6.8 Contact
A form: name, email, subject (optional), message, plus a **Cloudflare Turnstile** widget and a hidden honeypot (`POST /api/contact`). Design: idle, submitting, success (`202` — a charming confirmation), validation errors per field, spam-rejected, and rate-limited (`429`) states — localized. Karel's email/phone are **not** splashed here (see §6.9); the form is the primary channel.

### 6.9 Legal / Impressum
`GET /api/business-info`. The Czech OSVČ disclosure — **jméno a příjmení, IČO, sídlo/adresa, and the živnostenský-rejstřík statement** (Karel is **neplátce DPH**, so **DIČ is hidden** unless present). **Privacy-critical:** the **sídlo is Karel's home address** — it appears **only on this legal page** (reachable via a footer link), **never** on the homepage, contact section, or hero. Design a clean, readable disclosure page (Czech primary, English alongside), footer-linked, low-key by intent. Don't over-decorate it — this one is quiet and correct.

### 6.10 Footer & global chrome
A footer with the Impressum link, the language + sound-mute toggles (if not placed elsewhere), and light navigation. Plus the persistent **control cluster** — design where it lives so it's always findable without fighting the hero.

### 6.11 404 / error
An on-theme, playful not-found page with a way home — in both languages.

## Admin CMS (`/admin`) — the quiet, function-first face

A **deliberately different register**: here the fleet's normal "clarity and density over decoration" rule applies. Karel is the only user; this is a tool, not a showcase. It's a **lazy-loaded area of the same SPA**, behind **Mode B** login. Design it clean and efficient — it may share the type/color tokens but should shed the heavy motion and games.

- **Login (Mode B).** Self-hosted login (email + password + "Přihlásit / Sign in") issuing karel's own session — same pattern as `home`/`fin`/`status`. States: idle, submitting, invalid credentials (generic), server/auth-unreachable, and an **MFA-required** case that links out to finish on `auth.tilcer.cz`. **Do NOT design signup, password-reset, TOTP, or Google screens** — those are auth-hosted.
- **CMS editors** for: **projects** (list incl. drafts; editor with bilingual fields, category picker, media/gallery picker, out-links, draft/published, reorder), **categories** (CRUD + reorder; block delete while in use — surface the `409`), **links** (CRUD + reorder + show/hide), **sections** (About/Hero bilingual editors), **skills** (CRUD + reorder), **business-info** (the disclosure form), **media library** (upload incl. **SVG**, with the sanitize step; alt text CZ/EN; block delete of referenced media — the `409`), **contact inbox** (list + status triage new/read/archived/spam + delete), **arcade moderation** (delete a score, reset a board).
- **Bilingual editing:** every translatable field is a CZ/EN pair; **flag missing translations** so Karel sees gaps (the public site falls back, but the editor should nudge).
- Function-first states: loading, empty, error, and the destructive confirmations (delete project/category/link, reset leaderboard, delete media). Keep it calm and quick.

## Component inventory

Public: the hero interaction; project card + category-filter chips; project-detail gallery/lightbox; linktree row/button; skill chip/meter; **minigame shells** (attract/playing/game-over) + **score-submit card** (name + Turnstile) + **leaderboard**; the persistent **control cluster** (language toggle, sound mute); playful empty states + 404; contact form with Turnstile; footer.

Admin (function-first): text/Markdown field pairs (CZ/EN) with missing-translation flag; category/media pickers; reorderable list; draft/published toggle; media uploader (with SVG-sanitize + alt CZ/EN); confirm dialog; toast; inbox row + status control; login form.

## Hard problems — solve these explicitly (with rationale)

1. ~~**The full ↔ calm duality.**~~ **[SUPERSEDED]** Replaced by: **motion-safety via `prefers-reduced-motion`** — ambient motion/physics stilled, no content gated behind a game/animation, all in a single register.
2. **Wow that lands on a phone.** The hero/games must delight at 375 px on a mid-range phone, not only on a big desktop GPU. Show the mobile wow, and the performance fallback (§9).
3. **Delight for the wanderer *and* speed for the goal-directed visitor.** The person who came for one link or one project must reach it in seconds; the wanderer must be rewarded for exploring. Both, at once.
4. **Ice-cream/summer without becoming childish or illegible.** The theme must read as characterful and cool, not a kids' app, and it must never sacrifice content legibility or AA contrast (summer colors on cream is a contrast trap).
5. **Non-software projects showcased as richly as software.** The projects system spans crafts/music/writing/etc. — the card + detail must flatter a photo-led physical project as much as a repo-led code one.
6. **A dynamic, bilingual category filter.** Categories are admin-defined (2–~12, long Czech names) — the filter UI can't assume a fixed known set or English lengths.
7. **Two registers of the app in one identity.** The wow public site and the quiet admin CMS should feel like the same product family without the CMS inheriting the heavy motion, and without the public site feeling like a CMS.
8. **Frictionless minigames + honest leaderboards.** Play within a tap or two, no account; a score-submit that's fun but survives the Turnstile/anti-cheat reality (§5).
9. **Czech + a characterful display face.** The wow typography is exactly where diacritics clip — prove the headline face carries ě/š/č/ř/ž/ů and capitals cleanly, at display size.

## Accessibility & performance quality bar

- **WCAG 2.1 AA.** AA contrast for **all** text and meaningful UI in the chosen summer palette (verify — this palette fails easily). Never encode meaning in hue alone. Full keyboard operability (nav, filters, forms, the arcade entry, the CMS). Visible focus. Touch targets ≥44 px. **Run `design:accessibility-review` on the output before handing back.**
- **`prefers-reduced-motion` is a first-class path** — with calm mode removed, this is *the* motion-safety mechanism: ambient animation and hero physics are stilled; functional motion (loaders, active gameplay) is preserved. Nothing essential may depend on motion.
- **Motion-sensitivity safety:** no autoplaying audio; provide a sound mute; avoid strobe/rapid-flash; honor reduced-motion for any parallax/physics.
- **Performance:** the wow must not tank load. Budget the hero for a fast first paint; lazy-load games and heavy assets; keep the admin bundle out of the public bundle (it's lazy `/admin`). Target smooth on a mid-range phone; provide a reduced-fidelity path for low-power devices (tie it to reduced-motion).
- **SEO note (context, not a design task):** v1 uses client-side meta + sitemap (no SSR yet), so design correct per-page titles and **Open Graph/social-share cards** — a shared link should unfurl beautifully (this is part of the wow, off-site). Provide OG image direction.

## Do NOT design

- **Signup / password-reset / TOTP / Google / MFA-setup screens** — auth-hosted; only the Mode B *login* is ours.
- **A blog / articles / writing section** — deferred to a future version (do not imply it in nav or IA; leave conceptual room but design nothing).
- **A CV/experience timeline** — Skills only, no chronological career timeline.
- **Comments, public accounts, newsletter signup, e-commerce/payments, booking** — none exist.
- **An analytics dashboard** — the link click-count is admin data only; no charts.
- **Karel's home address anywhere but the legal page** — no address in the hero, contact, or footer body (only the footer *link* to the Impressum).
- **DIČ / VAT fields shown** — Karel is a neplátce; DIČ stays hidden.
- **Multi-user / roles in the CMS** — single admin, no role design.
- **File/attachment management beyond the image media library** — no arbitrary file uploads.

## Deliverables from Claude Design

1. **Design-system doc** — tokens (summer color palette with **AA-verified** contrast; type scale incl. a Czech-diacritic-safe display + body pairing; spacing, radii, elevation; the **motion vocabulary** and its reduced-motion behaviour), the component inventory with all states, plus OG-image direction. Expressed against the build stack (React + TS + Tailwind; note any library beyond it).
2. **Hi-fi interactive prototype** — self-contained files showing the key public screens at **375 px and 1440 px**, in **CZ and EN**, with at least one minigame demonstrated end-to-end (attract → play → game-over → submit → leaderboard). In-memory state, no backend.
3. **A pass with `design:accessibility-review`** (return the report), then **`design:design-handoff`** to produce the engineering redline (tokens, spacing, component props, breakpoints, motion specs) for the frontend build.
4. **UX copy in both languages** (`design:ux-copy`, once per language) for hero, nav, empty states, 404, game-over/leaderboard, contact form (incl. errors), and CMS confirmations — with the three Czech plural forms for count labels.

## Definition of done

- [ ] A hero/first-impression that delivers a genuine **wow**, on mobile and desktop.
- [ ] **`prefers-reduced-motion`** fully designed: ambient motion/physics stilled, every screen reachable and legible, nothing essential behind motion/games. _(Calm-mode toggle removed — §top addendum.)_
- [ ] All public screens (§6.1–6.11) in **default/empty/loading/error**, **375/1440**, **CZ + EN**.
- [ ] The **Arcade**: hub + ≥2 minigames with attract/playing/game-over/submit(+Turnstile)/leaderboard.
- [ ] **Projects** grid with a **dynamic bilingual category filter**, and a **detail** that flatters both software and non-software work; **Linktree**; **About**; **Skills** (no timeline); **Contact** (Turnstile + all states); **Legal/Impressum** (Czech-primary, home sídlo **only** here, DIČ hidden).
- [ ] **Admin CMS** in the quiet function-first register: Mode B login (+ error/MFA-redirect states, no signup/reset/TOTP/Google), and editors for projects/categories/links/sections/skills/business-info/media(+SVG sanitize)/contact-inbox/arcade-moderation, with CZ/EN pairs + missing-translation flags and destructive-action confirms.
- [ ] Summer/ice-cream identity tokenised; **AA contrast verified**; **Czech diacritics uncramped** at display size; realistic Czech content in the prototype (never lorem/English placeholders).
- [ ] The **hard problems** addressed with stated rationale.
- [ ] Accessibility pass done; performance/motion-safety approach stated; OG social-card direction provided.
- [ ] Nothing from **Do NOT design** appears.

## Suggested skills

`design:design-system` (tokens + inventory), `design:ux-copy` (CZ and EN — playful microcopy, form errors, game-over/leaderboard, CMS confirms), `design:accessibility-review` (before handoff), `design:design-critique` (self-review), then `design:design-handoff` (engineering redline).

## Open design questions (propose answers)

1. **Daylight-only, or add an "evening/night" theme variant?** _[Resolved as built: day + night both shipped.]_
2. **Sound design** — how far? A subtle opt-in layer, or richer game audio? (Off by default regardless.)
3. **The minigame roster** — which games, and how many? _[As built: 4-game roster, 2 fully playable.]_
4. **Navigation model** — single long scrolly page, a routed multi-page site, or a hybrid? What best serves both the wanderer and the goal-directed visitor?
5. **Control-cluster placement** — where do the language + sound toggles live so they're instantly findable without fighting the hero?
6. **Illustration medium** for the ice-cream world — 2D sticker, soft-3D, hand-drawn? Pick one and commit.
