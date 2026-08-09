import { useLang } from '../i18n/lang'
import { useBusinessInfo } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton } from '../components/states'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 720, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

// The Czech OSVČ disclosure — Czech-primary, English alongside. The sídlo (home
// address) appears ONLY here; DIČ renders only when set (Karel is a neplátce).
export default function Legal() {
  const { t, pick } = useLang()
  const { data, isLoading } = useBusinessInfo()

  const seat = data
    ? [data.address_lines, [data.zip, data.city].filter(Boolean).join(' '), data.country].filter(Boolean).join(', ')
    : ''
  const register = pick(data?.register_note_cs, data?.register_note_en)

  return (
    <Wrap>
      <Seo title={t.legal.title} description={t.legal.sub} path="/legal" />
      <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(30px,5vw,46px)', margin: 0 }}>
        {t.legal.title}
      </h1>
      <p style={{ color: 'var(--ink-soft)', margin: '8px 0 24px' }}>{t.legal.sub}</p>

      {isLoading ? (
        <Skeleton height={280} />
      ) : (
        <div className="card" style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <div className="display" style={{ fontWeight: 800, fontSize: 26, marginBottom: 12 }}>
            {data?.legal_name || '—'}
          </div>
          <Row label={t.legal.ico} value={data?.ico} />
          {data?.dic ? <Row label={t.legal.dic} value={data.dic} /> : null}
          <Row label={t.legal.seat} value={seat} />
          {register ? <Row label={t.legal.register} value={register} /> : null}
          <Row label={t.legal.email} value={data?.contact_email} isEmail />
          {data?.contact_phone ? <Row label={t.legal.phone} value={data.contact_phone} /> : null}
          {pick(data?.data_note_cs, data?.data_note_en) && (
            <p style={{ marginTop: 16, color: 'var(--ink-soft)', fontSize: 14, lineHeight: 1.6 }}>{pick(data?.data_note_cs, data?.data_note_en)}</p>
          )}
        </div>
      )}
    </Wrap>
  )
}

function Row({ label, value, isEmail }: { label: string; value?: string | null; isEmail?: boolean }) {
  if (!value) return null
  return (
    <div style={{ display: 'flex', gap: 12, padding: '10px 0', borderTop: '1px dashed var(--cream-line)' }}>
      <span style={{ width: 160, flexShrink: 0, fontWeight: 700, color: 'var(--ink-soft)', fontSize: 14 }}>{label}</span>
      <span style={{ fontWeight: 600 }}>{isEmail ? <a href={`mailto:${value}`}>{value}</a> : value}</span>
    </div>
  )
}
