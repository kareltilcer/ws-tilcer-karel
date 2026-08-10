import { useLang } from '../i18n/lang'
import { useSkills } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton, CenterState } from '../components/states'
import { SkillIcon, getIconDef, ICON_GROUP_ORDER } from '../lib/skillIcons'
import type { IconGroup } from '../lib/skillIcons'
import type { Skill } from '../api/types'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 1080, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

// Fallback marker per registry icon-group, so a category whose label matches no
// keyword still gets a fitting emoji from the skills it actually contains.
const GROUP_EMOJI: Record<IconGroup, string> = {
  lang: '💻',
  framework: '🧩',
  tool: '🛠️',
  other: '🍦',
}

// Specific markers for well-known category names (finer than the registry's
// coarse groups — e.g. data/design/craft). Returns null so the caller can fall
// back to the icon-derived group below.
function keywordEmoji(cat: string): string | null {
  const c = cat.toLowerCase()
  const map: [RegExp, string][] = [
    [/lang|jazyk/, '💻'],
    [/frame|knihov|library/, '🧩'],
    [/tool|nástroj|nastroj|infra|devops/, '🛠️'],
    [/data|db|databáz|databaz/, '🗄️'],
    [/design|ux|ui/, '🎨'],
    [/craft|hand|ruč|ruc|hobb/, '🧶'],
  ]
  for (const [re, e] of map) if (re.test(c)) return e
  return null
}

// A group header emoji. A matched keyword wins (most specific); otherwise the
// dominant registry group among the category's icons drives it — tracking the
// real skills instead of guessing from the free-text label, so a category named
// in any language still gets a fitting marker before the '🍦' default.
function groupEmoji(cat: string, skills: Skill[]): string {
  const kw = keywordEmoji(cat)
  if (kw) return kw
  const counts = new Map<IconGroup, number>()
  for (const s of skills) {
    const def = getIconDef(s.icon)
    if (def) counts.set(def.group, (counts.get(def.group) ?? 0) + 1)
  }
  let best: IconGroup | null = null
  for (const g of ICON_GROUP_ORDER) {
    if ((counts.get(g) ?? 0) > (best ? counts.get(best)! : 0)) best = g
  }
  return best ? GROUP_EMOJI[best] : '🍦'
}

export default function Skills() {
  const { t, pick } = useLang()
  const { data, isLoading } = useSkills()

  // A Map (not a bare object) so a category literally named "constructor",
  // "toString", etc. can't collide with Object.prototype and blank the page.
  const groups = new Map<string, Skill[]>()
  for (const s of data ?? []) {
    const arr = groups.get(s.category)
    if (arr) arr.push(s)
    else groups.set(s.category, [s])
  }
  const entries = [...groups]

  return (
    <Wrap>
      <Seo title={t.skills.title} description={t.skills.sub} path="/skills" />

      <div style={{ textAlign: 'center', maxWidth: 620, margin: '0 auto' }}>
        <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(34px,6vw,58px)', lineHeight: 1.03, letterSpacing: '-.02em', margin: 0 }}>
          {t.skills.title}
        </h1>
        <p style={{ color: 'var(--ink-soft)', fontSize: 'clamp(16px,1.6vw,19px)', margin: '12px 0 0', lineHeight: 1.5 }}>{t.skills.sub}</p>
      </div>

      {isLoading && (
        <div style={{ marginTop: 40, display: 'flex', flexDirection: 'column', gap: 30 }}>
          {[0, 1].map((i) => (
            <div key={i}>
              <Skeleton height={26} style={{ maxWidth: 220, marginBottom: 16 }} />
              <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
                {Array.from({ length: 4 }).map((_, j) => (
                  <Skeleton key={j} height={78} style={{ width: 160, borderRadius: 18 }} />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
      {data && data.length === 0 && <CenterState emoji="🧠" title={t.skills.empty} />}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 30, marginTop: 40 }}>
        {entries.map(([cat, skills]) => (
          <div key={cat}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
              <span style={{ fontSize: 24 }} aria-hidden>
                {groupEmoji(cat, skills)}
              </span>
              <h2 className="display" style={{ fontWeight: 700, fontSize: 22, margin: 0, textTransform: 'capitalize' }}>
                {cat}
              </h2>
              <div style={{ flex: 1, height: 2, background: 'var(--cream-line)', borderRadius: 2 }} />
            </div>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
              {skills.map((s) => (
                <div
                  key={s.id}
                  className="schip card"
                  style={{ display: 'flex', flexDirection: 'column', gap: 9, padding: '15px 17px', borderRadius: 18, minWidth: 150 }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                    <SkillIcon icon={s.icon} size={20} />
                    <span style={{ fontWeight: 700, fontSize: 16 }}>{pick(s.name_cs, s.name_en)}</span>
                  </div>
                  <Meter level={s.level} />
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </Wrap>
  )
}

// Skill level as scoops out of five (design/v2), with a text label for a11y.
function Meter({ level }: { level: number | null }) {
  const { pick } = useLang()
  if (level == null)
    return (
      <span
        className="mono"
        role="img"
        aria-label={pick('Úroveň neuvedena', 'Level not specified')}
        style={{ fontSize: 12, color: 'var(--ink-soft)' }}
      >
        —
      </span>
    )
  return (
    <div
      style={{ display: 'flex', gap: 4, alignItems: 'center' }}
      role="img"
      aria-label={pick(`Úroveň ${level} z 5`, `Level ${level} of 5`)}
    >
      {[0, 1, 2, 3, 4].map((i) => (
        <span
          key={i}
          aria-hidden
          style={{
            fontSize: 17,
            lineHeight: 1,
            display: 'inline-block',
            filter: i < level ? undefined : 'grayscale(1)',
            opacity: i < level ? 1 : 0.3,
          }}
        >
          🍦
        </span>
      ))}
    </div>
  )
}
