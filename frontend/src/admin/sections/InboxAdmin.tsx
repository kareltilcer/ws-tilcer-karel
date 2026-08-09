import { useState } from 'react'
import type { MessageStatus } from '../../api/types'
import { deleteMessage, setMessageStatus, useAdminContact } from '../adminApi'
import { AdminCard, ConfirmButton, PageHeader, runAction } from '../ui'
import { Spinner, CenterState } from '../../components/states'
import { useLang } from '../../i18n/lang'
import { formatDateTime } from '../../i18n/format'

const STATUSES: (MessageStatus | '')[] = ['', 'new', 'read', 'archived', 'spam']

export default function InboxAdmin() {
  const { lang } = useLang()
  const [filter, setFilter] = useState<MessageStatus | ''>('')
  const { data, isLoading } = useAdminContact(filter || undefined)

  return (
    <div style={{ maxWidth: 860 }}>
      <PageHeader title="Zprávy / Inbox" />
      <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
        {STATUSES.map((s) => (
          <button
            key={s || 'all'}
            onClick={() => setFilter(s)}
            style={{
              padding: '6px 14px',
              borderRadius: 999,
              fontWeight: 700,
              fontSize: 13,
              cursor: 'pointer',
              border: filter === s ? 'none' : '2px solid var(--cream-line)',
              background: filter === s ? 'var(--cherry-strong)' : 'var(--surface-2)',
              color: filter === s ? '#fff' : 'var(--ink)',
            }}
          >
            {s || 'Vše / All'}
          </button>
        ))}
      </div>

      {isLoading && <Spinner />}
      {data && data.items.length === 0 && <CenterState emoji="📭" title="Žádné zprávy / No messages" />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {data?.items.map((m) => (
          <AdminCard key={m.id}>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, flexWrap: 'wrap' }}>
              <div>
                <strong>{m.name}</strong>{' '}
                <a href={`mailto:${m.email}`} style={{ fontSize: 14 }}>
                  {m.email}
                </a>
                {m.subject && <div style={{ fontWeight: 700, marginTop: 4 }}>{m.subject}</div>}
              </div>
              <span className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)' }}>
                {formatDateTime(lang, m.created_at)} · {m.status}
              </span>
            </div>
            <p style={{ whiteSpace: 'pre-wrap', margin: '10px 0', color: 'var(--ink)' }}>{m.message}</p>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {(['read', 'archived', 'spam', 'new'] as MessageStatus[]).map((st) => (
                <button
                  key={st}
                  className="btn-secondary"
                  style={{ height: 32, padding: '0 12px', fontSize: 13, opacity: m.status === st ? 0.5 : 1 }}
                  disabled={m.status === st}
                  onClick={() => runAction(() => setMessageStatus(m.id, st), 'Aktualizováno / Updated')}
                >
                  → {st}
                </button>
              ))}
              <div style={{ flex: 1 }} />
              <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteMessage(m.id), 'Smazáno / Deleted')} />
            </div>
          </AdminCard>
        ))}
      </div>
    </div>
  )
}
