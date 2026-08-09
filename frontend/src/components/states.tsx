import { useLang } from '../i18n/lang'

export function Skeleton({ height = 120, style }: { height?: number | string; style?: React.CSSProperties }) {
  return <div className="skeleton" style={{ height, width: '100%', ...style }} aria-hidden />
}

export function Spinner({ size = 28 }: { size?: number }) {
  return (
    <span
      className="spin-fast"
      style={{
        display: 'inline-block',
        width: size,
        height: size,
        borderRadius: '50%',
        border: '3px solid var(--cream-line)',
        borderTopColor: 'var(--cherry-strong)',
      }}
      role="status"
      aria-label="loading"
    />
  )
}

export function CenterState({
  emoji,
  title,
  body,
  action,
}: {
  emoji: string
  title: string
  body?: string
  action?: React.ReactNode
}) {
  return (
    <div
      className="card"
      style={{
        textAlign: 'center',
        padding: '48px 24px',
        borderStyle: 'dashed',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 8,
      }}
    >
      <div style={{ fontSize: 48 }}>{emoji}</div>
      <h3 className="display" style={{ fontWeight: 800, fontSize: 22, margin: 0, color: 'var(--ink)' }}>
        {title}
      </h3>
      {body && <p style={{ color: 'var(--ink-soft)', margin: 0, maxWidth: '40ch' }}>{body}</p>}
      {action && <div style={{ marginTop: 12 }}>{action}</div>}
    </div>
  )
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  const { t } = useLang()
  return (
    <CenterState
      emoji="🫠"
      title={message}
      action={
        onRetry ? (
          <button className="btn-secondary" style={{ height: 44, padding: '0 20px' }} onClick={onRetry}>
            {t.common.retry}
          </button>
        ) : undefined
      }
    />
  )
}
