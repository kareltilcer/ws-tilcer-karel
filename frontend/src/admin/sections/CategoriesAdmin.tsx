import { useState } from 'react'
import type { Category } from '../../api/types'
import { createCategory, deleteCategory, reorderCategories, updateCategory, useAdminCategories } from '../adminApi'
import { AdminCard, BilingualField, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { Spinner } from '../../components/states'

export default function CategoriesAdmin() {
  const { data, isLoading } = useAdminCategories()
  const [slug, setSlug] = useState('')
  const [nameCs, setNameCs] = useState('')
  const [nameEn, setNameEn] = useState('')

  async function add() {
    const ok = await runAction(() => createCategory({ slug, name_cs: nameCs, name_en: nameEn }), 'Přidáno / Added')
    if (ok) {
      setSlug('')
      setNameCs('')
      setNameEn('')
    }
  }

  const cats = data ?? []
  const move = (i: number, dir: -1 | 1) => {
    const next = [...cats]
    const j = i + dir
    if (j < 0 || j >= next.length) return
    ;[next[i], next[j]] = [next[j], next[i]]
    void runAction(() => reorderCategories(next.map((c) => c.id)))
  }

  return (
    <div style={{ maxWidth: 820 }}>
      <PageHeader title="Kategorie / Categories" />
      <AdminCard style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Field label="Slug">
            <input className="input" placeholder="software" value={slug} onChange={(e) => setSlug(e.target.value)} />
          </Field>
          <BilingualField label="Název / Name" cs={nameCs} en={nameEn} onCs={setNameCs} onEn={setNameEn} />
          <button className="btn-primary" style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }} onClick={add}>
            + Přidat / Add
          </button>
        </div>
      </AdminCard>

      {isLoading && <Spinner />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {cats.map((c, i) => (
          <Row key={c.id} c={c} onUp={() => move(i, -1)} onDown={() => move(i, 1)} />
        ))}
      </div>
    </div>
  )
}

function Row({ c, onUp, onDown }: { c: Category; onUp: () => void; onDown: () => void }) {
  const [slug, setSlug] = useState(c.slug)
  const [nameCs, setNameCs] = useState(c.name_cs)
  const [nameEn, setNameEn] = useState(c.name_en)
  const [visible, setVisible] = useState(c.visible)
  return (
    <AdminCard>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
          <input className="input" value={slug} onChange={(e) => setSlug(e.target.value)} />
          <input className="input" value={nameCs} onChange={(e) => setNameCs(e.target.value)} />
          <input className="input" value={nameEn} onChange={(e) => setNameEn(e.target.value)} />
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
          <label style={{ display: 'flex', gap: 6, alignItems: 'center', fontSize: 14 }}>
            <input type="checkbox" checked={visible} onChange={(e) => setVisible(e.target.checked)} /> Viditelné / Visible
          </label>
          <div style={{ flex: 1 }} />
          <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={onUp}>
            ↑
          </button>
          <button className="btn-secondary" style={{ height: 32, padding: '0 10px' }} onClick={onDown}>
            ↓
          </button>
          <button
            className="btn-primary"
            style={{ height: 32, padding: '0 14px', fontSize: 13 }}
            onClick={() => runAction(() => updateCategory(c.id, { slug, name_cs: nameCs, name_en: nameEn, visible }), 'Uloženo / Saved')}
          >
            Uložit / Save
          </button>
          <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteCategory(c.id), 'Smazáno / Deleted')} />
        </div>
      </div>
    </AdminCard>
  )
}
