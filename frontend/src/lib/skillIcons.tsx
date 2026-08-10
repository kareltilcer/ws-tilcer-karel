// Renders a skill/link icon from the generated registry, tinted to the design
// palette. The stored `icon` string is either a registry key (→ tinted SVG) or
// any other text/emoji (→ rendered literally, so old emoji data keeps working).
import { SKILL_ICONS, ICON_GROUP_ORDER, ICON_GROUP_LABELS } from './skillIcons.data'
import type { IconGroup, SkillIconDef } from './skillIcons.data'

export { SKILL_ICONS, ICON_GROUP_ORDER, ICON_GROUP_LABELS }
export type { IconGroup, SkillIconDef }

// Registry glyphs are stored with this sentinel prefix (e.g. "ic:react"), so a
// literal string a user typed that happens to equal a registry key — "go", "c",
// "star", "react" — is rendered as that text, never silently swapped for a logo.
export const REGISTRY_PREFIX = 'ic:'

// The stored value that selects a registry glyph by its key.
export function iconValue(key: string): string {
  return REGISTRY_PREFIX + key
}

export function getIconDef(icon: string | null | undefined): SkillIconDef | undefined {
  if (!icon || !icon.startsWith(REGISTRY_PREFIX)) return undefined
  const key = icon.slice(REGISTRY_PREFIX.length)
  // Object.hasOwn (not `in`/bracket access) so prototype members like
  // "toString" or "constructor" stored as a key don't resolve to a function.
  return Object.hasOwn(SKILL_ICONS, key) ? SKILL_ICONS[key] : undefined
}

// True when the stored value maps to a registry glyph (vs. a literal emoji).
export function isRegistryIcon(icon: string | null | undefined): boolean {
  return !!getIconDef(icon)
}

/**
 * A skill icon at a given pixel size. Registry glyphs render as an inline SVG
 * recoloured to their palette accent; anything else falls back to the raw string
 * (emoji) so hand-entered icons still show. Decorative by default — the skill
 * name label carries the meaning — so it is aria-hidden unless `title` is given.
 */
export function SkillIcon({
  icon,
  size = 20,
  title,
}: {
  icon: string | null | undefined
  size?: number
  title?: string
}) {
  const def = getIconDef(icon)
  const box: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: size,
    height: size,
    flex: 'none',
  }

  if (!def) {
    // Literal emoji / text, or a soft dot when empty.
    return (
      <span
        style={{ ...box, fontSize: Math.round(size * 0.95), lineHeight: 1 }}
        role={title ? 'img' : undefined}
        aria-label={title}
        aria-hidden={title ? undefined : true}
      >
        {icon || <span style={{ color: 'var(--ink-soft)', opacity: 0.5 }}>•</span>}
      </span>
    )
  }

  const color = `var(--${def.accent})`
  return (
    <span style={box}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 24 24"
        role={title ? 'img' : undefined}
        aria-hidden={title ? undefined : true}
        aria-label={title}
        fill={def.type === 'fill' ? color : 'none'}
        stroke={def.type === 'stroke' ? color : 'none'}
        strokeWidth={def.type === 'stroke' ? 2 : undefined}
        strokeLinecap="round"
        strokeLinejoin="round"
        dangerouslySetInnerHTML={{ __html: def.body }}
      />
    </span>
  )
}
