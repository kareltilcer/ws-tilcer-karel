import { useEffect, useState } from 'react'
import { NavLink, Route, Routes, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { ApiError, setUnauthorizedHandler } from '../api/client'
import { qk } from '../api/keys'
import { Spinner } from '../components/states'
import { login, logout, useSession } from './adminApi'
import ProjectsAdmin from './sections/ProjectsAdmin'
import CategoriesAdmin from './sections/CategoriesAdmin'
import LinksAdmin from './sections/LinksAdmin'
import SkillsAdmin from './sections/SkillsAdmin'
import SectionsAdmin from './sections/SectionsAdmin'
import BusinessAdmin from './sections/BusinessAdmin'
import MediaAdmin from './sections/MediaAdmin'
import InboxAdmin from './sections/InboxAdmin'
import ArcadeAdmin from './sections/ArcadeAdmin'

const NAV = [
  { to: '/admin', label: 'Projekty / Projects', end: true },
  { to: '/admin/categories', label: 'Kategorie / Categories' },
  { to: '/admin/links', label: 'Odkazy / Links' },
  { to: '/admin/sections', label: 'Sekce / Sections' },
  { to: '/admin/skills', label: 'Dovednosti / Skills' },
  { to: '/admin/business', label: 'Firma / Business' },
  { to: '/admin/media', label: 'Média / Media' },
  { to: '/admin/inbox', label: 'Zprávy / Inbox' },
  { to: '/admin/arcade', label: 'Arkáda / Arcade' },
]

export default function AdminApp() {
  const qc = useQueryClient()
  const session = useSession()

  useEffect(() => {
    setUnauthorizedHandler(() => {
      void qc.invalidateQueries({ queryKey: qk.session() })
    })
    return () => setUnauthorizedHandler(null)
  }, [qc])

  if (session.isLoading) {
    return (
      <div style={{ display: 'grid', placeItems: 'center', minHeight: '100vh' }}>
        <Spinner size={36} />
      </div>
    )
  }
  if (session.isError || !session.data) return <Login onDone={() => qc.invalidateQueries({ queryKey: qk.session() })} />

  return <Shell />
}

function Shell() {
  const qc = useQueryClient()
  return (
    <div style={{ display: 'grid', gridTemplateColumns: '240px 1fr', minHeight: '100vh' }} className="admin-shell">
      <aside style={{ background: '#2A1D14', color: '#F3E7D6', padding: '20px 14px', display: 'flex', flexDirection: 'column', gap: 4 }}>
        <div className="display" style={{ fontWeight: 800, fontSize: 20, padding: '4px 10px 16px' }}>
          karel<span style={{ color: '#FF7A1A' }}>.</span> CMS
        </div>
        {NAV.map((n) => (
          <NavLink
            key={n.to}
            to={n.to}
            end={n.end}
            style={({ isActive }) => ({
              padding: '10px 12px',
              borderRadius: 10,
              color: isActive ? '#fff' : '#D8C4AE',
              background: isActive ? 'rgba(255,122,26,.25)' : 'transparent',
              fontWeight: 600,
              fontSize: 14,
            })}
          >
            {n.label}
          </NavLink>
        ))}
        <div style={{ flex: 1 }} />
        <button
          onClick={async () => {
            await logout()
            void qc.invalidateQueries({ queryKey: qk.session() })
          }}
          style={{ marginTop: 12, padding: '10px 12px', borderRadius: 10, border: '1px solid rgba(255,255,255,.2)', background: 'transparent', color: '#D8C4AE', fontWeight: 600, cursor: 'pointer' }}
        >
          Odhlásit / Sign out
        </button>
        <a href="/" style={{ padding: '8px 12px', color: '#9c8467', fontSize: 13 }}>
          ← Zpět na web
        </a>
      </aside>
      <main style={{ padding: 'clamp(20px,3vw,40px)', background: 'var(--bg)', overflow: 'auto' }}>
        <Routes>
          <Route index element={<ProjectsAdmin />} />
          <Route path="categories" element={<CategoriesAdmin />} />
          <Route path="links" element={<LinksAdmin />} />
          <Route path="sections" element={<SectionsAdmin />} />
          <Route path="skills" element={<SkillsAdmin />} />
          <Route path="business" element={<BusinessAdmin />} />
          <Route path="media" element={<MediaAdmin />} />
          <Route path="inbox" element={<InboxAdmin />} />
          <Route path="arcade" element={<ArcadeAdmin />} />
        </Routes>
      </main>
    </div>
  )
}

function Login({ onDone }: { onDone: () => void }) {
  const nav = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [state, setState] = useState<'idle' | 'sending'>('idle')
  const [error, setError] = useState('')

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setState('sending')
    try {
      await login(email, password)
      onDone()
      nav('/admin')
    } catch (err) {
      setState('idle')
      if (err instanceof ApiError) {
        if (err.status === 401) setError('Neplatné přihlašovací údaje / Invalid credentials')
        else if (err.status === 409) setError('MFA required — dokončete na auth.tilcer.cz')
        else if (err.status === 502) setError('Ověřovací služba nedostupná / Auth unreachable')
        else setError('Chyba / Error')
      } else setError('Chyba / Error')
    }
  }

  return (
    <div style={{ display: 'grid', placeItems: 'center', minHeight: '100vh', background: 'var(--bg)', padding: 20 }}>
      <form onSubmit={submit} className="card" style={{ width: 'min(400px,100%)', padding: 30, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div className="display" style={{ fontWeight: 800, fontSize: 24 }}>
          karel<span style={{ color: 'var(--cherry-strong)' }}>.</span> CMS
        </div>
        <p style={{ color: 'var(--ink-soft)', margin: 0, fontSize: 14 }}>Přihlášení / Sign in</p>
        <input className="input" type="email" placeholder="E-mail" required value={email} onChange={(e) => setEmail(e.target.value)} />
        <input className="input" type="password" placeholder="Heslo / Password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        {error && <div style={{ color: 'var(--cherry-strong)', fontWeight: 600, fontSize: 14 }}>{error}</div>}
        <button type="submit" className="btn-primary" disabled={state === 'sending'} style={{ height: 48, fontSize: 16 }}>
          {state === 'sending' ? '…' : 'Přihlásit / Sign in'}
        </button>
        <a href="https://auth.tilcer.cz" style={{ fontSize: 13, textAlign: 'center' }}>
          Zapomenuté heslo / MFA → auth.tilcer.cz
        </a>
      </form>
    </div>
  )
}
