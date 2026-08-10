import { createContext, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from 'react'

// Opt-in, off by default (never autoplay). A small WebAudio blip layer used by
// the hero and the arcade; when muted, play() is a no-op.
interface SoundCtx {
  muted: boolean
  toggle: () => void
  play: (freq: number, dur?: number, type?: OscillatorType, vol?: number) => void
}

const Ctx = createContext<SoundCtx | null>(null)
const KEY = 'karel.sound'

export function SoundProvider({ children }: { children: ReactNode }) {
  const [muted, setMuted] = useState<boolean>(() => localStorage.getItem(KEY) !== 'on')
  const acRef = useRef<AudioContext | null>(null)
  // play() reads muted through a ref so its identity stays stable across toggles.
  // (A play() whose identity changed on every mute toggle would tear down and
  // restart any effect that depends on it — e.g. the arcade game engine.)
  const mutedRef = useRef(muted)
  mutedRef.current = muted

  const play = useCallback<SoundCtx['play']>((freq, dur = 0.15, type = 'sine', vol = 0.15) => {
    if (mutedRef.current) return
    if (!acRef.current) {
      const AC = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
      try {
        acRef.current = new AC()
      } catch {
        return
      }
    }
    const ctx = acRef.current
    if (!ctx) return
    try {
      const o = ctx.createOscillator()
      const g = ctx.createGain()
      o.type = type
      o.frequency.value = freq
      o.connect(g)
      g.connect(ctx.destination)
      const t = ctx.currentTime
      g.gain.setValueAtTime(0.0001, t)
      g.gain.exponentialRampToValueAtTime(vol, t + 0.01)
      g.gain.exponentialRampToValueAtTime(0.0001, t + dur)
      o.start(t)
      o.stop(t + dur)
    } catch {
      /* ignore */
    }
  }, [])

  const toggle = useCallback(() => {
    setMuted((m) => {
      const next = !m
      localStorage.setItem(KEY, next ? 'off' : 'on')
      return next
    })
  }, [])

  const value = useMemo<SoundCtx>(() => ({ muted, toggle, play }), [muted, toggle, play])

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useSound(): SoundCtx {
  const c = useContext(Ctx)
  if (!c) throw new Error('useSound outside SoundProvider')
  return c
}
