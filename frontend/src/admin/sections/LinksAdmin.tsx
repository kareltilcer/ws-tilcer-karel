import { useState } from 'react'
import type { Link } from '../../api/types'
import { createLink, deleteLink, reorderLinks, updateLink, useAdminLinks } from '../adminApi'
import { AdminCard, BilingualField, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { Spinner } from '../../components/states'

export default function LinksAdmin() {
  const { data, isLoading } = useAdminLinks()
  const [labelCs, setLabelCs] = useState('')
  const [labelEn, setLabelEn] = useState('')
  const [url, setUrl] = useState('')
  const [icon, setIcon] = useState('')

  async function add() {
    const ok = await runAction(() => createLink({ label_cs: labelCs, label_en: labelEn, url, icon: icon || undefined }), 'Přidáno / Added')
    if (ok) {
      setLabelCs('')
      setLabelEn('')
      setUrl('')
      setIcon('')
    }
  }

  const links = data ?? []
  const move = (i: number, dir: -1 | 1) => {
    const next = [...links]
    const j = i + dir
    if (j < 0 || j >= next.length) return
    ;[next[i], next[j]] = [next[j], next[i]]
    void runAction(() => reorderLinks(next.map((l) => l.id)))
  }

  return (
    <div style={{ maxWidth: 820 }}>
      <PageHeader title="Odkazy / Links" />
      <AdminCard style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <BilingualField label="Popisek / Label" cs={labelCs} en={labelEn} onCs={setLabelCs} onEn={setLabelEn} />
          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 8 }}>
            <Field label="URL">
              <input className="input" placeholder="https://…" value={url} onChange={(e) => setUrl(e.target.value)} />
            </Field>
            <Field label="Ikona / Icon">
              <input className="input" placeholder="github" value={icon} onChange={(e) => setIcon(e.target.value)} />
            </Field>
          </div>
          <button className="btn-primary" style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }} onClick={add}>
            + Přidat / Add
          </button>
        </div>
      </AdminCard>

      {isLoading && <Spinner />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {links.map((l, i) => (
          <Row key={l.id} l={l} onUp={() => move(i, -1)} onDown={() => move(i, 1)} />
        ))}
      </div>
    </div>
  )
}

function Row({ l, onUp, onDown }: { l: Link; onUp: () => void; onDown: () => void }) {
  const [labelCs, setLabelCs] = useState(l.label_cs)
  const [labelEn, setLabelEn] = useState(l.label_en)
  const [url, setUrl] = useState(l.url)
  const [icon, setIcon] = useState(l.icon ?? '')
  const [visible, setVisible] = useState(l.visible)
  return (
    <AdminCard>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
          <input className="input" value={labelCs} onChange={(e) => setLabelCs(e.target.value)} />
          <input className="input" value={labelEn} onChange={(e) => setLabelEn(e.target.value)} />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 8 }}>
          <input className="input" value={url} onChange={(e) => setUrl(e.target.value)} />
          <input className="input" value={icon} onChange={(e) => setIcon(e.target.value)} />
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
          <label style={{ display: 'flex', gap: 6, alignItems: 'center', fontSize: 14 }}>
            <input type="checkbox" checked={visible} onChange={(e) => setVisible(e.target.checked)} /> Viditelné / Visible
          </label>
          <span className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)' }}>
            {l.click_count} kliknutí / clicks
          </span>
          <div style={{ flex: 1 }} />
          <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={onUp}>
            ↑
          </button>
          <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={onDown}>
            ↓
          </button>
          <button
            className="btn-primary"
            style={{ height: 32, padding: '0 14px', fontSize: 13 }}
            onClick={() => runAction(() => updateLink(l.id, { label_cs: labelCs, label_en: labelEn, url, icon, visible }), 'Uloženo / Saved')}
          >
            Uložit / Save
          </button>
          <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteLink(l.id), 'Smazáno / Deleted')} />
        </div>
      </div>
    </AdminCard>
  )
}
