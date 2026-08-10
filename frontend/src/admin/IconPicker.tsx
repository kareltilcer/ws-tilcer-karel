import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { SkillIcon, SKILL_ICONS, ICON_GROUP_ORDER, ICON_GROUP_LABELS, getIconDef, isRegistryIcon, iconValue } from '../lib/skillIcons'

const POPOVER_WIDTH = 340

// A compact icon chooser: a trigger that previews the current icon, and an
// anchored popover with search + palette-tinted logos grouped by kind, plus an
// escape hatch to type any emoji. Stores a registry key (→ tinted SVG on the
// site) or a literal emoji; empty clears it.
export function IconPicker({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  // Whether the current value came from the palette (vs. typed as custom). Kept
  // as its own flag rather than derived from isRegistryIcon(value) so that
  // typing a string that happens to equal a registry key into the custom field
  // doesn't blank the field mid-keystroke.
  const [fromPalette, setFromPalette] = useState(() => isRegistryIcon(value))
  // The last value we emitted, so the effect below can tell our own edits (which
  // already keep fromPalette correct) apart from an external reset of `value`.
  const lastEmittedRef = useRef(value)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const popRef = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null)

  // Resync fromPalette only when `value` changes from outside (e.g. the row is
  // reset to its saved icon). Our own choosePalette/clear/custom-input handlers
  // set both onChange and fromPalette together and record lastEmittedRef, so
  // they're skipped here — preserving the mid-keystroke safety described above.
  const emit = (v: string) => {
    lastEmittedRef.current = v
    onChange(v)
  }
  useEffect(() => {
    if (value === lastEmittedRef.current) return
    lastEmittedRef.current = value
    setFromPalette(isRegistryIcon(value))
  }, [value])

  const keys = useMemo(() => Object.keys(SKILL_ICONS), [])
  const filtered = useMemo(() => {
    const s = q.trim().toLowerCase()
    if (!s) return keys
    return keys.filter((k) => k.includes(s) || SKILL_ICONS[k].name.toLowerCase().includes(s))
  }, [q, keys])

  const groups = ICON_GROUP_ORDER.map((g) => ({ g, items: filtered.filter((k) => SKILL_ICONS[k].group === g) })).filter(
    (x) => x.items.length > 0,
  )

  const label = value ? (getIconDef(value)?.name ?? value) : 'Vybrat / Choose'

  // Anchor the popover with `position: fixed` (viewport-relative) so the admin
  // <main>'s `overflow: auto` can't clip it on the last row. Recompute on
  // scroll/resize, and flip above the trigger when there isn't room below.
  useLayoutEffect(() => {
    if (!open) return
    const place = () => {
      const el = triggerRef.current
      if (!el) return
      const r = el.getBoundingClientRect()
      const left = Math.max(8, Math.min(r.left, window.innerWidth - POPOVER_WIDTH - 8))
      const ph = popRef.current?.offsetHeight ?? 0
      const belowTop = r.bottom + 6
      const fitsBelow = ph === 0 || belowTop + ph <= window.innerHeight - 8
      const top = fitsBelow ? belowTop : Math.max(8, r.top - 6 - ph)
      setPos({ top, left })
    }
    place()
    // Coalesce bursts of scroll/resize events into one reposition per frame,
    // so scrolling a long admin list while the picker is open doesn't fire a
    // setState-driven re-render on every event.
    let raf = 0
    const schedule = () => {
      if (raf) return
      raf = requestAnimationFrame(() => {
        raf = 0
        place()
      })
    }
    window.addEventListener('resize', schedule)
    window.addEventListener('scroll', schedule, true)
    return () => {
      if (raf) cancelAnimationFrame(raf)
      window.removeEventListener('resize', schedule)
      window.removeEventListener('scroll', schedule, true)
    }
    // filtered.length: re-place when a search changes the popover height, so a
    // flipped-above popover doesn't drift off-screen as results grow/shrink.
  }, [open, filtered.length])

  // Closing also clears the search so reopening starts from the full palette,
  // and returns focus to the trigger (keyboard users can't click the backdrop).
  function close() {
    setOpen(false)
    setQ('')
    triggerRef.current?.focus()
  }

  function choosePalette(k: string) {
    emit(iconValue(k))
    setFromPalette(true)
    close()
  }

  function clear() {
    emit('')
    setFromPalette(false)
    close()
  }

  return (
    <div style={{ position: 'relative' }}>
      <button
        ref={triggerRef}
        type="button"
        className="btn-secondary"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="dialog"
        aria-expanded={open}
        style={{ height: 46, width: '100%', padding: '0 12px', display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'flex-start' }}
      >
        <span style={{ display: 'grid', placeItems: 'center', width: 24, height: 24 }}>
          <SkillIcon icon={value} size={20} />
        </span>
        <span style={{ flex: 1, textAlign: 'left', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 14 }}>
          {label}
        </span>
        <span style={{ color: 'var(--ink-soft)' }}>▾</span>
      </button>

      {open && (
        <>
          <div onClick={close} style={{ position: 'fixed', inset: 0, zIndex: 40 }} />
          <div
            ref={popRef}
            className="card"
            role="dialog"
            aria-label="Výběr ikony / Icon picker"
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                e.stopPropagation()
                close()
              }
            }}
            style={{
              position: 'fixed',
              top: pos?.top ?? 0,
              left: pos?.left ?? 0,
              zIndex: 50,
              width: POPOVER_WIDTH,
              maxWidth: '86vw',
              padding: 12,
              borderRadius: 16,
              visibility: pos ? 'visible' : 'hidden',
            }}
          >
            <input
              className="input"
              autoFocus
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Hledat / Search…"
              style={{ height: 40, marginBottom: 10 }}
            />

            <div style={{ maxHeight: 300, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 12 }}>
              {groups.map(({ g, items }) => (
                <div key={g}>
                  <div style={{ fontWeight: 700, fontSize: 12, color: 'var(--ink-soft)', margin: '0 0 6px' }}>
                    {ICON_GROUP_LABELS[g].cs} / {ICON_GROUP_LABELS[g].en}
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(40px,1fr))', gap: 6 }}>
                    {items.map((k) => {
                      const active = value === iconValue(k)
                      return (
                        <button
                          key={k}
                          type="button"
                          title={SKILL_ICONS[k].name}
                          onClick={() => choosePalette(k)}
                          style={{
                            height: 40,
                            display: 'grid',
                            placeItems: 'center',
                            borderRadius: 10,
                            cursor: 'pointer',
                            border: `2px solid ${active ? 'var(--cherry-strong)' : 'var(--cream-line)'}`,
                            background: active ? 'color-mix(in srgb, var(--cherry-strong) 12%, var(--surface))' : 'var(--surface)',
                          }}
                        >
                          <SkillIcon icon={iconValue(k)} size={22} />
                        </button>
                      )
                    })}
                  </div>
                </div>
              ))}
              {groups.length === 0 && (
                <p style={{ color: 'var(--ink-soft)', fontSize: 13, margin: '4px 0' }}>Nic nenalezeno / No matches</p>
              )}
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 12, paddingTop: 10, borderTop: '2px solid var(--cream-line)' }}>
              <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink-soft)', whiteSpace: 'nowrap' }}>Emoji / vlastní:</span>
              <input
                className="input"
                value={fromPalette ? '' : value}
                onChange={(e) => {
                  setFromPalette(false)
                  emit(e.target.value)
                }}
                placeholder="🧶"
                style={{ height: 38, flex: 1 }}
              />
              <button
                type="button"
                className="btn-secondary"
                onClick={clear}
                style={{ height: 38, padding: '0 12px', fontSize: 13 }}
              >
                Vyčistit / Clear
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
