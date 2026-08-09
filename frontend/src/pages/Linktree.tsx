import { useLang } from '../i18n/lang'
import { useLinks, registerClick } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton, CenterState } from '../components/states'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 620, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,32px) 90px' }}>{children}</div>
)

export default function Linktree() {
  const { t } = useLang()
  const { pick } = useLang()
  const { data, isLoading } = useLinks()

  return (
    <Wrap>
      <Seo title={t.linktree.title} description={t.linktree.sub} path="/links" />
      <div style={{ textAlign: 'center', marginBottom: 28 }}>
        <div className="anim-bob" style={{ fontSize: 52 }}>🍦</div>
        <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(30px,6vw,46px)', margin: '6px 0 0' }}>
          {t.linktree.title}
        </h1>
        <p style={{ color: 'var(--ink-soft)', margin: '6px 0 0' }}>{t.linktree.sub}</p>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {isLoading && Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} height={64} style={{ borderRadius: 18 }} />)}
        {data && data.length === 0 && <CenterState emoji="🔗" title={t.linktree.empty} />}
        {data?.map((l) => (
          <a
            key={l.id}
            href={l.url}
            target="_blank"
            rel="noreferrer noopener"
            onClick={() => registerClick(l.id)}
            className="card"
            style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '16px 20px', borderRadius: 18, color: 'var(--ink)', fontWeight: 700, fontSize: 17, transition: 'transform .15s' }}
            onMouseEnter={(e) => (e.currentTarget.style.transform = 'translateY(-3px)')}
            onMouseLeave={(e) => (e.currentTarget.style.transform = 'translateY(0)')}
          >
            <span style={{ fontSize: 24 }}>{iconFor(l.icon)}</span>
            <span style={{ flex: 1 }}>{pick(l.label_cs, l.label_en)}</span>
            <span style={{ color: 'var(--ink-soft)' }}>↗</span>
          </a>
        ))}
      </div>
    </Wrap>
  )
}

function iconFor(icon: string | null): string {
  const map: Record<string, string> = {
    github: '🐙',
    linkedin: '💼',
    twitter: '🐦',
    x: '🐦',
    instagram: '📸',
    mastodon: '🐘',
    email: '✉️',
    web: '🌐',
    youtube: '📺',
  }
  return (icon && map[icon.toLowerCase()]) || '🔗'
}
