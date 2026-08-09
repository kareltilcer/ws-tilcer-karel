import { scoopSprite, catcherConeSprite, cssVar } from '../sprites'

export interface GameHandle {
  stop: () => void
}

interface Callbacks {
  onScore: (score: number) => void
  onLives: (lives: number) => void
  onOver: (finalScore: number) => void
  playCatch: () => void
  playMiss: () => void
}

interface Scoop {
  el: HTMLDivElement
  x: number
  y: number
  v: number
  rot: number
  vr: number
}

// Catch-the-scoop: the cone follows the pointer / arrow keys; falling scoops are
// caught for +10, a miss costs a life, three misses end the run. Difficulty ramps
// with elapsed time. Bounded score aligns with the backend registry cap.
export function startCatchTheScoop(field: HTMLElement, cb: Callbacks): GameHandle {
  field.innerHTML = ''
  field.style.cursor = 'none'
  const rect = field.getBoundingClientRect()
  let W = rect.width
  let H = rect.height

  const cone = document.createElement('div')
  cone.innerHTML = catcherConeSprite()
  const coneW = 82
  cone.style.cssText = `position:absolute;width:${coneW}px;height:88px;bottom:8px;left:0;transform:translateX(${W / 2 - coneW / 2}px);filter:drop-shadow(0 6px 8px rgba(58,36,22,.3));will-change:transform;z-index:3`
  field.appendChild(cone)

  let coneX = W / 2
  let coneVel = 0
  let score = 0
  let lives = 3
  const scoops: Scoop[] = []
  let spawnAcc = 0
  let last = performance.now()
  let elapsed = 0
  const cols = [cssVar('--mango'), cssVar('--cherry'), cssVar('--pistachio'), cssVar('--poppy'), cssVar('--yuzu')]

  const onMove = (e: PointerEvent) => {
    const r = field.getBoundingClientRect()
    coneX = Math.max(coneW / 2, Math.min(W - coneW / 2, e.clientX - r.left))
  }
  const onTouch = (e: TouchEvent) => {
    const r = field.getBoundingClientRect()
    coneX = Math.max(coneW / 2, Math.min(W - coneW / 2, e.touches[0].clientX - r.left))
  }
  const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'ArrowLeft') coneVel = -1
    else if (e.key === 'ArrowRight') coneVel = 1
  }
  const onKeyUp = (e: KeyboardEvent) => {
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') coneVel = 0
  }
  field.addEventListener('pointermove', onMove)
  field.addEventListener('touchmove', onTouch, { passive: true })
  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('keyup', onKeyUp)
  const onResize = () => {
    const r = field.getBoundingClientRect()
    W = r.width
    H = r.height
  }
  window.addEventListener('resize', onResize)

  let raf = 0
  let stopped = false

  const spawn = (diff: number) => {
    const s: Scoop = {
      el: document.createElement('div'),
      x: 30 + Math.random() * (W - 60),
      y: -30,
      v: 2.0 + Math.random() * 1.4 + diff * 0.7,
      rot: (Math.random() - 0.5) * 30,
      vr: (Math.random() - 0.5) * 4,
    }
    s.el.innerHTML = scoopSprite(cols[Math.floor(Math.random() * cols.length)])
    s.el.style.cssText = 'position:absolute;width:44px;height:52px;left:0;top:0;filter:drop-shadow(0 4px 5px rgba(58,36,22,.25));will-change:transform;z-index:2'
    field.appendChild(s.el)
    scoops.push(s)
  }

  const tick = (now: number) => {
    if (stopped) return
    raf = requestAnimationFrame(tick)
    const dt = Math.min(40, now - last)
    last = now
    elapsed += dt
    coneX = Math.max(coneW / 2, Math.min(W - coneW / 2, coneX + coneVel * 0.6 * dt))
    cone.style.transform = `translateX(${coneX - coneW / 2}px)`
    const diff = 1 + elapsed / 45000
    const spawnEvery = Math.max(560, 1050 / diff)
    spawnAcc += dt
    if (spawnAcc > spawnEvery) {
      spawnAcc = 0
      spawn(diff)
    }
    const coneTop = H - 92
    const catchHalf = coneW / 2 + 6
    for (let i = scoops.length - 1; i >= 0; i--) {
      const s = scoops[i]
      s.y += (s.v * dt) / 16
      s.rot += s.vr
      s.el.style.transform = `translate(${s.x - 22}px,${s.y - 22}px) rotate(${s.rot}deg)`
      if (s.y >= coneTop && Math.abs(s.x - coneX) < catchHalf) {
        score += 10
        cb.playCatch()
        s.el.remove()
        scoops.splice(i, 1)
        cb.onScore(score)
      } else if (s.y > H + 30) {
        s.el.remove()
        scoops.splice(i, 1)
        lives--
        cb.playMiss()
        cb.onLives(lives)
        if (lives <= 0) {
          stop()
          cb.onOver(score)
          return
        }
      }
    }
  }
  raf = requestAnimationFrame(tick)

  function stop() {
    stopped = true
    if (raf) cancelAnimationFrame(raf)
    field.removeEventListener('pointermove', onMove)
    field.removeEventListener('touchmove', onTouch)
    window.removeEventListener('keydown', onKeyDown)
    window.removeEventListener('keyup', onKeyUp)
    window.removeEventListener('resize', onResize)
  }

  return { stop }
}
