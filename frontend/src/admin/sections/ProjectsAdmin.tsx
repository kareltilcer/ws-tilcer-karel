import { useState } from 'react'
import type { Media, ProjectDetail, ProjectLink } from '../../api/types'
import {
  createProject,
  deleteProject,
  reorderProjects,
  updateProject,
  useAdminCategories,
  useAdminProjects,
} from '../adminApi'
import { AdminCard, BilingualField, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { MediaPicker } from '../MediaPicker'
import { Spinner, CenterState } from '../../components/states'

export default function ProjectsAdmin() {
  const { data, isLoading } = useAdminProjects()
  const [editing, setEditing] = useState<ProjectDetail | 'new' | null>(null)

  if (editing) return <Editor project={editing === 'new' ? null : editing} onClose={() => setEditing(null)} />

  const items = data?.items ?? []
  const move = (i: number, dir: -1 | 1) => {
    const next = [...items]
    const j = i + dir
    if (j < 0 || j >= next.length) return
    ;[next[i], next[j]] = [next[j], next[i]]
    void runAction(() => reorderProjects(next.map((p) => p.id)))
  }

  return (
    <div style={{ maxWidth: 900 }}>
      <PageHeader
        title="Projekty / Projects"
        action={
          <button className="btn-primary" style={{ height: 42, padding: '0 18px' }} onClick={() => setEditing('new')}>
            + Nový projekt / New
          </button>
        }
      />
      {isLoading && <Spinner />}
      {data && items.length === 0 && <CenterState emoji="🧶" title="Zatím žádné projekty / No projects yet" />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {items.map((p, i) => (
          <AdminCard key={p.id} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{ width: 52, height: 52, borderRadius: 10, overflow: 'hidden', background: 'var(--surface-2)', display: 'grid', placeItems: 'center', flexShrink: 0 }}>
              {p.cover ? <img src={p.cover.public_url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : '🍦'}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontWeight: 700 }}>
                {p.title_cs || p.title_en} {p.featured && '⭐'}
              </div>
              <div className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)' }}>
                {p.slug} · {p.category.name_cs}
              </div>
            </div>
            <span
              style={{
                padding: '3px 10px',
                borderRadius: 999,
                fontSize: 12,
                fontWeight: 700,
                background: p.status === 'published' ? 'color-mix(in srgb,var(--pistachio) 40%,var(--surface))' : 'var(--surface-2)',
                color: 'var(--ink)',
              }}
            >
              {p.status}
            </span>
            <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={() => move(i, -1)}>
              ↑
            </button>
            <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={() => move(i, 1)}>
              ↓
            </button>
            <button className="btn-primary" style={{ height: 32, padding: '0 14px', fontSize: 13 }} onClick={() => setEditing(p)}>
              Upravit / Edit
            </button>
            <ConfirmButton small label="Smazat" onConfirm={() => runAction(() => deleteProject(p.id), 'Smazáno / Deleted')} />
          </AdminCard>
        ))}
      </div>
    </div>
  )
}

