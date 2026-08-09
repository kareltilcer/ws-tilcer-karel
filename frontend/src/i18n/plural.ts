// Czech has three plural forms (1 / 2–4 / 5+). Every count label runs through
// count(). English uses the usual singular/plural pair.
export type Lang = 'cs' | 'en'

/** [one, few, many] for Czech; [one, other] for English. */
export type CzForms = readonly [one: string, few: string, many: string]
export type EnForms = readonly [one: string, other: string]

export function czPlural(n: number, f: CzForms): string {
  if (n === 1) return f[0]
  if (n >= 2 && n <= 4) return f[1]
  return f[2]
}

export function enPlural(n: number, f: EnForms): string {
  return n === 1 ? f[0] : f[1]
}

/** count renders "n form" with the right plural for the language. */
export function count(lang: Lang, n: number, cs: CzForms, en: EnForms): string {
  return `${n} ${lang === 'cs' ? czPlural(n, cs) : enPlural(n, en)}`
}

// Shared plural triples/pairs used across screens.
export const PLURAL = {
  projects: { cs: ['projekt', 'projekty', 'projektů'] as CzForms, en: ['project', 'projects'] as EnForms },
  scores: { cs: ['skóre', 'skóre', 'skóre'] as CzForms, en: ['score', 'scores'] as EnForms },
  messages: { cs: ['zpráva', 'zprávy', 'zpráv'] as CzForms, en: ['message', 'messages'] as EnForms },
} as const
