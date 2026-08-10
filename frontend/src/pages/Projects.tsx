import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useLang } from '../i18n/lang'
import { useCategories, useProjects } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton, ErrorState, CenterState } from '../components/states'
import type { ProjectSummary } from '../api/types'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

export default function Projects() {
  const { t, pick, lang } = useLang()
  const [active, setActive] = useState<string>('') // '' = all
  const categories = useCategories()
  const projects = useProjects(active ? { category: active } : undefined)

  return (
    <Wrap>
      <Seo title={t.projects.title} description={t.projects.sub} path="/projects" />
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(34px,6vw,58px)', margin: 0 }}>
        {t.projects.title}
      </h1>
      <p style={{ color: 'var(--ink-soft)', fontSize: 18, margin: '8px 0 0' }}>{t.projects.sub}</p>

      {/* category filter (dynamic, bilingual) */}
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', margin: '24px 0 28px' }}>
        <Chip label={t.projects.all} active={active === ''} onClick={() => setActive('')} />
        {(categories.data ?? []).map((c) => (
          <Chip key={c.id} label={pick(c.name_cs, c.name_en)} active={active === c.slug} onClick={() => setActive(c.slug)} />
        ))}
      </div>

      {projects.isLoading && (
        <Grid>
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} height={280} style={{ borderRadius: 24 }} />
          ))}
        </Grid>
      )}
      {projects.isError && <ErrorState message={t.projects.error} onRetry={() => projects.refetch()} />}
      {projects.data && projects.data.length === 0 && <CenterState emoji="🍨" title={t.projects.empty} />}
      {projects.data && projects.data.length > 0 && (
        <Grid>
          {projects.data.map((p) => (
            <Card key={p.id} p={p} title={pick(p.title_cs, p.title_en)} summary={pick(p.summary_cs, p.summary_en)} catName={pick(p.category.name_cs, p.category.name_en)} lang={lang} />
          ))}
        </Grid>
      )}
    </Wrap>
  )
}

const Grid = ({ children }: { children: React.ReactNode }) => (
  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(280px,1fr))', gap: 20 }}>{children}</div>
)

function Chip({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        padding: '8px 16px',
        borderRadius: 999,
        fontWeight: 700,
        fontSize: 14,
        cursor: 'pointer',
        border: active ? 'none' : '2px solid var(--cream-line)',
        background: active ? 'var(--cherry-strong)' : 'var(--surface-2)',
        color: active ? '#fff' : 'var(--ink)',
      }}
    >
      {label}
    </button>
  )
}

function Card({ p, title, summary, catName }: { p: ProjectSummary; title: string; summary: string; catName: string; lang: string }) {
  return (
    <Link
      to={`/projects/${p.slug}`}
      className="card"
      style={{
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        borderRadius: 24,
        color: 'var(--ink)',
        outline: p.featured ? '3px solid var(--yuzu)' : 'none',
        outlineOffset: p.featured ? 2 : 0,
      }}
    >
      <div style={{ height: 168, background: 'var(--surface-2)', overflow: 'hidden', display: 'grid', placeItems: 'center' }}>
        {p.cover ? (
          <img src={p.cover.public_url} alt={p.cover.alt_cs ?? p.cover.alt_en ?? title} style={{ width: '100%', height: '100%', objectFit: 'cover' }} loading="lazy" />
        ) : (
          <span style={{ fontSize: 44 }}>🍦</span>
        )}
      </div>
      <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 8, flex: 1 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ padding: '3px 10px', borderRadius: 999, background: 'var(--surface-2)', border: '2px solid var(--cream-line)', fontSize: 12, fontWeight: 700, color: 'var(--ink-soft)' }}>
            {catName}
          </span>
          {p.featured && <span title="featured" style={{ fontSize: 14 }}>⭐</span>}
        </div>
        <h3 className="display" style={{ fontWeight: 700, fontSize: 21, margin: 0 }}>
          {title}
        </h3>
        {summary && <p style={{ color: 'var(--ink-soft)', fontSize: 14, lineHeight: 1.5, margin: 0 }}>{summary}</p>}
      </div>
    </Link>
  )
}
