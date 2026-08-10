import { useState } from 'react'
import type { Skill } from '../../api/types'
import { createSkill, deleteSkill, reorderSkills, updateSkill, useAdminSkills } from '../adminApi'
import { AdminCard, BilingualField, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { Spinner } from '../../components/states'
import { IconPicker } from '../IconPicker'

export default function SkillsAdmin() {
  const { data, isLoading } = useAdminSkills()
  const [nameCs, setNameCs] = useState('')
  const [nameEn, setNameEn] = useState('')
  const [category, setCategory] = useState('')
  const [level, setLevel] = useState('')
  const [icon, setIcon] = useState('')

  async function add() {
    const ok = await runAction(
      () => createSkill({ name_cs: nameCs, name_en: nameEn, category, icon: icon || undefined, level: level ? Number(level) : undefined }),
      'Přidáno / Added',
    )
    if (ok) {
      setNameCs('')
      setNameEn('')
      setCategory('')
      setLevel('')
      setIcon('')
    }
  }

  const skills = data ?? []
  const move = (i: number, dir: -1 | 1) => {
    const next = [...skills]
    const j = i + dir
    if (j < 0 || j >= next.length) return
    ;[next[i], next[j]] = [next[j], next[i]]
    void runAction(() => reorderSkills(next.map((s) => s.id)))
  }

  return (
    <div style={{ maxWidth: 820 }}>
      <PageHeader title="Dovednosti / Skills" />
      <AdminCard style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <BilingualField label="Název / Name" cs={nameCs} en={nameEn} onCs={setNameCs} onEn={setNameEn} />
          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr', gap: 8, alignItems: 'end' }}>
            <Field label="Skupina / Group">
              <input className="input" placeholder="languages" value={category} onChange={(e) => setCategory(e.target.value)} />
            </Field>
            <Field label="Úroveň 1–5">
              <input className="input" type="number" min={1} max={5} value={level} onChange={(e) => setLevel(e.target.value)} />
            </Field>
            {/* Not <Field>: its <label> would forward backdrop clicks back to the
                picker's trigger, so the popover couldn't be dismissed by clicking out. */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
              <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-soft)' }}>Ikona / Icon</span>
              <IconPicker value={icon} onChange={setIcon} />
            </div>
          </div>
          <button className="btn-primary" style={{ height: 42, alignSelf: 'flex-start', padding: '0 20px' }} onClick={add}>
            + Přidat / Add
          </button>
        </div>
      </AdminCard>

      {isLoading && <Spinner />}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {skills.map((s, i) => (
          <Row key={s.id} s={s} onUp={() => move(i, -1)} onDown={() => move(i, 1)} />
        ))}
      </div>
    </div>
  )
}

function Row({ s, onUp, onDown }: { s: Skill; onUp: () => void; onDown: () => void }) {
  const [nameCs, setNameCs] = useState(s.name_cs)
  const [nameEn, setNameEn] = useState(s.name_en)
  const [category, setCategory] = useState(s.category)
  const [level, setLevel] = useState(s.level != null ? String(s.level) : '')
  const [icon, setIcon] = useState(s.icon ?? '')
  const [visible, setVisible] = useState(s.visible)
  return (
    <AdminCard>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <BilingualField label="Název / Name" cs={nameCs} en={nameEn} onCs={setNameCs} onEn={setNameEn} />
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr', gap: 8 }}>
          <input className="input" value={category} onChange={(e) => setCategory(e.target.value)} />
          <input className="input" type="number" min={1} max={5} value={level} onChange={(e) => setLevel(e.target.value)} />
          <IconPicker value={icon} onChange={setIcon} />
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
            onClick={() =>
              runAction(
                () => updateSkill(s.id, { name_cs: nameCs, name_en: nameEn, category, icon, visible, level: level ? Number(level) : null }),
                'Uloženo / Saved',
              )
            }
          >
            Uložit / Save
          </button>
          <ConfirmButton small label="Smazat / Delete" onConfirm={() => runAction(() => deleteSkill(s.id), 'Smazáno / Deleted')} />
        </div>
      </div>
    </AdminCard>
  )
}
