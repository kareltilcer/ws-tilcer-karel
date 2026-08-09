import { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { useLang } from '../i18n/lang'
import { useTheme } from '../lib/theme'
import { useSound } from '../lib/sound'
import { Seo } from '../components/Seo'
import { coneSprite, popSprite, cupSprite, cssVar, prefersReducedMotion } from '../lib/sprites'

interface Item {
  el: HTMLDivElement
  w: number
  h: number
  x: number
  y: number
  vx: number
  vy: number
  ang: number
  va: number
  drag: boolean
  px?: number
  py?: number
  lx?: number
  ly?: number
}
interface Sprinkle {
  el: HTMLDivElement
  x: number
  y: number
  vx: number
  vy: number
  rot: number
  vr: number
  hits: number
}

export default function Home() {
  const { t } = useLang()
  const { theme } = useTheme()
  const { play } = useSound()
  const playRef = useRef<HTMLDivElement>(null)
  const glowRef = useRef<HTMLDivElement>(null)

  // Glow cursor (decorative; harmless under reduced-motion — it just follows).
  useEffect(() => {
    const g = glowRef.current
    if (!g) return
    const onMove = (e: PointerEvent) => {
      g.style.left = e.clientX + 'px'
      g.style.top = e.clientY + 'px'
      g.style.opacity = '0.9'
    }
    window.addEventListener('pointermove', onMove)
    return () => window.removeEventListener('pointermove', onMove)
  }, [])

  // Hero physics — grab-and-fling scoops + falling sprinkles. Rebuilt on theme
  // change so the sprites re-read the palette. Stilled under reduced-motion.
  useEffect(() => {
    const host = playRef.current
    if (!host) return
    host.innerHTML = ''
    const rect = host.getBoundingClientRect()
    let W = rect.width
    let H = rect.height
    const reduce = prefersReducedMotion()

    const defs = [
      { svg: coneSprite(cssVar('--mango'), cssVar('--yuzu')), w: 96, h: 132 },
      { svg: coneSprite(cssVar('--pistachio'), cssVar('--cherry')), w: 92, h: 126 },
      { svg: popSprite(cssVar('--poppy')), w: 70, h: 132 },
      { svg: popSprite(cssVar('--sky')), w: 66, h: 124 },
      { svg: cupSprite(cssVar('--cherry')), w: 98, h: 116 },
    ]
    const items: Item[] = []
    const cleanups: (() => void)[] = []

    defs.forEach((d, i) => {
      const el = document.createElement('div')
      el.innerHTML = d.svg
      el.style.cssText = `position:absolute;width:${d.w}px;height:${d.h}px;pointer-events:${reduce ? 'none' : 'auto'};cursor:${reduce ? 'default' : 'grab'};touch-action:none;filter:drop-shadow(0 10px 14px rgba(58,36,22,.28));will-change:transform`
      host.appendChild(el)
      const it: Item = {
        el,
        w: d.w,
        h: d.h,
        x: 20 + i * 0.19 * (W - d.w),
        y: 10 + (i % 2) * 70,
        vx: (Math.random() - 0.5) * 0.6,
        vy: 0,
        ang: (Math.random() - 0.5) * 18,
        va: (Math.random() - 0.5) * 1.2,
        drag: false,
      }
      items.push(it)
      if (!reduce) cleanups.push(bindDrag(it, () => ({ W, H }), play))
    })

    if (reduce) {
      items.forEach((it, i) => {
        it.y = H - it.h - 6
        it.ang = (i - 2) * 4
        it.el.style.transform = `translate(${it.x}px,${it.y}px) rotate(${it.ang}deg)`
      })
      return () => {
        host.innerHTML = ''
      }
    }

    // ambient sprinkles
    const cols = ['--mango', '--cherry', '--pistachio', '--sky', '--yuzu', '--poppy']
    const sprinkles: Sprinkle[] = []
    for (let i = 0; i < 14; i++) {
      const s = document.createElement('div')
      s.style.cssText = `position:absolute;top:0;left:0;width:6px;height:16px;border-radius:6px;background:${cssVar(cols[i % cols.length])};opacity:.8;pointer-events:none;will-change:transform`
      host.appendChild(s)
      sprinkles.push({
        el: s,
        x: Math.random() * W,
        y: -20 - Math.random() * H,
        vx: 0,
        vy: 0.6 + Math.random() * 0.9,
        rot: Math.random() * 180,
        vr: (Math.random() - 0.5) * 6,
        hits: 0,
      })
    }

    const grav = 0.16
    let raf = 0
    const loop = () => {
      raf = requestAnimationFrame(loop)
      items.forEach((it) => {
        if (!it.drag) {
          it.vy += grav
          it.x += it.vx
          it.y += it.vy
          it.ang += it.va
          it.va *= 0.985
          if (it.y + it.h > H) {
            it.y = H - it.h
            if (Math.abs(it.vy) > 1.2) it.vy *= -0.5
            else it.vy = 0
            it.vx *= 0.9
          }
          if (it.x < 0) {
            it.x = 0
            it.vx *= -0.6
          }
          if (it.x + it.w > W) {
            it.x = W - it.w
            it.vx *= -0.6
          }
          it.vx *= 0.995
        }
        it.el.style.transform = `translate(${it.x}px,${it.y}px) rotate(${it.ang}deg)`
      })
      sprinkles.forEach((s) => {
        s.vy = Math.min(3.6, s.vy + 0.03)
        s.x += s.vx
        s.vx *= 0.98
        s.y += s.vy
        s.rot += s.vr
        const sh = 16
        if (s.hits < 3) {
          items.forEach((it) => {
            if (it.drag) return
            if (s.x + 6 > it.x + it.w * 0.18 && s.x < it.x + it.w * 0.82) {
              const topY = it.y + it.h * 0.08
              if (s.vy > 0 && s.y + sh >= topY && s.y + sh <= topY + s.vy + 10) {
                s.y = topY - sh
                s.vy *= -0.42
                const cx = it.x + it.w / 2
                s.vx += (s.x < cx ? -1 : 1) * (0.5 + Math.random() * 0.7)
                s.hits += 1
              }
            }
          })
        }
        if (s.y > H + 24) {
          s.y = -20
          s.x = Math.random() * W
          s.vx = 0
          s.vy = 0.6 + Math.random() * 0.9
          s.hits = 0
        }
        if (s.x < -20) s.x = W + 10
        else if (s.x > W + 20) s.x = -10
        s.el.style.transform = `translate(${s.x}px,${s.y}px) rotate(${s.rot}deg)`
      })
    }
    raf = requestAnimationFrame(loop)

    const onResize = () => {
      const r = host.getBoundingClientRect()
      W = r.width
      H = r.height
    }
    window.addEventListener('resize', onResize)

    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener('resize', onResize)
      cleanups.forEach((c) => c())
      host.innerHTML = ''
    }
  }, [theme, play])

  return (
    <div
      style={{
        position: 'relative',
        overflow: 'hidden',
        background: 'linear-gradient(180deg,var(--sky-1) 0%,var(--sky-2) 62%,var(--bg) 100%)',
      }}
    >
      <Seo title={t.home.name} description={t.home.sub} path="/" />
      <div
        ref={glowRef}
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: 150,
          height: 150,
          margin: '-75px 0 0 -75px',
          borderRadius: '50%',
          background: 'radial-gradient(circle,rgba(255,210,63,.35),transparent 70%)',
          pointerEvents: 'none',
          zIndex: 60,
          mixBlendMode: 'screen',
          opacity: 0,
          transition: 'opacity .3s',
        }}
      />
      {/* sun */}
      <div style={{ position: 'absolute', top: '4%', right: '7%', zIndex: 1 }} className="anim-spin">
        <div
          className="anim-tw"
          style={{
            width: 120,
            height: 120,
            borderRadius: '50%',
            background: 'radial-gradient(circle at 38% 38%,#FFF0B0,var(--sun))',
            boxShadow: '0 0 60px 12px var(--sun),0 0 0 10px rgba(255,210,63,.18)',
          }}
        />
      </div>

      <section
        style={{
          position: 'relative',
          zIndex: 20,
          display: 'grid',
          gridTemplateColumns: 'minmax(0,1.05fr) minmax(0,.95fr)',
          gap: 24,
          alignItems: 'center',
          padding: 'clamp(16px,3vw,48px) clamp(18px,5vw,64px) 40px',
          maxWidth: 1440,
          margin: '0 auto',
        }}
        className="hero-grid"
      >
        <div style={{ maxWidth: 560 }}>
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 9,
              padding: '8px 15px',
              borderRadius: 999,
              background: 'var(--surface)',
              border: '2px solid var(--cream-line)',
              fontWeight: 600,
              fontSize: 14,
              color: 'var(--ink-soft)',
              boxShadow: 'var(--shadow)',
            }}
          >
            <span style={{ width: 9, height: 9, borderRadius: '50%', background: 'var(--pistachio)' }} />
            {t.home.kicker}
          </div>
          <h1
            className="display"
            style={{
              fontWeight: 800,
              lineHeight: 1.04,
              letterSpacing: '-.02em',
              fontSize: 'clamp(40px,6.4vw,84px)',
              margin: '18px 0 0',
              textWrap: 'balance',
            }}
          >
            {t.home.name}
          </h1>
          <p
            className="display"
            style={{
              fontWeight: 600,
              fontSize: 'clamp(20px,2.4vw,30px)',
              lineHeight: 1.2,
              margin: '6px 0 0',
              color: 'var(--cherry-strong)',
            }}
          >
            {t.home.tag}
          </p>
          <p style={{ fontSize: 'clamp(16px,1.5vw,19px)', lineHeight: 1.55, margin: '20px 0 0', color: 'var(--ink-soft)', maxWidth: '48ch' }}>
            {t.home.sub}
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 14, marginTop: 30 }}>
            <Link
              to="/projects"
              className="btn-primary display"
              style={{ display: 'inline-flex', alignItems: 'center', gap: 10, padding: '15px 26px', borderRadius: 16, fontSize: 18 }}
            >
              {t.home.ctaPrimary} <span style={{ fontSize: 20 }}>→</span>
            </Link>
            <Link
              to="/arcade"
              className="btn-secondary display"
              style={{ display: 'inline-flex', alignItems: 'center', gap: 10, padding: '15px 26px', borderRadius: 16, fontSize: 18, background: 'var(--surface)', boxShadow: 'var(--shadow)' }}
            >
              {t.home.ctaSecondary}
            </Link>
          </div>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 9, marginTop: 22, fontSize: 14, color: 'var(--ink-soft)', fontWeight: 500 }}>
            <span
              style={{ display: 'grid', placeItems: 'center', width: 26, height: 26, borderRadius: '50%', background: 'var(--yuzu)', color: 'var(--ink)', fontWeight: 800 }}
            >
              ✱
            </span>
            {t.home.flingHint}
          </div>
        </div>

        <div style={{ position: 'relative', minHeight: 'clamp(320px,44vw,520px)' }}>
          <div ref={playRef} style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }} />
        </div>
      </section>

      <section style={{ position: 'relative', zIndex: 20, maxWidth: 1440, margin: '0 auto', padding: '8px clamp(18px,5vw,64px) 70px' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(200px,1fr))', gap: 16 }}>
          {t.home.paths.map((p) => (
            <Link
              key={p.href}
              to={p.href}
              className="card anim-bob"
              style={{ display: 'flex', flexDirection: 'column', gap: 10, padding: 20, borderRadius: 22, color: 'var(--ink)' }}
            >
              <span style={{ width: 46, height: 46, borderRadius: 14, display: 'grid', placeItems: 'center', fontSize: 24, background: p.tint }}>
                {p.emoji}
              </span>
              <span className="display" style={{ fontWeight: 700, fontSize: 20 }}>
                {p.title}
              </span>
              <span style={{ fontSize: 14, color: 'var(--ink-soft)', lineHeight: 1.4 }}>{p.desc}</span>
            </Link>
          ))}
        </div>
      </section>
    </div>
  )
}

