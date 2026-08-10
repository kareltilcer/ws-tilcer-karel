import { scoopSprite, cssVar } from '../sprites'
import type { GameHandle } from './catchTheScoop'

interface Callbacks {
  onScore: (score: number) => void
  onMatched: (matched: number) => void
  onWin: (finalScore: number) => void
  playFlip: (id: number) => void
  playMatch: () => void
  playMiss: () => void
}

interface Card {
  id: number
  c: string
  el: HTMLButtonElement
  done: boolean
}

// Scoop Match: a 6-pair flavour memory game on a 4×3 grid. Matches score 200 +
// a rising combo bonus; clearing all pairs adds a move bonus and ends the game.
export function startScoopMatch(field: HTMLElement, cb: Callbacks): GameHandle {
  field.innerHTML = ''
  field.style.cursor = 'default'
  const cols = [cssVar('--mango'), cssVar('--cherry'), cssVar('--pistachio'), cssVar('--poppy'), cssVar('--yuzu'), cssVar('--sky')]
  const deck: Card[] = []
  cols.forEach((c, i) => {
    deck.push({ id: i, c, el: null as unknown as HTMLButtonElement, done: false })
    deck.push({ id: i, c, el: null as unknown as HTMLButtonElement, done: false })
  })
  for (let i = deck.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[deck[i], deck[j]] = [deck[j], deck[i]]
  }

  let score = 0
  let matched = 0
  let moves = 0
  let combo = 0
  let lock = false
  let flipped: { idx: number; card: Card }[] = []
  let stopped = false
  const timers: number[] = []

  const cover = () =>
    `<span class="display" style="font-weight:800;font-size:clamp(22px,5vw,36px);color:#3A2416;opacity:.45">?</span>`

  const wrap = document.createElement('div')
  wrap.style.cssText =
    'position:absolute;inset:0;display:grid;grid-template-columns:repeat(4,1fr);grid-template-rows:repeat(3,1fr);gap:clamp(6px,1.8%,12px);padding:clamp(10px,3%,22px);min-height:0'

  deck.forEach((card, idx) => {
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.style.cssText =
      'position:relative;border:none;border-radius:16px;width:100%;height:100%;min-width:0;min-height:0;cursor:pointer;background:rgba(255,255,255,.55);box-shadow:0 6px 14px -6px rgba(58,36,22,.4);display:grid;place-items:center;transition:transform .16s,background .2s,box-shadow .2s;padding:8%'
    btn.innerHTML = cover()
    btn.onclick = () => flip(idx, card, btn)
    wrap.appendChild(btn)
    card.el = btn
  })
  field.appendChild(wrap)

  function flip(idx: number, card: Card, btn: HTMLButtonElement) {
    if (stopped || lock || card.done || flipped.find((x) => x.idx === idx)) return
    btn.innerHTML = scoopSprite(card.c)
    btn.style.background = '#fff'
    btn.style.transform = 'scale(1.05)'
    cb.playFlip(card.id)
    flipped.push({ idx, card })
    if (flipped.length === 2) {
      moves++
      lock = true
      const [a, b] = flipped
      if (a.card.id === b.card.id) {
        a.card.done = b.card.done = true
        matched++
        combo++
        score += 200 + (combo - 1) * 60
        cb.playMatch()
        const tt = window.setTimeout(() => {
          ;[a.card.el, b.card.el].forEach((x) => {
            x.style.background = 'color-mix(in srgb,var(--pistachio) 42%,#fff)'
            x.style.boxShadow = '0 0 0 3px var(--pistachio)'
            x.style.transform = 'scale(1)'
          })
        }, 10)
        timers.push(tt)
        flipped = []
        lock = false
        cb.onScore(score)
        cb.onMatched(matched)
        if (matched === 6) {
          const bonus = Math.max(0, 1000 - moves * 55)
          score += bonus
          const wt = window.setTimeout(() => {
            cb.onScore(score)
            stop()
            cb.onWin(score)
          }, 460)
          timers.push(wt)
        }
      } else {
        combo = 0
        cb.playMiss()
        const rt = window.setTimeout(() => {
          ;[a.card.el, b.card.el].forEach((x) => {
            x.innerHTML = cover()
            x.style.background = 'rgba(255,255,255,.55)'
            x.style.transform = 'scale(1)'
          })
          flipped = []
          lock = false
        }, 720)
        timers.push(rt)
      }
    }
  }

  function stop() {
    stopped = true
    timers.forEach((t) => clearTimeout(t))
  }

  return { stop }
}
