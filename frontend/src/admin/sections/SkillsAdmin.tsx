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

  // Group skills by category in first-appearance order (matches the public page).
  // Each group's items keep their relative order; category order follows the
  // order groups are first seen in the flat, sort_order-driven list.
  const groups: { category: string; items: Skill[] }[] = []
  const groupIndex = new Map<string, number>()
  for (const s of skills) {
    let gi = groupIndex.get(s.category)
    if (gi === undefined) {
      gi = groups.length
      groupIndex.set(s.category, gi)
      groups.push({ category: s.category, items: [] })
    }
    groups[gi].items.push(s)
  }

  // Flatten groups (in their current order, groups kept contiguous) to a flat id
  // list and persist. Contiguity is what keeps the public page's groups clean.
  const persist = (gs: { category: string; items: Skill[] }[]) =>
    void runAction(() => reorderSkills(gs.flatMap((g) => g.items.map((s) => s.id))))

  const clone = () => groups.map((g) => ({ category: g.category, items: [...g.items] }))

  const moveGroup = (gi: number, dir: -1 | 1) => {
    const j = gi + dir
    if (j < 0 || j >= groups.length) return
    const next = clone()
    ;[next[gi], next[j]] = [next[j], next[gi]]
    persist(next)
  }

  const moveInGroup = (gi: number, i: number, dir: -1 | 1) => {
    const next = clone()
    const items = next[gi].items
    const j = i + dir
    if (j < 0 || j >= items.length) return
    ;[items[i], items[j]] = [items[j], items[i]]
    persist(next)
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
      <div style={{ display: 'flex', flexDirection: 'column', gap: 26 }}>
        {groups.map((g, gi) => (
          <div key={g.category}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 12 }}>
              <h2 className="display" style={{ fontWeight: 700, fontSize: 20, margin: 0, textTransform: 'capitalize' }}>
                {g.category}
              </h2>
              <span
                style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink-soft)', background: 'var(--cream-line)', borderRadius: 999, padding: '2px 9px' }}
              >
                {g.items.length}
              </span>
              <div style={{ flex: 1, height: 2, background: 'var(--cream-line)', borderRadius: 2 }} />
              <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink-soft)' }}>Skupina / Group</span>
              <button className="btn-secondary" style={arrowBtn(gi === 0)} disabled={gi === 0} onClick={() => moveGroup(gi, -1)}>
                ↑
              </button>
              <button
                className="btn-secondary"
                style={arrowBtn(gi === groups.length - 1)}
                disabled={gi === groups.length - 1}
                onClick={() => moveGroup(gi, 1)}
              >
                ↓
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {g.items.map((s, i) => (
                <Row
                  key={s.id}
                  s={s}
                  onUp={() => moveInGroup(gi, i, -1)}
                  onDown={() => moveInGroup(gi, i, 1)}
                  upDisabled={i === 0}
                  downDisabled={i === g.items.length - 1}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// Shared style for the ↑/↓ reorder buttons; dims + disables at a list boundary.
const arrowBtn = (disabled: boolean): React.CSSProperties => ({
  height: 32,
  padding: '0 10px',
  opacity: disabled ? 0.35 : 1,
  cursor: disabled ? 'default' : 'pointer',
})

function Row({
  s,
  onUp,
  onDown,
  upDisabled,
  downDisabled,
}: {
  s: Skill
  onUp: () => void
  onDown: () => void
  upDisabled: boolean
  downDisabled: boolean
}) {
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
          <button className="btn-secondary" style={arrowBtn(upDisabled)} disabled={upDisabled} onClick={onUp}>
            ↑
          </button>
          <button className="btn-secondary" style={arrowBtn(downDisabled)} disabled={downDisabled} onClick={onDown}>
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
