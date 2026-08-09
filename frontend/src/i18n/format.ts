import type { Lang } from './plural'

// Czech dates render as "d. M. yyyy" (5. 8. 2026), English as "5 Aug 2026".
const EN_MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

export function formatDate(lang: Lang, value: string | null | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const day = d.getDate()
  const month = d.getMonth()
  const year = d.getFullYear()
  if (lang === 'cs') return `${day}. ${month + 1}. ${year}`
  return `${day} ${EN_MONTHS[month]} ${year}`
}

export function formatDateTime(lang: Lang, value: string | null | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${formatDate(lang, value)} ${hh}:${mm}`
}
