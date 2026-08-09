# Source association

repo: kareltilcer/ws-tilcer-karel
branch: main
path: handoff/v1

This is a **greenfield design project**, not a code recreation. The repo contains no
application code — only the design handoff brief. Designs here are authored from:
- `handoff/v1/HANDOFF-design.md` — the wow/summer/ice-cream design brief
- `handoff/v1/PRD.md` — behaviour + data model (source of truth)
- `handoff/v1/openapi.yaml` — exact API response shapes

## Last sync

date: 2026-08-07T00:00:00Z
commit: 791ce6c5bfd1
(note: 791ce6c5bfd1 is the tree hash reported by the GitHub tools, not a verified commit sha)

### Updated in this project (2026-08-07 pre-handoff polish pass)
- **Accessibility gap CLOSED — `prefers-reduced-motion` now honoured on every screen.** Added a `@media (prefers-reduced-motion: reduce)` block to all seven site screens + the design system that stills the decorative ambient loops (bob / float / spin / twinkle / drip / wobble) and sets `scroll-behavior:auto`. Functional motion is deliberately preserved: loading spinners (`.spin`), skeleton shimmers (`.skel`), Admin entrance fades, and active gameplay.
- **Home hero physics gated.** Home ran a continuous requestAnimationFrame simulation (gravity on the draggable scoops + 14 falling sprinkles) that CSS cannot stop. Under reduced-motion the sim and sprinkles no longer start; the five scoops are placed at rest and set non-interactive. Verified: 0 sprinkles / 5 resting scoops when the query matches. This resolves the WCAG motion-safety item flagged (not yet actioned) in the calm-mode-removal note below.
- Audited the other screens for JS ambient motion: only Home had it. Arcade's rAF loop runs solely during `phase==='playing'` (user-initiated), so it needs no gate.
- DesignSystem motion section updated to document the reduced-motion behaviour in place of the old "tokens always run" note.
- Two suspected layout bugs (404 heading/subtitle overlap; Pages sub-nav clipped by header) were investigated and found to be html-to-image screenshot artifacts — the live DOM measures clean (header 0–76px, sub-nav pills 88–126px; 404 box grows correctly on wrap). No change needed.

### Updated in this project (2026-08-06 calm-mode removal)
- **Calm mode removed from every design at the client's explicit request.** The site now ships a single full-experience register. Removed across Home, Projects, Arcade, Pages, NotFound, DesignSystem: the calm/full toggle in every control cluster, the `[data-calm]` attribute + `.full-only`/`.calm-only` CSS, all calm state/props/logic, the Arcade calm opt-in gate, and the `prefers-reduced-motion` → calm auto-activation.
- **Handoff / brief impact:** this overrides HANDOFF-design.md's one hard constraint (§"calm mode — non-negotiable") and the related DoD items and hard-problem #1. Recorded here as the client's decision.
- **Accessibility follow-up (flagged, not yet actioned):** removing calm also removed the `prefers-reduced-motion` path, so motion-sensitive visitors currently have no reduced-motion fallback. Recommend the engineering build either honor `prefers-reduced-motion` at the CSS level (disable ambient animation/physics) or reintroduce a lighter motion-off switch — WCAG motion-safety is otherwise unmet.
- DesignSystem.dc.html updated to match: motion section no longer documents a calm collapse, the component inventory is now a single "Behaviour" column, and hard-problem #1 records the removal.

### Updated in this project (2026-08-06 gap-fill pass)
- Closed 3 gaps found in a handoff audit: a second playable minigame, the 404 screen, and OG social-card direction.
- Arcade — **Scoop Match** now playable end-to-end (attract → flip-memory play → board-cleared → score submit w/ Turnstile+honeypot → leaderboard). Two of four games now fully playable, satisfying the DoD “≥2 minigames” bar; per-game HUD (pairs found), how-to, and game-over copy CZ/EN; seeded scoop-match leaderboard.
- **NotFound.dc.html** — on-theme 404 (dropped-scoop hero, playful CZ/EN microcopy + easter-egg line), full + calm registers, day + night, home/projects/arcade escape routes, footer Impressum link. Respects prefers-reduced-motion.
- **OG-Card.dc.html** — social-share direction deliverable: 1200×630 card mocks (default/project/arcade) + spec cards (canvas/safe-area/type/colour) + per-route og:title/description templates, bilingual og:locale note, canonical-host + Twitter summary_large_image guidance.
- **DesignSystem.dc.html** — Deliverable 1 extracted from the prototypes: colour tokens (AA notes, day+night), Czech-safe type scale, spacing/radii/elevation, motion vocabulary + calm collapse, component inventory (full vs calm, all states), live primitives, and the nine hard-problems map. Live theme + calm toggles.

