import { useState } from 'react'
import type { Skill } from '../../api/types'
import { createSkill, deleteSkill, reorderSkills, updateSkill, useAdminSkills } from '../adminApi'
import { AdminCard, BilingualField, ConfirmButton, Field, PageHeader, runAction } from '../ui'
import { Spinner } from '../../components/states'
import { IconPicker } from '../IconPicker'

// Skills group by their bilingual label pair: two skills sit in the same group
// when both category_cs and category_en match. This composite key threads through
// grouping, React keys, and reorder — the analog of the old single category string.
const groupKey = (s: { category_cs: string; category_en: string }) => `${s.category_cs} ${s.category_en}`

export default function SkillsAdmin() {
  const { data, isLoading } = useAdminSkills()
  const [nameCs, setNameCs] = useState('')
  const [nameEn, setNameEn] = useState('')
  const [categoryCs, setCategoryCs] = useState('')
  const [categoryEn, setCategoryEn] = useState('')
  const [level, setLevel] = useState('')
  const [icon, setIcon] = useState('')

  async function add() {
    const ok = await runAction(
      () =>
        createSkill({
          name_cs: nameCs,
          name_en: nameEn,
          category_cs: categoryCs,
          category_en: categoryEn,
          icon: icon || undefined,
          level: level ? Number(level) : undefined,
        }),
      'Přidáno / Added',
    )
    if (ok) {
      setNameCs('')
      setNameEn('')
      setCategoryCs('')
      setCategoryEn('')
      setLevel('')
      setIcon('')
    }
  }

  const skills = data ?? []

  // Group skills by their bilingual label in first-appearance order (matches the
  // public page). Each group's items keep their relative order; group order
  // follows the order groups are first seen in the flat, sort_order-driven list.
  const groups: { cs: string; en: string; items: Skill[] }[] = []
  const groupIndex = new Map<string, number>()
  for (const s of skills) {
    const key = groupKey(s)
    let gi = groupIndex.get(key)
    if (gi === undefined) {
      gi = groups.length
      groupIndex.set(key, gi)
      groups.push({ cs: s.category_cs, en: s.category_en, items: [] })
    }
    groups[gi].items.push(s)
  }

  // Flatten groups (in their current order, groups kept contiguous) to a flat id
  // list and persist. Contiguity is what keeps the public page's groups clean.
  const persist = (gs: { cs: string; en: string; items: Skill[] }[]) =>
    void runAction(() => reorderSkills(gs.flatMap((g) => g.items.map((s) => s.id))))

  const clone = () => groups.map((g) => ({ cs: g.cs, en: g.en, items: [...g.items] }))

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
          <BilingualField label="Skupina / Group" cs={categoryCs} en={categoryEn} onCs={setCategoryCs} onEn={setCategoryEn} />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, alignItems: 'end' }}>
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
          <div key={`${g.cs} ${g.en}`}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 12 }}>
              <h2 className="display" style={{ fontWeight: 700, fontSize: 20, margin: 0, textTransform: 'capitalize' }}>
                {g.cs}
                {g.en && <span style={{ color: 'var(--ink-soft)', fontWeight: 600 }}> / {g.en}</span>}
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
  const [categoryCs, setCategoryCs] = useState(s.category_cs)
  const [categoryEn, setCategoryEn] = useState(s.category_en)
  const [level, setLevel] = useState(s.level != null ? String(s.level) : '')
  const [icon, setIcon] = useState(s.icon ?? '')
  const [visible, setVisible] = useState(s.visible)
  return (
    <AdminCard>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <BilingualField label="Název / Name" cs={nameCs} en={nameEn} onCs={setNameCs} onEn={setNameEn} />
        <BilingualField label="Skupina / Group" cs={categoryCs} en={categoryEn} onCs={setCategoryCs} onEn={setCategoryEn} />
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
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
                () =>
                  updateSkill(s.id, {
                    name_cs: nameCs,
                    name_en: nameEn,
                    category_cs: categoryCs,
                    category_en: categoryEn,
                    icon,
                    visible,
                    level: level ? Number(level) : null,
                  }),
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
