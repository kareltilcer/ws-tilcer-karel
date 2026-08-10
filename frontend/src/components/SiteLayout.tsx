import { NavLink, Link, Outlet, useLocation } from 'react-router-dom'
import { useLang } from '../i18n/lang'
import { ControlCluster } from './ControlCluster'

const navItems = [
  { to: '/projects', key: 'projects' as const },
  { to: '/about', key: 'about' as const },
  { to: '/skills', key: 'skills' as const },
  { to: '/arcade', key: 'arcade' as const },
  { to: '/links', key: 'links' as const },
  { to: '/contact', key: 'contact' as const },
]

function Logo() {
  return (
    <Link to="/" style={{ display: 'flex', alignItems: 'center', gap: 12, color: 'var(--ink)' }}>
      <span
        className="display"
        style={{
          display: 'grid',
          placeItems: 'center',
          width: 42,
          height: 42,
          borderRadius: '15px 15px 15px 5px',
          background: 'var(--cherry-strong)',
          color: '#fff',
          fontWeight: 800,
          fontSize: 22,
          boxShadow: 'var(--shadow)',
          transform: 'rotate(-6deg)',
        }}
      >
        k
      </span>
      <span className="display" style={{ fontWeight: 700, fontSize: 21 }}>
        karel<span style={{ color: 'var(--cherry-strong)' }}>.</span>
      </span>
    </Link>
  )
}

export function SiteLayout() {
  const { t } = useLang()
  const loc = useLocation()
  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 40,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 16,
          padding: '14px clamp(18px,5vw,64px)',
          background: 'color-mix(in srgb, var(--bg) 82%, transparent)',
          backdropFilter: 'blur(10px)',
          borderBottom: '2px solid var(--cream-line)',
        }}
      >
        <Logo />
        <nav style={{ display: 'flex', alignItems: 'center', gap: 24 }} className="site-nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className="navlink"
              style={{ fontWeight: 600, fontSize: 16, color: 'var(--ink)' }}
              data-active={loc.pathname.startsWith(item.to)}
            >
              {t.nav[item.key]}
            </NavLink>
          ))}
        </nav>
        <ControlCluster />
      </header>

      <main style={{ flex: 1 }}>
        <Outlet />
      </main>

      <Footer />
    </div>
  )
}

function Footer() {
  const { t } = useLang()
  return (
    <footer
      style={{
        borderTop: '2px solid var(--cream-line)',
        padding: '28px clamp(18px,5vw,64px)',
        display: 'flex',
        flexWrap: 'wrap',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 16,
        color: 'var(--ink-soft)',
        fontSize: 14,
      }}
    >
      <span className="mono">© {new Date().getFullYear()} karel · {t.footer.madeWith}</span>
      <Link to="/legal" style={{ fontWeight: 700 }}>
        {t.footer.legal}
      </Link>
    </footer>
  )
}
