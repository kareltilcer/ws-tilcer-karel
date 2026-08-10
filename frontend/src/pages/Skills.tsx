import { useLang } from '../i18n/lang'
import { useSkills } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton, CenterState } from '../components/states'
import type { Skill } from '../api/types'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 980, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

export default function Skills() {
  const { t } = useLang()
  const { data, isLoading } = useSkills()

  const groups: Record<string, Skill[]> = {}
  for (const s of data ?? []) (groups[s.category] ??= []).push(s)

  return (
    <Wrap>
      <Seo title={t.skills.title} description={t.skills.sub} path="/skills" />
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(34px,6vw,58px)', margin: 0 }}>
        {t.skills.title}
      </h1>
      <p style={{ color: 'var(--ink-soft)', fontSize: 18, margin: '8px 0 28px' }}>{t.skills.sub}</p>

      {isLoading && <Skeleton height={200} />}
      {data && data.length === 0 && <CenterState emoji="🧠" title={t.skills.empty} />}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(280px,1fr))', gap: 18 }}>
        {Object.entries(groups).map(([cat, skills]) => (
          <div key={cat} className="card" style={{ padding: 22 }}>
            <h3 className="display" style={{ fontWeight: 700, fontSize: 20, margin: '0 0 14px', textTransform: 'capitalize' }}>
              {cat}
            </h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {skills.map((s) => (
                <div key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span style={{ fontSize: 18, width: 24, textAlign: 'center' }}>{s.icon || '•'}</span>
                  <span style={{ fontWeight: 600, flex: 1 }}>{s.name}</span>
                  {s.level != null && <Meter level={s.level} />}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </Wrap>
  )
}

function Meter({ level }: { level: number }) {
  return (
    <div style={{ display: 'flex', gap: 4 }} aria-label={`level ${level} of 5`}>
      {[1, 2, 3, 4, 5].map((i) => (
        <span
          key={i}
          style={{
            width: 22,
            height: 12,
            borderRadius: 5,
            background: i <= level ? 'var(--cherry-strong)' : 'var(--cream-line)',
          }}
        />
      ))}
    </div>
  )
}
