// Ink-outlined ice-cream sprites (flat SVG, hand-drawn outline) — the committed
// illustration medium. Colours are read from CSS variables at call time so the
// sprites respond to the day/night theme.

export function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

export function coneSprite(a: string, b: string): string {
  return `<svg viewBox="0 0 100 140" width="100%" height="100%"><g stroke="#3A2416" stroke-width="4" stroke-linejoin="round"><path d="M50 138 L26 74 H74 Z" fill="#E7B26B"/><path d="M34 86 h32 M40 100 h20" stroke="#B9823C" stroke-width="3" fill="none"/><circle cx="50" cy="66" r="24" fill="${b}"/><circle cx="36" cy="52" r="20" fill="${a}"/><circle cx="64" cy="52" r="20" fill="${a}"/><circle cx="50" cy="40" r="19" fill="${b}"/><circle cx="50" cy="28" r="6" fill="#D62A54"/></g></svg>`
}

export function popSprite(c: string): string {
  return `<svg viewBox="0 0 80 150" width="100%" height="100%"><g stroke="#3A2416" stroke-width="4" stroke-linejoin="round"><rect x="34" y="96" width="12" height="46" rx="6" fill="#E7B26B"/><rect x="16" y="12" width="48" height="92" rx="24" fill="${c}"/><path d="M40 12 v92" stroke="rgba(255,255,255,.5)" stroke-width="6" stroke-linecap="round"/><circle cx="30" cy="40" r="4" fill="#fff"/><circle cx="50" cy="66" r="4" fill="#fff"/></g></svg>`
}

export function cupSprite(c: string): string {
  return `<svg viewBox="0 0 110 130" width="100%" height="100%"><g stroke="#3A2416" stroke-width="4" stroke-linejoin="round"><path d="M22 62 L30 126 H80 L88 62 Z" fill="#FFE9C7"/><path d="M55 20 C40 20 30 34 34 50 C24 52 20 64 28 70 H82 C90 64 86 52 76 50 C80 34 70 20 55 20Z" fill="${c}"/><circle cx="55" cy="16" r="6" fill="#D62A54"/></g></svg>`
}

// A single falling scoop (for Catch-the-scoop) and the catcher cone.
export function scoopSprite(c: string): string {
  return `<svg viewBox="0 0 60 70" width="100%" height="100%"><g stroke="#3A2416" stroke-width="4" stroke-linejoin="round"><circle cx="30" cy="34" r="22" fill="${c}"/><circle cx="21" cy="24" r="13" fill="${c}"/><circle cx="39" cy="24" r="13" fill="${c}"/><circle cx="30" cy="16" r="6" fill="#D62A54"/></g></svg>`
}

export function catcherConeSprite(): string {
  return `<svg viewBox="0 0 90 96" width="100%" height="100%"><g stroke="#3A2416" stroke-width="4" stroke-linejoin="round"><path d="M45 92 L20 34 H70 Z" fill="#E7B26B"/><path d="M30 48 h30 M37 64 h16" stroke="#B9823C" stroke-width="3" fill="none"/><ellipse cx="45" cy="34" rx="27" ry="10" fill="#FFF0DD"/></g></svg>`
}

export function prefersReducedMotion(): boolean {
  return !!window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
}
