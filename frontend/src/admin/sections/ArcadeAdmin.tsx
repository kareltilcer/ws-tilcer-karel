import { useState } from 'react'
import { deleteScore, resetBoard, useAdminScores } from '../adminApi'
import { AdminCard, ConfirmButton, PageHeader, runAction } from '../ui'
import { Spinner, CenterState } from '../../components/states'
import { useLang } from '../../i18n/lang'
import { formatDate } from '../../i18n/format'

const GAMES = [
  { key: 'catch-the-scoop', name: 'Catch the Scoop' },
  { key: 'scoop-match', name: 'Scoop Match' },
]

export default function ArcadeAdmin() {
  const [game, setGame] = useState(GAMES[0].key)
  const { lang } = useLang()
  const { data, isLoading } = useAdminScores(game)

  return (
    <div style={{ maxWidth: 720 }}>
      <PageHeader
        title="Arkáda — moderace / Arcade moderation"
        action={
          <ConfirmButton label="Vymazat žebříček / Reset board" onConfirm={() => runAction(() => resetBoard(game), 'Vymazáno / Cleared')} />
        }
      />
      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        {GAMES.map((g) => (
          <button
            key={g.key}
            onClick={() => setGame(g.key)}
            style={{
              padding: '8px 16px',
              borderRadius: 999,
              fontWeight: 700,
              fontSize: 14,
              cursor: 'pointer',
              border: game === g.key ? 'none' : '2px solid var(--cream-line)',
              background: game === g.key ? 'var(--cherry-strong)' : 'var(--surface-2)',
              color: game === g.key ? '#fff' : 'var(--ink)',
            }}
          >
            {g.name}
          </button>
        ))}
      </div>

      {isLoading && <Spinner />}
      {data && data.length === 0 && <CenterState emoji="🍨" title="Žádné skóre / No scores" />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {data?.map((s, i) => (
          <AdminCard key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px' }}>
            <span className="mono" style={{ width: 28, color: 'var(--ink-soft)' }}>
              {i + 1}
            </span>
            <span style={{ flex: 1, fontWeight: 700 }}>{s.player_name}</span>
            <span className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)' }}>
              {formatDate(lang, s.created_at)}
            </span>
            <span className="mono" style={{ fontWeight: 700, color: 'var(--cherry-strong)', minWidth: 56, textAlign: 'right' }}>
              {s.score}
            </span>
            <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteScore(s.id, game), 'Smazáno / Deleted')} />
          </AdminCard>
        ))}
      </div>
    </div>
  )
}
