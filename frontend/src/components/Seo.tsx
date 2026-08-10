import { useEffect } from 'react'

const CANONICAL_BASE = 'https://karel.tilcer.cz'

function setMeta(attr: 'name' | 'property', key: string, content: string) {
  let el = document.head.querySelector<HTMLMetaElement>(`meta[${attr}="${key}"]`)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

function setCanonical(href: string) {
  let el = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!el) {
    el = document.createElement('link')
    el.rel = 'canonical'
    document.head.appendChild(el)
  }
  el.href = href
}

/** Seo sets per-route title, description, canonical, and Open Graph/Twitter tags
 *  client-side (v1 SEO; SSR is a documented future rework when the blog lands). */
export function Seo({ title, description, path = '/' }: { title: string; description?: string; path?: string }) {
  useEffect(() => {
    const full = `${title} · karel`
    document.title = full
    const url = CANONICAL_BASE + path
    setCanonical(url)
    if (description) setMeta('name', 'description', description)
    setMeta('property', 'og:title', full)
    setMeta('property', 'og:type', 'website')
    setMeta('property', 'og:url', url)
    if (description) setMeta('property', 'og:description', description)
    setMeta('property', 'og:image', `${CANONICAL_BASE}/og.svg`)
    setMeta('name', 'twitter:card', 'summary_large_image')
  }, [title, description, path])
  return null
}
