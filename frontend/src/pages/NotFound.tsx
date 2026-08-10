import { Link } from 'react-router-dom'
import { useLang } from '../i18n/lang'
import { Seo } from '../components/Seo'

export default function NotFound() {
  const { t } = useLang()
  return (
    <div style={{ maxWidth: 620, margin: '0 auto', padding: 'clamp(48px,8vw,96px) clamp(18px,5vw,64px)', textAlign: 'center' }}>
      <Seo title={t.notFound.title} path="/404" />
      <div className="anim-bob2" style={{ fontSize: 90 }}>
        🍨
      </div>
      <div className="display" style={{ fontWeight: 800, fontSize: 'clamp(56px,14vw,110px)', color: 'var(--cherry-strong)', lineHeight: 1 }}>
        {t.notFound.code}
      </div>
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(24px,4vw,34px)', margin: '6px 0 0' }}>
        {t.notFound.title}
      </h1>
      <p style={{ color: 'var(--ink-soft)', margin: '10px 0 0', maxWidth: '40ch', marginLeft: 'auto', marginRight: 'auto' }}>{t.notFound.body}</p>
      <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap', marginTop: 28 }}>
        <Link to="/" className="btn-primary" style={{ padding: '13px 22px' }}>
          {t.notFound.home}
        </Link>
        <Link to="/projects" className="btn-secondary" style={{ padding: '13px 22px' }}>
          {t.notFound.projects}
        </Link>
        <Link to="/arcade" className="btn-secondary" style={{ padding: '13px 22px' }}>
          {t.notFound.arcade}
        </Link>
      </div>
      <p className="mono" style={{ color: 'var(--ink-soft)', fontSize: 12, marginTop: 30, opacity: 0.7 }}>
        {t.notFound.egg}
      </p>
    </div>
  )
}
