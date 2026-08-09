import { useState } from 'react'
import { useLang } from '../i18n/lang'
import { Seo } from '../components/Seo'
import { CenterState } from '../components/states'
import { Turnstile } from '../components/Turnstile'
import { apiFetch, ApiError } from '../api/client'
import type { ContactSubmit } from '../api/types'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 620, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

type State = 'idle' | 'sending' | 'success'

export default function Contact() {
  const { t, lang } = useLang()
  const [state, setState] = useState<State>('idle')
  const [error, setError] = useState<string>('')
  const [token, setToken] = useState('')
  const [tsReset, setTsReset] = useState(0)
  const [form, setForm] = useState({ name: '', email: '', subject: '', message: '', website: '' })

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }))

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setState('sending')
    const body: ContactSubmit = {
      name: form.name,
      email: form.email,
      subject: form.subject || undefined,
      message: form.message,
      locale: lang,
      website: form.website,
      turnstile_token: token,
    }
    try {
      await apiFetch<{ ok: boolean }>('/api/contact', { method: 'POST', body })
      setState('success')
    } catch (err) {
      setState('idle')
      // The Turnstile token was consumed by the attempt; re-arm the widget so a
      // retry gets a fresh one instead of re-sending a spent (always-failing) token.
      setTsReset((n) => n + 1)
      if (err instanceof ApiError) {
        if (err.status === 422) setError(t.contact.errValidation)
        else if (err.status === 400) setError(t.contact.errSpam)
        else if (err.status === 429) setError(t.contact.errRate)
        else setError(t.contact.errServer)
      } else setError(t.contact.errServer)
    }
  }

  if (state === 'success')
    return (
      <Wrap>
        <Seo title={t.contact.title} path="/contact" />
        <CenterState emoji="🍦" title={t.contact.successT} body={t.contact.successB} />
      </Wrap>
    )

  return (
    <Wrap>
      <Seo title={t.contact.title} description={t.contact.sub} path="/contact" />
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(32px,5vw,52px)', margin: 0 }}>
        {t.contact.title}
      </h1>
      <p style={{ color: 'var(--ink-soft)', fontSize: 18, margin: '8px 0 24px' }}>{t.contact.sub}</p>

      <form onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Field label={t.contact.name}>
          <input className="input" required value={form.name} onChange={set('name')} />
        </Field>
        <Field label={t.contact.email}>
          <input className="input" type="email" required value={form.email} onChange={set('email')} />
        </Field>
        <Field label={t.contact.subject}>
          <input className="input" value={form.subject} onChange={set('subject')} />
        </Field>
        <Field label={t.contact.message}>
          <textarea className="input" required rows={6} style={{ height: 'auto', paddingTop: 12, resize: 'vertical' }} value={form.message} onChange={set('message')} />
        </Field>
        {/* honeypot — hidden from humans */}
        <input
          value={form.website}
          onChange={set('website')}
          tabIndex={-1}
          autoComplete="off"
          aria-hidden
          style={{ position: 'absolute', left: -9999, width: 1, height: 1, opacity: 0 }}
        />
        <Turnstile onToken={setToken} resetSignal={tsReset} />
        {error && <div style={{ color: 'var(--cherry-strong)', fontWeight: 600, fontSize: 14 }}>{error}</div>}
        <button
          type="submit"
          className="btn-primary"
          disabled={state === 'sending' || !token}
          style={{ height: 52, fontSize: 17, opacity: state === 'sending' || !token ? 0.6 : 1, cursor: state === 'sending' || !token ? 'not-allowed' : 'pointer' }}
        >
          {state === 'sending' ? t.contact.sending : t.contact.send}
        </button>
      </form>
    </Wrap>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>{label}</span>
      {children}
    </label>
  )
}