### Updated in this project
- Home / hero screen — full + calm registers, day + night themes, CZ/EN, grab-and-fling ice-cream physics, opt-in sound.
- Projects grid + detail — dynamic bilingual category filter, featured treatment, gallery/lightbox, rendered Markdown, out-links; default/loading/empty/error states, full + calm, day + night, CZ/EN. Built against openapi.yaml ProjectSummary / ProjectDetail / Category shapes.
- Arcade — hub with 4-game roster; Catch-the-scoop playable end-to-end (attract → play → game-over → score submit with mock Turnstile + honeypot → leaderboard), rich game audio, calm-mode opt-in gate. Built against FR-12/13 and openapi.yaml ScoreEntry / ScoreSubmit / ScoreSubmitResult.
- Admin CMS (`/admin`, function-first register) — espresso sidebar shell + quiet light editors; Mode B login (idle/submitting/invalid/unreachable/MFA-redirect, no signup/reset/TOTP/Google); editors for projects (list w/ drafts+featured+reorder+publish+states, editor with CZ/EN pairs + missing-translation flags, category/media pickers, out-links), categories (CRUD/reorder, 409 delete-in-use), links (CRUD/reorder/show-hide + click_count), sections (About/Hero CZ/EN), skills (grouped CRUD/reorder, optional level meter), business-info (OSVČ form, DİČ hidden note, sídlo home-address privacy note), media library (SVG-sanitize badge, alt CZ/EN, 409 delete-referenced), contact inbox (status triage new/read/archived/spam + delete), arcade moderation (delete score, reset board). Confirm dialogs + toasts throughout. Built against HANDOFF §6.11 + PRD FR-3/4/5/7/8/9/11/14 + openapi shapes.
- Pages (About / Skills / Linktree / Contact / Legal-Impressum) — in-page section switch; About (Markdown bio + facts), Skills (grouped scoop-meters, no timeline), Linktree (link-in-bio rows), Contact (all states: idle/validation/spam/rate/sending/success, mock Turnstile + honeypot), Legal (Czech-authoritative OSVČ disclosure, sídlo shown only here, DIČ hidden). Built against PRD FR-7/8/10, openapi Section/Skill/Link/BusinessInfo/ContactSubmit.

Runtime note: the DC runtime does NOT invoke componentDidUpdate or setState callbacks — only componentDidMount. Drive imperative DOM/game side-effects from event handlers via setTimeout after setState, never componentDidUpdate.

## Screen map

| Project screen | Built from |
|---|---|
| Home.dc.html | HANDOFF-design.md §6.1 (hero), §7 (calm duality, control cluster), PRD FR-15 (bilingual), audience notes from Karel |
| Projects.dc.html | HANDOFF-design.md §6.2 (grid+filter), §6.3 (detail), PRD FR-1/FR-3/FR-4, openapi.yaml (ProjectSummary, ProjectDetail, Category, ProjectLink, MediaRef) |
| Arcade.dc.html | HANDOFF-design.md §5, §6.7 (arcade + minigames), PRD FR-12/13/14, openapi.yaml (ScoreEntry, ScoreSubmit, ScoreSubmitResult) |
| Pages.dc.html | HANDOFF-design.md §6.4–6.6, §6.8–6.9, PRD FR-7/8/10, openapi.yaml (Section, Skill, Link, BusinessInfo, ContactSubmit) |
| NotFound.dc.html | HANDOFF-design.md §6.11 (404, playful, both languages, way home), PRD FR-15 |
| OG-Card.dc.html | HANDOFF-design.md §9 (OG social-card direction), PRD §7.5 (client-side meta + canonical), FR-2 (canonical host) |
| DesignSystem.dc.html | HANDOFF-design.md §Visual direction + §Component inventory + §Deliverables 1 + §Hard problems (tokens, type, motion, full/calm inventory, redline) |
| Admin.dc.html | HANDOFF-design.md §6.11 (Admin CMS, quiet function-first register), PRD FR-3/FR-4/FR-5/FR-7/FR-8/FR-9/FR-11/FR-14 + FR-15 (bilingual + missing-translation flag), openapi.yaml (all admin write shapes) |
