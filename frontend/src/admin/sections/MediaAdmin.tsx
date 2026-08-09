import { useRef, useState } from 'react'
import { deleteMedia, uploadMedia, useAdminMedia } from '../adminApi'
import { AdminCard, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { Spinner, CenterState } from '../../components/states'

export default function MediaAdmin() {
  const { data, isLoading } = useAdminMedia()
  const fileRef = useRef<HTMLInputElement>(null)
  const [altCs, setAltCs] = useState('')
  const [altEn, setAltEn] = useState('')
  const [uploading, setUploading] = useState(false)

  async function upload() {
    const file = fileRef.current?.files?.[0]
    if (!file) return
    const form = new FormData()
    form.append('file', file)
    if (altCs) form.append('alt_cs', altCs)
    if (altEn) form.append('alt_en', altEn)
    setUploading(true)
    const ok = await runAction(() => uploadMedia(form), 'Nahráno / Uploaded')
    setUploading(false)
    if (ok) {
      if (fileRef.current) fileRef.current.value = ''
      setAltCs('')
      setAltEn('')
    }
  }

  return (
    <div style={{ maxWidth: 900 }}>
      <PageHeader title="Média / Media" />
      <AdminCard style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Field label="Soubor / File (png · jpeg · webp · gif · svg)">
            <input ref={fileRef} type="file" accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml" />
          </Field>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="Alt (CS)">
              <input className="input" value={altCs} onChange={(e) => setAltCs(e.target.value)} />
            </Field>
            <Field label="Alt (EN)">
              <input className="input" value={altEn} onChange={(e) => setAltEn(e.target.value)} />
            </Field>
          </div>
          <p className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)', margin: 0 }}>
            SVG se před uložením sanitizuje. / SVGs are sanitized before storage.
          </p>
          <button className="btn-primary" style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }} onClick={upload} disabled={uploading}>
            {uploading ? 'Nahrávám…' : '↑ Nahrát / Upload'}
          </button>
        </div>
      </AdminCard>

      {isLoading && <Spinner />}
      {data && data.items.length === 0 && <CenterState emoji="🖼️" title="Žádná média / No media yet" />}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(160px,1fr))', gap: 14 }}>
        {data?.items.map((m) => (
          <AdminCard key={m.id} style={{ padding: 10 }}>
            <img src={m.public_url} alt={m.alt_cs ?? m.alt_en ?? ''} style={{ width: '100%', height: 120, objectFit: 'cover', borderRadius: 10, background: 'var(--surface-2)' }} />
            <div className="mono" style={{ fontSize: 11, color: 'var(--ink-soft)', margin: '8px 0', wordBreak: 'break-all' }}>
              {m.mime} · {Math.round(m.size_bytes / 1024)} KB
              {m.width && m.height ? ` · ${m.width}×${m.height}` : ''}
            </div>
            <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteMedia(m.id), 'Smazáno / Deleted')} />
          </AdminCard>
        ))}
      </div>
    </div>
  )
}
