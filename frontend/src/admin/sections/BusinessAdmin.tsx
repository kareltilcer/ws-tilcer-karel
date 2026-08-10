import { useEffect, useState } from 'react'
import { saveBusiness, useBusinessAdmin } from '../adminApi'
import { AdminCard, BilingualField, Field, PageHeader, runAction } from '../ui'
import { Spinner } from '../../components/states'

const EMPTY = {
  legal_name: '',
  ico: '',
  dic: '',
  address_lines: '',
  city: '',
  zip: '',
  country: '',
  register_note_cs: '',
  register_note_en: '',
  contact_email: '',
  contact_phone: '',
  data_note_cs: '',
  data_note_en: '',
}

export default function BusinessAdmin() {
  const { data, isLoading } = useBusinessAdmin()
  const [f, setF] = useState(EMPTY)
  const set = (k: keyof typeof EMPTY) => (v: string) => setF((prev) => ({ ...prev, [k]: v }))

  useEffect(() => {
    if (!data) return
    setF({
      legal_name: data.legal_name ?? '',
      ico: data.ico ?? '',
      dic: data.dic ?? '',
      address_lines: data.address_lines ?? '',
      city: data.city ?? '',
      zip: data.zip ?? '',
      country: data.country ?? '',
      register_note_cs: data.register_note_cs ?? '',
      register_note_en: data.register_note_en ?? '',
      contact_email: data.contact_email ?? '',
      contact_phone: data.contact_phone ?? '',
      data_note_cs: data.data_note_cs ?? '',
      data_note_en: data.data_note_en ?? '',
    })
  }, [data])

  if (isLoading)
    return (
      <AdminCard>
        <Spinner />
      </AdminCard>
    )

  return (
    <div style={{ maxWidth: 760 }}>
      <PageHeader title="Firemní údaje / Business (OSVČ)" />
      <AdminCard>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Field label="Jméno a příjmení / Legal name">
            <input className="input" value={f.legal_name} onChange={(e) => set('legal_name')(e.target.value)} />
          </Field>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="IČO">
              <input className="input" value={f.ico} onChange={(e) => set('ico')(e.target.value)} />
            </Field>
            <Field label="DIČ (skryté, pokud prázdné)">
              <input className="input" value={f.dic} onChange={(e) => set('dic')(e.target.value)} />
            </Field>
          </div>
          <Field label="Sídlo — adresa (jen na této stránce) / Seat (legal page only)">
            <input className="input" value={f.address_lines} onChange={(e) => set('address_lines')(e.target.value)} />
          </Field>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
            <Field label="PSČ / ZIP">
              <input className="input" value={f.zip} onChange={(e) => set('zip')(e.target.value)} />
            </Field>
            <Field label="Město / City">
              <input className="input" value={f.city} onChange={(e) => set('city')(e.target.value)} />
            </Field>
            <Field label="Země / Country">
              <input className="input" value={f.country} onChange={(e) => set('country')(e.target.value)} />
            </Field>
          </div>
          <BilingualField
            label="Zápis v živnostenském rejstříku / Trade register note"
            cs={f.register_note_cs}
            en={f.register_note_en}
            onCs={set('register_note_cs')}
            onEn={set('register_note_en')}
            textarea
          />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="E-mail">
              <input className="input" value={f.contact_email} onChange={(e) => set('contact_email')(e.target.value)} />
            </Field>
            <Field label="Telefon / Phone">
              <input className="input" value={f.contact_phone} onChange={(e) => set('contact_phone')(e.target.value)} />
            </Field>
          </div>
          <BilingualField label="Poznámka / Note" cs={f.data_note_cs} en={f.data_note_en} onCs={set('data_note_cs')} onEn={set('data_note_en')} textarea />
          <button className="btn-primary" style={{ height: 44, alignSelf: 'flex-start', padding: '0 22px' }} onClick={() => runAction(() => saveBusiness(f), 'Uloženo / Saved')}>
            Uložit / Save
          </button>
        </div>
      </AdminCard>
    </div>
  )
}
