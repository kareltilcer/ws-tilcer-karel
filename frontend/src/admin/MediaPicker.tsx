import { useState } from 'react'
import type { Media } from '../api/types'
import { useAdminMedia } from './adminApi'
import { Spinner } from '../components/states'

// A compact media chooser used by the project + section editors.
export function MediaPicker({
  selected,
  onPick,
  onClear,
  label,
}: {
  selected?: { id: number; public_url: string } | null
  onPick: (m: Media) => void
  onClear?: () => void
  label: string
}) {
  const [open, setOpen] = useState(false)
  const { data, isLoading } = useAdminMedia()
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>{label}</span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        {selected ? (
          <img src={selected.public_url} alt="" style={{ width: 64, height: 64, objectFit: 'cover', borderRadius: 10, border: '2px solid var(--cream-line)' }} />
        ) : (
          <div style={{ width: 64, height: 64, borderRadius: 10, border: '2px dashed var(--cream-line)', display: 'grid', placeItems: 'center', color: 'var(--ink-soft)' }}>—</div>
        )}
        <button type="button" className="btn-secondary" style={{ height: 36, padding: '0 14px', fontSize: 13 }} onClick={() => setOpen((o) => !o)}>
          {open ? 'Zavřít / Close' : 'Vybrat / Choose'}
        </button>
        {selected && onClear && (
          <button type="button" className="btn-secondary" style={{ height: 36, padding: '0 14px', fontSize: 13 }} onClick={onClear}>
            Odebrat / Clear
          </button>
        )}
      </div>
      {open && (
        <div style={{ border: '2px solid var(--cream-line)', borderRadius: 14, padding: 12, background: 'var(--surface-2)' }}>
          {isLoading ? (
            <Spinner />
          ) : (data?.items.length ?? 0) === 0 ? (
            <p style={{ color: 'var(--ink-soft)', margin: 0, fontSize: 14 }}>Žádná média — nahraj v sekci Média. / No media — upload in the Media tab.</p>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(72px,1fr))', gap: 8, maxHeight: 220, overflow: 'auto' }}>
              {data?.items.map((m) => (
                <button
                  key={m.id}
                  type="button"
                  onClick={() => {
                    onPick(m)
                    setOpen(false)
                  }}
                  style={{ padding: 0, border: '2px solid var(--cream-line)', borderRadius: 10, overflow: 'hidden', cursor: 'pointer', background: 'var(--surface)' }}
                >
                  <img src={m.public_url} alt={m.alt_cs ?? ''} style={{ width: '100%', height: 72, objectFit: 'cover', display: 'block' }} />
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
