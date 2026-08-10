import { useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { deleteMedia, invalidateMedia, uploadMediaFile, useAdminMedia } from '../adminApi'
import { AdminCard, ConfirmButton, Field, PageHeader, apiErrorText, runAction } from '../ui'
import { Spinner, CenterState } from '../../components/states'
import { ApiError } from '../../api/client'

const ACCEPT = 'image/png,image/jpeg,image/webp,image/gif,image/svg+xml'
const ACCEPT_TYPES = new Set(ACCEPT.split(','))
const ACCEPT_EXT = new Set(['png', 'jpg', 'jpeg', 'webp', 'gif', 'svg'])
const CONCURRENCY = 3

// The browser-reported MIME type is unreliable — empty or nonstandard for SVGs on
// some OSes — so accept a file when either its MIME or its extension matches and
// let the backend do the authoritative validation.
function isAccepted(f: File): boolean {
  if (ACCEPT_TYPES.has(f.type)) return true
  const ext = f.name.split('.').pop()?.toLowerCase() ?? ''
  return ACCEPT_EXT.has(ext)
}

type Status = 'pending' | 'uploading' | 'done' | 'error'

interface QueueItem {
  id: string
  file: File
  url: string // object URL for the thumbnail preview
  status: Status
  error?: string
}

let seq = 0
function makeItem(file: File): QueueItem {
  return { id: `${Date.now()}-${seq++}`, file, url: URL.createObjectURL(file), status: 'pending' }
}

// Pending and errored items are the ones still eligible to (re)upload.
const isPendingItem = (it: QueueItem) => it.status === 'pending' || it.status === 'error'

const STATUS_BADGE: Record<Status, { label: string; color: string }> = {
  pending: { label: '•', color: 'var(--ink-soft)' },
  uploading: { label: '…', color: 'var(--ink-soft)' },
  done: { label: '✓', color: '#2f9e44' },
  error: { label: '!', color: 'var(--cherry-strong)' },
}

// Run `worker` over items with at most `limit` in flight at once.
async function runPool<T>(items: T[], limit: number, worker: (item: T) => Promise<void>) {
  let idx = 0
  const runners = Array.from({ length: Math.min(limit, items.length) }, async () => {
    while (idx < items.length) await worker(items[idx++])
  })
  await Promise.all(runners)
}

export default function MediaAdmin() {
  const { data, isLoading } = useAdminMedia()
  const fileRef = useRef<HTMLInputElement>(null)
  const [items, setItems] = useState<QueueItem[]>([])
  const [altCs, setAltCs] = useState('')
  const [altEn, setAltEn] = useState('')
  const [uploading, setUploading] = useState(false)
  const [dragOver, setDragOver] = useState(false)

  // Mirror items into a ref so the unmount cleanup can revoke object URLs
  // without touching state on an unmounted component.
  const itemsRef = useRef<QueueItem[]>([])
  itemsRef.current = items
  useEffect(() => () => itemsRef.current.forEach((it) => URL.revokeObjectURL(it.url)), [])

  // The <input accept> is only a hint and drag-and-drop ignores it entirely, so
  // filter by MIME type here — the one path every file (picked or dropped) takes.
  function addFiles(files: FileList | File[]) {
    const all = Array.from(files)
    const accepted = all.filter(isAccepted)
    const rejected = all.length - accepted.length
    if (rejected) toast.error(`Nepodporovaný typ / Unsupported type: ${rejected}`)
    const next = accepted.map(makeItem)
    if (next.length) setItems((prev) => [...prev, ...next])
  }

  // Revoke object URLs outside the state updater — updaters must stay pure
  // (StrictMode invokes them twice in dev).
  function removeItem(id: string) {
    const gone = items.find((it) => it.id === id)
    if (gone) URL.revokeObjectURL(gone.url)
    setItems((prev) => prev.filter((it) => it.id !== id))
  }

  function clearAll() {
    items.forEach((it) => URL.revokeObjectURL(it.url))
    setItems([])
  }

  const pendingCount = items.filter(isPendingItem).length

  async function uploadAll() {
    const queue = items.filter(isPendingItem)
    if (queue.length === 0) return
    setUploading(true)
    let ok = 0
    let fail = 0
    let authLost = false // a 401 means the session expired and postForm has redirected to login
    const doneUrls: string[] = []
    const setStatus = (id: string, patch: Partial<QueueItem>) =>
      setItems((prev) => prev.map((it) => (it.id === id ? { ...it, ...patch } : it)))

    await runPool(queue, CONCURRENCY, async (it) => {
      if (authLost) return // stop launching more work once the session is gone
      setStatus(it.id, { status: 'uploading', error: undefined })
      const form = new FormData()
      form.append('file', it.file)
      if (altCs) form.append('alt_cs', altCs)
      if (altEn) form.append('alt_en', altEn)
      try {
        await uploadMediaFile(form)
        ok++
        doneUrls.push(it.url)
        setStatus(it.id, { status: 'done' })
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          authLost = true
          setStatus(it.id, { status: 'error', error: 'Odhlášeno / Signed out' })
          return
        }
        fail++
        setStatus(it.id, { status: 'error', error: apiErrorText(err) })
      }
    })

    if (ok > 0) invalidateMedia() // refetch the grid once for the whole batch (nothing changed if all failed)
    setUploading(false)

    // Drop the successful uploads; keep failures in place so they can be retried.
    // Revoke outside the updater to keep it pure (see removeItem).
    doneUrls.forEach((u) => URL.revokeObjectURL(u))
    setItems((prev) => prev.filter((it) => it.status !== 'done'))

    // Session expired mid-batch: postForm already routed to login, so surface the
    // auth loss plainly instead of a misleading success/failure count.
    if (authLost) {
      toast.error('Odhlášeno / Session expired — sign in again')
      return
    }

    if (fail === 0) {
      setAltCs('')
      setAltEn('')
      toast.success(`Nahráno / Uploaded: ${ok}`)
    } else {
      toast.error(`Nahráno / Uploaded: ${ok} · Chyba / Failed: ${fail}`)
    }
  }

  return (
    <div style={{ maxWidth: 900 }}>
      <PageHeader title="Média / Media" />
      <AdminCard style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Field label="Soubory / Files (png · jpeg · webp · gif · svg)">
            <div
              onDragOver={(e) => {
                if (uploading) return
                e.preventDefault()
                setDragOver(true)
              }}
              onDragLeave={(e) => {
                // dragleave also fires when the pointer moves onto a child node;
                // only clear the highlight when it actually leaves the drop zone.
                if (e.relatedTarget instanceof Node && e.currentTarget.contains(e.relatedTarget)) return
                setDragOver(false)
              }}
              onDrop={(e) => {
                e.preventDefault()
                setDragOver(false)
                if (uploading || !e.dataTransfer.files?.length) return
                addFiles(e.dataTransfer.files)
              }}
              style={{
                border: `2px dashed ${dragOver ? 'var(--cherry-strong)' : 'var(--cream-line)'}`,
                borderRadius: 12,
                padding: 16,
                background: dragOver ? 'var(--surface-2)' : 'transparent',
                transition: 'border-color .15s, background .15s',
              }}
            >
              <input
                ref={fileRef}
                type="file"
                multiple
                accept={ACCEPT}
                disabled={uploading}
                onChange={(e) => {
                  if (e.target.files?.length) addFiles(e.target.files)
                  e.target.value = '' // queue is the source of truth; allow re-picking the same file
                }}
              />
              <p className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)', margin: '8px 0 0' }}>
                Vyberte více souborů nebo je sem přetáhněte. / Pick multiple files or drop them here.
              </p>
            </div>
          </Field>

          {items.length > 0 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>
                  {items.length} {items.length === 1 ? 'soubor / file' : 'souborů / files'}
                </span>
                <button className="btn-ghost" style={{ fontSize: 12 }} onClick={clearAll} disabled={uploading}>
                  Vyčistit / Clear
                </button>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {items.map((it) => (
                  <div
                    key={it.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: 8,
                      borderRadius: 10,
                      background: 'var(--surface-2)',
                    }}
                  >
                    <img
                      src={it.url}
                      alt=""
                      style={{ width: 40, height: 40, objectFit: 'cover', borderRadius: 8, flexShrink: 0, background: 'var(--surface)' }}
                    />
                    <div style={{ minWidth: 0, flex: 1 }}>
                      <div className="mono" style={{ fontSize: 12, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {it.file.name}
                      </div>
                      <div className="mono" style={{ fontSize: 11, color: it.status === 'error' ? 'var(--cherry-strong)' : 'var(--ink-soft)' }}>
                        {it.status === 'error' ? it.error : `${Math.round(it.file.size / 1024)} KB`}
                      </div>
                    </div>
                    <StatusBadge status={it.status} />
                    {!uploading && (
                      <button
                        className="btn-ghost"
                        aria-label="Odebrat / Remove"
                        title="Odebrat / Remove"
                        style={{ fontSize: 14, lineHeight: 1, padding: '2px 8px' }}
                        onClick={() => removeItem(it.id)}
                      >
                        ✕
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="Alt (CS)">
              <input className="input" value={altCs} onChange={(e) => setAltCs(e.target.value)} disabled={uploading} />
            </Field>
            <Field label="Alt (EN)">
              <input className="input" value={altEn} onChange={(e) => setAltEn(e.target.value)} disabled={uploading} />
            </Field>
          </div>
          <p className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)', margin: 0 }}>
            Alt se použije na všechny soubory dávky. SVG se před uložením sanitizuje. / Alt applies to every file in the
            batch. SVGs are sanitized before storage.
          </p>
          <button
            className="btn-primary"
            style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }}
            onClick={uploadAll}
            disabled={uploading || pendingCount === 0}
          >
            {uploading ? 'Nahrávám…' : `↑ Nahrát / Upload${pendingCount ? ` (${pendingCount})` : ''}`}
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

function StatusBadge({ status }: { status: Status }) {
  const s = STATUS_BADGE[status]
  return (
    <span className="mono" style={{ fontSize: 14, fontWeight: 800, color: s.color, width: 18, textAlign: 'center', flexShrink: 0 }}>
      {s.label}
    </span>
  )
}
