import { useEffect, useId, useRef, useState } from 'react'
import { useLang } from '../i18n/lang'

const SITE_KEY = import.meta.env.VITE_TURNSTILE_SITE_KEY as string | undefined

declare global {
  interface Window {
    turnstile?: {
      render: (el: HTMLElement, opts: Record<string, unknown>) => string
      reset: (id?: string) => void
      remove: (id?: string) => void
    }
    onloadTurnstileCallback?: () => void
  }
}

let scriptPromise: Promise<void> | null = null
function loadScript(): Promise<void> {
  if (scriptPromise) return scriptPromise
  scriptPromise = new Promise((resolve, reject) => {
    if (window.turnstile) return resolve()
    const s = document.createElement('script')
    s.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js'
    s.async = true
    s.defer = true
    s.onload = () => resolve()
    s.onerror = () => reject(new Error('turnstile script failed'))
    document.head.appendChild(s)
  })
  return scriptPromise
}

/** Turnstile renders the Cloudflare widget when VITE_TURNSTILE_SITE_KEY is set;
 *  otherwise (local dev) it shows a mock checkbox that yields a dev token — the
 *  backend skips verification when TURNSTILE_SECRET is unset. onToken receives
 *  the token (or '' when reset/expired). Bumping `resetSignal` forces a fresh
 *  challenge and clears the current token — callers do this after a failed submit,
 *  since a Turnstile token is single-use and a spent one can never re-verify. */
export function Turnstile({ onToken, resetSignal }: { onToken: (token: string) => void; resetSignal?: number }) {
  const { t } = useLang()
  const ref = useRef<HTMLDivElement>(null)
  const widgetId = useRef<string | null>(null)
  const [mockState, setMockState] = useState<'idle' | 'checking' | 'ok'>('idle')
  const id = useId()

  // Reset on demand: discard the spent token and re-arm the widget. Skips the
  // initial mount (nothing to reset yet).
  const firstReset = useRef(true)
  useEffect(() => {
    if (firstReset.current) {
      firstReset.current = false
      return
    }
    if (SITE_KEY) {
      if (widgetId.current && window.turnstile) window.turnstile.reset(widgetId.current)
    } else {
      setMockState('idle')
    }
    onToken('')
  }, [resetSignal, onToken])

  useEffect(() => {
    if (!SITE_KEY) return
    let cancelled = false
    loadScript()
      .then(() => {
        if (cancelled || !ref.current || !window.turnstile) return
        widgetId.current = window.turnstile.render(ref.current, {
          sitekey: SITE_KEY,
          callback: (token: string) => onToken(token),
          'expired-callback': () => onToken(''),
          'error-callback': () => onToken(''),
        })
      })
      .catch(() => {
        /* leaves the challenge unverified; submit stays disabled */
      })
    return () => {
      cancelled = true
      if (widgetId.current && window.turnstile) window.turnstile.remove(widgetId.current)
    }
  }, [onToken])

  if (SITE_KEY) {
    return <div ref={ref} id={id} />
  }

  // Dev mock (no site key configured).
  return (
    <button
      type="button"
      onClick={() => {
        if (mockState !== 'idle') return
        setMockState('checking')
        setTimeout(() => {
          setMockState('ok')
          onToken('dev-mock-token')
        }, 900)
      }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        height: 56,
        padding: '0 16px',
        borderRadius: 12,
        border: `2px solid ${mockState === 'ok' ? 'var(--pistachio)' : 'var(--cream-line)'}`,
        background: 'var(--surface-2)',
        cursor: mockState === 'idle' ? 'pointer' : 'default',
        textAlign: 'left',
        width: '100%',
      }}
    >
      <span
        style={{
          width: 26,
          height: 26,
          borderRadius: 7,
          border: `2px solid ${mockState === 'ok' ? 'var(--pistachio)' : 'var(--ink-soft)'}`,
          display: 'grid',
          placeItems: 'center',
          background: mockState === 'ok' ? 'var(--pistachio)' : 'transparent',
          color: '#fff',
          fontSize: 16,
        }}
      >
        {mockState === 'ok' ? '✓' : ''}
      </span>
      <span style={{ fontWeight: 600, fontSize: 14, color: 'var(--ink)' }}>
        {mockState === 'ok' ? t.contact.tsOk : mockState === 'checking' ? t.arcade.tsChecking : t.contact.ts}
      </span>
      <span className="mono" style={{ marginLeft: 'auto', fontSize: 10, color: 'var(--ink-soft)', textAlign: 'right', lineHeight: 1.1 }}>
        Cloudflare
        <br />
        Turnstile (dev)
      </span>
    </button>
  )
}
