import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useLang } from '../i18n/lang'
import { useProject } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton, CenterState } from '../components/states'
import { MarkdownView } from '../components/MarkdownView'
import { formatDate } from '../i18n/format'
import type { MediaRef } from '../api/types'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 920, margin: '0 auto', padding: 'clamp(24px,4vw,48px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

export default function ProjectDetail() {
  const { slug = '' } = useParams()
  const { t, pick, lang } = useLang()
  const { data, isLoading, isError } = useProject(slug)
  const [lightbox, setLightbox] = useState<MediaRef | null>(null)

  if (isLoading)
    return (
      <Wrap>
        <Skeleton height={40} style={{ maxWidth: 400 }} />
        <Skeleton height={360} style={{ marginTop: 20, borderRadius: 24 }} />
      </Wrap>
    )
  if (isError || !data)
    return (
      <Wrap>
        <Seo title={t.projects.notFound} path={`/projects/${slug}`} />
        <CenterState emoji="🫠" title={t.projects.notFound} action={<Link className="btn-primary" style={{ padding: '12px 20px', display: 'inline-block' }} to="/projects">{t.projects.back}</Link>} />
      </Wrap>
    )

  const title = pick(data.title_cs, data.title_en)
  const body = pick(data.body_cs, data.body_en)
  return (
    <Wrap>
      <Seo title={title} description={pick(data.summary_cs, data.summary_en)} path={`/projects/${slug}`} />
      <Link to="/projects" style={{ fontWeight: 700, fontSize: 14 }}>
        ← {t.projects.back}
      </Link>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 14 }}>
        <span style={{ padding: '4px 12px', borderRadius: 999, background: 'var(--surface-2)', border: '2px solid var(--cream-line)', fontSize: 13, fontWeight: 700, color: 'var(--ink-soft)' }}>
          {pick(data.category.name_cs, data.category.name_en)}
        </span>
        {data.project_date && <span className="mono" style={{ fontSize: 13, color: 'var(--ink-soft)' }}>{formatDate(lang, data.project_date)}</span>}
      </div>
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(30px,5vw,52px)', margin: '10px 0 0' }}>
        {title}
      </h1>

      {data.cover && (
        <img
          src={data.cover.public_url}
          alt={data.cover.alt_cs ?? data.cover.alt_en ?? title}
          style={{ width: '100%', borderRadius: 24, marginTop: 24, border: '2px solid var(--cream-line)' }}
        />
      )}

      {data.links.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10, marginTop: 20 }}>
          {data.links.map((l, i) => (
            <a key={i} href={l.url} target="_blank" rel="noreferrer noopener" className="btn-secondary" style={{ padding: '10px 16px', fontSize: 14 }}>
              {l.type ? `${l.label} · ${l.type}` : l.label} ↗
            </a>
          ))}
        </div>
      )}

      {body && (
        <div style={{ marginTop: 28, fontSize: 17 }}>
          <MarkdownView>{body}</MarkdownView>
        </div>
      )}

      {data.gallery.length > 0 && (
        <>
          <h2 className="display" style={{ fontWeight: 700, fontSize: 24, margin: '36px 0 14px' }}>
            {t.projects.gallery}
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(180px,1fr))', gap: 12 }}>
            {data.gallery.map((m) => (
              <button key={m.id} onClick={() => setLightbox(m)} style={{ padding: 0, border: '2px solid var(--cream-line)', borderRadius: 16, overflow: 'hidden', cursor: 'zoom-in', background: 'var(--surface-2)' }}>
                <img src={m.public_url} alt={m.alt_cs ?? m.alt_en ?? ''} style={{ width: '100%', height: 150, objectFit: 'cover', display: 'block' }} loading="lazy" />
              </button>
            ))}
          </div>
        </>
      )}

      {lightbox && (
        <div
          onClick={() => setLightbox(null)}
          style={{ position: 'fixed', inset: 0, background: 'rgba(27,21,48,.72)', zIndex: 80, display: 'grid', placeItems: 'center', padding: 24, cursor: 'zoom-out' }}
        >
          <img src={lightbox.public_url} alt={lightbox.alt_cs ?? lightbox.alt_en ?? ''} style={{ maxWidth: '92vw', maxHeight: '88vh', borderRadius: 16, boxShadow: '0 24px 60px rgba(0,0,0,.5)' }} />
        </div>
      )}
    </Wrap>
  )
}
