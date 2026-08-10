import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { Lang } from './plural'
import { STRINGS, type Strings } from './strings'

interface LangCtx {
  lang: Lang
  setLang: (l: Lang) => void
  toggle: () => void
  t: Strings
  /** pick returns the value for the active language, falling back to the other
   *  when one side is empty (bilingual content contract). */
  pick: (cs: string | null | undefined, en: string | null | undefined) => string
}

const Ctx = createContext<LangCtx | null>(null)
const KEY = 'karel.lang'

function initial(): Lang {
  const saved = localStorage.getItem(KEY)
  if (saved === 'cs' || saved === 'en') return saved
  // First paint may follow Accept-Language (via navigator); Czech is the default.
  const nav = navigator.language?.toLowerCase() ?? ''
  return nav.startsWith('en') ? 'en' : 'cs'
}

export function LangProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(initial)

  useEffect(() => {
    localStorage.setItem(KEY, lang)
    document.documentElement.setAttribute('lang', lang)
  }, [lang])

  const setLang = (l: Lang) => setLangState(l)
  const toggle = () => setLangState((l) => (l === 'cs' ? 'en' : 'cs'))
  const pick = (cs: string | null | undefined, en: string | null | undefined) => {
    const primary = lang === 'cs' ? cs : en
    const fallback = lang === 'cs' ? en : cs
    return (primary && primary.trim() ? primary : fallback) ?? ''
  }

  return <Ctx.Provider value={{ lang, setLang, toggle, t: STRINGS[lang], pick }}>{children}</Ctx.Provider>
}

export function useLang(): LangCtx {
  const c = useContext(Ctx)
  if (!c) throw new Error('useLang outside LangProvider')
  return c
}