function Editor({ project, onClose }: { project: ProjectDetail | null; onClose: () => void }) {
  const cats = useAdminCategories()
  const [slug, setSlug] = useState(project?.slug ?? '')
  const [categoryId, setCategoryId] = useState<number | ''>(project?.category.id ?? '')
  const [titleCs, setTitleCs] = useState(project?.title_cs ?? '')
  const [titleEn, setTitleEn] = useState(project?.title_en ?? '')
  const [summaryCs, setSummaryCs] = useState(project?.summary_cs ?? '')
  const [summaryEn, setSummaryEn] = useState(project?.summary_en ?? '')
  const [bodyCs, setBodyCs] = useState(project?.body_cs ?? '')
  const [bodyEn, setBodyEn] = useState(project?.body_en ?? '')
  const [cover, setCover] = useState<{ id: number; public_url: string } | null>(project?.cover ?? null)
  const [gallery, setGallery] = useState<{ id: number; public_url: string }[]>(project?.gallery ?? [])
  const [links, setLinks] = useState<ProjectLink[]>(project?.links ?? [])
  const [projectDate, setProjectDate] = useState(project?.project_date ?? '')
  const [featured, setFeatured] = useState(project?.featured ?? false)
  const [status, setStatus] = useState(project?.status ?? 'draft')

  async function save() {
    if (categoryId === '') return
    const payload = {
      slug,
      category_id: categoryId,
      title_cs: titleCs,
      title_en: titleEn,
      summary_cs: summaryCs,
      summary_en: summaryEn,
      body_cs: bodyCs,
      body_en: bodyEn,
      cover_media_id: cover ? cover.id : null,
      gallery: gallery.map((g) => g.id),
      links,
      project_date: projectDate || null,
      featured,
      status,
    }
    const ok = await runAction(() => (project ? updateProject(project.id, payload) : createProject(payload)), 'Uloženo / Saved')
    if (ok) onClose()
  }

  return (
    <div style={{ maxWidth: 820 }}>
      <PageHeader
        title={project ? 'Upravit projekt / Edit' : 'Nový projekt / New'}
        action={
          <button className="btn-secondary" style={{ height: 40, padding: '0 16px' }} onClick={onClose}>
            ← Zpět / Back
          </button>
        }
      />
      <AdminCard>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="Slug">
              <input className="input" value={slug} onChange={(e) => setSlug(e.target.value)} />
            </Field>
            <Field label="Kategorie / Category">
              <select className="input" value={categoryId} onChange={(e) => setCategoryId(e.target.value ? Number(e.target.value) : '')}>
                <option value="">—</option>
                {cats.data?.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name_cs} / {c.name_en}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <BilingualField label="Název / Title" cs={titleCs} en={titleEn} onCs={setTitleCs} onEn={setTitleEn} />
          <BilingualField label="Shrnutí / Summary" cs={summaryCs} en={summaryEn} onCs={setSummaryCs} onEn={setSummaryEn} />
          <BilingualField label="Text (Markdown)" cs={bodyCs} en={bodyEn} onCs={setBodyCs} onEn={setBodyEn} textarea />

          <MediaPicker label="Cover" selected={cover} onPick={(m: Media) => setCover({ id: m.id, public_url: m.public_url })} onClear={() => setCover(null)} />

          <div>
            <MediaPicker label="Galerie / Gallery (přidat / add)" onPick={(m: Media) => setGallery((g) => (g.find((x) => x.id === m.id) ? g : [...g, { id: m.id, public_url: m.public_url }]))} />
            {gallery.length > 0 && (
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 8 }}>
                {gallery.map((g) => (
                  <div key={g.id} style={{ position: 'relative' }}>
                    <img src={g.public_url} alt="" style={{ width: 60, height: 60, objectFit: 'cover', borderRadius: 8, border: '2px solid var(--cream-line)' }} />
                    <button
                      onClick={() => setGallery((prev) => prev.filter((x) => x.id !== g.id))}
                      style={{ position: 'absolute', top: -6, right: -6, width: 20, height: 20, borderRadius: '50%', border: 'none', background: 'var(--cherry-strong)', color: '#fff', cursor: 'pointer', fontSize: 12 }}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <LinksEditor links={links} onChange={setLinks} />

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, alignItems: 'end' }}>
            <Field label="Datum / Date (YYYY-MM-DD)">
              <input className="input" value={projectDate} onChange={(e) => setProjectDate(e.target.value)} placeholder="2026-08-01" />
            </Field>
            <Field label="Stav / Status">
              <select className="input" value={status} onChange={(e) => setStatus(e.target.value as 'draft' | 'published')}>
                <option value="draft">draft</option>
                <option value="published">published</option>
              </select>
            </Field>
          </div>
          <label style={{ display: 'flex', gap: 8, alignItems: 'center', fontSize: 14 }}>
            <input type="checkbox" checked={featured} onChange={(e) => setFeatured(e.target.checked)} /> Vybraný / Featured
          </label>

          <button className="btn-primary" style={{ height: 46, alignSelf: 'flex-start', padding: '0 24px', fontSize: 16 }} onClick={save}>
            Uložit / Save
          </button>
        </div>
      </AdminCard>
    </div>
  )
}

function LinksEditor({ links, onChange }: { links: ProjectLink[]; onChange: (l: ProjectLink[]) => void }) {
  const update = (i: number, patch: Partial<ProjectLink>) => onChange(links.map((l, idx) => (idx === i ? { ...l, ...patch } : l)))
  return (
    <div>
      <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>Odkazy / Out-links</span>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 6 }}>
        {links.map((l, i) => (
          <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 2fr 1fr auto', gap: 6 }}>
            <input className="input" placeholder="label" value={l.label} onChange={(e) => update(i, { label: e.target.value })} />
            <input className="input" placeholder="https://…" value={l.url} onChange={(e) => update(i, { url: e.target.value })} />
            <input className="input" placeholder="repo/live" value={l.type ?? ''} onChange={(e) => update(i, { type: e.target.value })} />
            <button className="btn-secondary" style={{ height: 46, padding: '0 12px' }} onClick={() => onChange(links.filter((_, idx) => idx !== i))}>
              ×
            </button>
          </div>
        ))}
        <button className="btn-secondary" style={{ height: 38, alignSelf: 'flex-start', padding: '0 16px', fontSize: 13 }} onClick={() => onChange([...links, { label: '', url: '' }])}>
          + Odkaz / Add link
        </button>
      </div>
    </div>
  )
}