function bindDrag(it: Item, bounds: () => { W: number; H: number }, play: (f: number, d?: number, t?: OscillatorType) => void): () => void {
  const down = (e: PointerEvent) => {
    e.preventDefault()
    it.drag = true
    it.vx = 0
    it.vy = 0
    it.el.style.cursor = 'grabbing'
    it.el.style.zIndex = '10'
    it.px = e.clientX
    it.py = e.clientY
    it.lx = e.clientX
    it.ly = e.clientY
    play(520, 0.08, 'triangle')
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
  }
  const move = (e: PointerEvent) => {
    if (!it.drag) return
    it.x += e.clientX - (it.lx ?? e.clientX)
    it.y += e.clientY - (it.ly ?? e.clientY)
    it.va = (e.clientX - (it.lx ?? e.clientX)) * 0.3
    it.lx = e.clientX
    it.ly = e.clientY
  }
  const up = (e: PointerEvent) => {
    it.drag = false
    it.el.style.cursor = 'grab'
    it.el.style.zIndex = '1'
    const { W, H } = bounds()
    void W
    void H
    it.vx = Math.max(-22, Math.min(22, ((it.lx ?? 0) - (it.px ?? 0)) * 0.25 + (Math.random() - 0.5)))
    it.vy = Math.max(-24, Math.min(6, ((it.ly ?? 0) - (it.py ?? 0)) * 0.18))
    void e
    play(300, 0.12, 'sine')
    window.removeEventListener('pointermove', move)
    window.removeEventListener('pointerup', up)
  }
  it.el.addEventListener('pointerdown', down)
  return () => {
    it.el.removeEventListener('pointerdown', down)
    window.removeEventListener('pointermove', move)
    window.removeEventListener('pointerup', up)
  }
}
