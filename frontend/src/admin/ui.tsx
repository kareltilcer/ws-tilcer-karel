import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '../api/client'

// Quiet, function-first admin primitives (distinct from the wow public register).

export function AdminCard({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return <div className="card" style={{ padding: 20, ...style }}>{children}</div>
}

export function PageHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, marginBottom: 20 }}>
      <h1 className="display" style={{ fontWeight: 800, fontSize: 28, margin: 0 }}>
        {title}
      </h1>
      {action}
    </div>
  )
}

export function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
      <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>
        {label} {hint && <span style={{ color: 'var(--cherry-strong)', fontWeight: 600 }}>{hint}</span>}
      </span>
      {children}
    </label>
  )
}

// A CS/EN input pair that flags a missing translation (public site falls back).
export function BilingualField({
  label,
  cs,
  en,
  onCs,
  onEn,
  textarea,
}: {
  label: string
  cs: string
  en: string
  onCs: (v: string) => void
  onEn: (v: string) => void
  textarea?: boolean
}) {
  const missing = (!cs.trim() && en.trim()) || (cs.trim() && !en.trim())
  const Input = textarea ? 'textarea' : 'input'
  const extra: React.CSSProperties = textarea ? { height: 120, paddingTop: 10, resize: 'vertical' } : {}
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>
        {label} {missing && <span style={{ color: 'var(--tanger)' }}>⚠ chybí překlad / missing translation</span>}
      </span>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
        <Input className="input" placeholder="Česky (CS)" style={extra} value={cs} onChange={(e) => onCs(e.target.value)} />
        <Input className="input" placeholder="English (EN)" style={extra} value={en} onChange={(e) => onEn(e.target.value)} />
      </div>
    </div>
  )
}

export function Toolbar({ children }: { children: React.ReactNode }) {
  return <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>{children}</div>
}

// A destructive action with a two-step inline confirm.
export function ConfirmButton({ label, onConfirm, small }: { label: string; onConfirm: () => unknown; small?: boolean }) {
  const [armed, setArmed] = useState(false)
  const size = small ? { height: 32, padding: '0 10px', fontSize: 13 } : { height: 40, padding: '0 16px', fontSize: 14 }
  return (
    <button
      onClick={() => {
        if (!armed) {
          setArmed(true)
          setTimeout(() => setArmed(false), 3000)
          return
        }
        void onConfirm()
        setArmed(false)
      }}
      style={{
        ...size,
        borderRadius: 11,
        border: '2px solid var(--cherry-strong)',
        background: armed ? 'var(--cherry-strong)' : 'transparent',
        color: armed ? '#fff' : 'var(--cherry-strong)',
        fontWeight: 700,
        cursor: 'pointer',
      }}
    >
      {armed ? 'Opravdu? / Sure?' : label}
    </button>
  )
}

// runAction wraps a mutation with success/error toasts (surfaces 409s etc.).
// Single bilingual code→message table for API errors, shared by runAction (toasts)
// and the media batch uploader (per-item text) so both stay in sync.
export function apiErrorText(err: unknown): string {
  if (!(err instanceof ApiError)) return 'Chyba / Error'
  if (err.detail) return err.detail
  switch (err.status) {
    case 409:
      return 'Konflikt / Conflict (still referenced)'
    case 413:
      return 'Příliš velké / Too large'
    case 415:
      return 'Nepodporovaný typ / Unsupported type'
    case 422:
      return 'Neplatné údaje / Invalid input'
    default:
      return err.code
  }
}

export async function runAction(fn: () => Promise<unknown>, successMsg?: string) {
  try {
    await fn()
    if (successMsg) toast.success(successMsg)
    return true
  } catch (err) {
    toast.error(apiErrorText(err))
    return false
  }
}
