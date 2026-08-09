import { useLang } from '../i18n/lang'
import { useSection } from '../api/hooks'
import { Seo } from '../components/Seo'
import { Skeleton } from '../components/states'
import { MarkdownView } from '../components/MarkdownView'

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 820, margin: '0 auto', padding: 'clamp(28px,4vw,56px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

export default function About() {
  const { t, pick } = useLang()
  const { data, isLoading } = useSection('about')
  const heading = pick(data?.heading_cs, data?.heading_en) || t.about.title
  const body = pick(data?.body_cs, data?.body_en)

  return (
    <Wrap>
      <Seo title={t.about.title} description={body.slice(0, 150)} path="/about" />
      <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap', alignItems: 'flex-start' }}>
        {data?.media && (
          <img
            src={data.media.public_url}
            alt={data.media.alt_cs ?? data.media.alt_en ?? heading}
            style={{ width: 220, maxWidth: '100%', borderRadius: 24, border: '2px solid var(--cream-line)', boxShadow: 'var(--shadow)' }}
          />
        )}
        <div style={{ flex: 1, minWidth: 260 }}>
          <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(32px,5vw,52px)', margin: 0 }}>
            {heading}
          </h1>
          {isLoading ? (
            <Skeleton height={200} style={{ marginTop: 20 }} />
          ) : (
            <div style={{ marginTop: 18, fontSize: 17 }}>{body && <MarkdownView>{body}</MarkdownView>}</div>
          )}
        </div>
      </div>
    </Wrap>
  )
}
