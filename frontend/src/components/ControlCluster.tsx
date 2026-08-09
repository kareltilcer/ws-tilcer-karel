import { useTheme } from '../lib/theme'
import { useSound } from '../lib/sound'
import { useLang } from '../i18n/lang'

// The persistent control cluster: language toggle + theme toggle + sound mute.
// Always reachable in the header so a motion-sensitive or bilingual visitor finds
// it fast. (Calm mode was removed; prefers-reduced-motion handles motion safety.)
export function ControlCluster({ compact = false }: { compact?: boolean }) {
  const { lang, toggle: toggleLang } = useLang()
  const { theme, toggle: toggleTheme } = useTheme()
  const { muted, toggle: toggleSound, play } = useSound()
  const h = compact ? 38 : 42

  const base: React.CSSProperties = {
    height: h,
    border: '2px solid var(--cream-line)',
    background: 'var(--surface)',
    color: 'var(--ink)',
    borderRadius: 13,
    cursor: 'pointer',
    boxShadow: 'var(--shadow)',
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <button
        onClick={() => {
          play(660, 0.07, 'triangle')
          toggleLang()
        }}
        title="Language / Jazyk"
        aria-label="Toggle language"
        className="display"
        style={{ ...base, padding: '0 13px', fontWeight: 700, fontSize: 14 }}
      >
        {lang === 'cs' ? 'EN' : 'CS'}
      </button>
      <button
        onClick={() => {
          play(440, 0.09, 'sine')
          toggleTheme()
        }}
        title="Theme / Motiv"
        aria-label="Toggle theme"
        style={{ ...base, width: h, fontSize: 17 }}
      >
        {theme === 'day' ? '🌙' : '☀️'}
      </button>
      <button
        onClick={() => {
          toggleSound()
          if (muted) play(700, 0.12, 'triangle')
        }}
        title="Sound / Zvuk"
        aria-label="Toggle sound"
        style={{
          ...base,
          width: h,
          fontSize: 15,
          borderColor: muted ? 'var(--cream-line)' : 'var(--pistachio)',
          background: muted ? 'var(--surface)' : 'color-mix(in srgb, var(--pistachio) 30%, var(--surface))',
        }}
      >
        {muted ? '🔇' : '🔊'}
      </button>
    </div>
  )
}
