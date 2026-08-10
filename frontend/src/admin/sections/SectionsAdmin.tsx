import { useEffect, useState } from 'react'
import type { Media } from '../../api/types'
import { saveSection, useSectionAdmin } from '../adminApi'
import { AdminCard, BilingualField, PageHeader, runAction } from '../ui'
import { MediaPicker } from '../MediaPicker'
import { Spinner } from '../../components/states'

export default function SectionsAdmin() {
  return (
    <div style={{ maxWidth: 820 }}>
      <PageHeader title="Sekce / Sections" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionEditor sectionKey="about" title="O mně / About" />
        <SectionEditor sectionKey="hero" title="Hero" />
      </div>
    </div>
  )
}

function SectionEditor({ sectionKey, title }: { sectionKey: string; title: string }) {
  const { data, isLoading } = useSectionAdmin(sectionKey)
  const [headingCs, setHeadingCs] = useState('')
  const [headingEn, setHeadingEn] = useState('')
  const [bodyCs, setBodyCs] = useState('')
  const [bodyEn, setBodyEn] = useState('')
  const [media, setMedia] = useState<{ id: number; public_url: string } | null>(null)

  useEffect(() => {
    if (!data) return
    setHeadingCs(data.heading_cs ?? '')
    setHeadingEn(data.heading_en ?? '')
    setBodyCs(data.body_cs ?? '')
    setBodyEn(data.body_en ?? '')
    setMedia(data.media ? { id: data.media.id, public_url: data.media.public_url } : null)
  }, [data])

  if (isLoading) return <AdminCard><Spinner /></AdminCard>

  return (
    <AdminCard>
      <h2 className="display" style={{ fontWeight: 700, fontSize: 20, margin: '0 0 14px' }}>
        {title}
      </h2>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <BilingualField label="Nadpis / Heading" cs={headingCs} en={headingEn} onCs={setHeadingCs} onEn={setHeadingEn} />
        <BilingualField label="Text (Markdown)" cs={bodyCs} en={bodyEn} onCs={setBodyCs} onEn={setBodyEn} textarea />
        <MediaPicker
          label="Foto / Media"
          selected={media}
          onPick={(m: Media) => setMedia({ id: m.id, public_url: m.public_url })}
          onClear={() => setMedia(null)}
        />
        <button
          className="btn-primary"
          style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }}
          onClick={() =>
            runAction(
              () =>
                saveSection(sectionKey, {
                  heading_cs: headingCs,
                  heading_en: headingEn,
                  body_cs: bodyCs,
                  body_en: bodyEn,
                  media_id: media ? media.id : null,
                }),
              'Uloženo / Saved',
            )
          }
        >
          Uložit / Save
        </button>
      </div>
    </AdminCard>
  )
}
