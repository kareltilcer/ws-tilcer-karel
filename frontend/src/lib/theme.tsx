import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'

export type Theme = 'day' | 'night'

interface ThemeCtx {
  theme: Theme
  toggle: () => void
}

const Ctx = createContext<ThemeCtx | null>(null)
const KEY = 'karel.theme'

function initial(): Theme {
  const saved = localStorage.getItem(KEY)
  if (saved === 'day' || saved === 'night') return saved
  // First visit: honour the OS preference; day is the hero otherwise.
  if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) return 'night'
  return 'day'
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<Theme>(initial)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem(KEY, theme)
  }, [theme])

  return (
    <Ctx.Provider value={{ theme, toggle: () => setTheme((t) => (t === 'day' ? 'night' : 'day')) }}>
      {children}
    </Ctx.Provider>
  )
}

export function useTheme(): ThemeCtx {
  const c = useContext(Ctx)
  if (!c) throw new Error('useTheme outside ThemeProvider')
  return c
}
