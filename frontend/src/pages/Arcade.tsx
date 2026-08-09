import { useEffect, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useLang } from '../i18n/lang'
import { useSound } from '../lib/sound'
import { Seo } from '../components/Seo'
import { Turnstile } from '../components/Turnstile'
import { CenterState, Spinner } from '../components/states'
import { useLeaderboard } from '../api/hooks'
import { apiFetch, ApiError } from '../api/client'
import { qk } from '../api/keys'
import { formatDate } from '../i18n/format'
import { startCatchTheScoop, type GameHandle } from '../lib/games/catchTheScoop'
import { startScoopMatch } from '../lib/games/scoopMatch'
import type { ScoreSubmit, ScoreSubmitResult } from '../api/types'

type Phase = 'attract' | 'playing' | 'over'
type SubmitState = 'idle' | 'sending' | 'done'

const ROSTER = [
  { key: 'catch-the-scoop', icon: '🍦', bg: '--sky', playable: true },
  { key: 'scoop-match', icon: '🍧', bg: '--poppy', playable: true },
  { key: 'stack-the-cones', icon: '🍨', bg: '--mango', playable: false },
  { key: 'sprinkle-flick', icon: '✨', bg: '--pistachio', playable: false },
] as const

const GAME_META: Record<string, { title: Record<'cs' | 'en', string>; desc: Record<'cs' | 'en', string>; howto: Record<'cs' | 'en', string> }> = {
  'catch-the-scoop': {
    title: { en: 'Catch the Scoop', cs: 'Chyť kopeček' },
    desc: { en: 'Slide your cone, catch falling scoops, don’t let them splat. Three misses and it’s over.', cs: 'Posouvej kornout, chytej padající kopečky a nenech je žuchnout. Tři minutí a končíš.' },
    howto: { en: 'Move your cone with the mouse, finger, or arrow keys. Catch the scoops — miss three and you’re done.', cs: 'Hýbej kornoutem myší, prstem nebo šipkami. Chytej kopečky — tři minutí a končíš.' },
  },
  'scoop-match': {
    title: { en: 'Scoop Match', cs: 'Páry kopečků' },
    desc: { en: 'A flavour-matching memory game. Flip, remember, clear the board in as few moves as you can.', cs: 'Pexeso s příchutěmi. Otáčej, pamatuj si a vyčisti desku na co nejméně tahů.' },
    howto: { en: 'Flip two cards and find the matching flavour. Fast combos score more.', cs: 'Otáčej dvě karty a hledej stejnou příchuť. Rychlá komba dávají víc.' },
  },
  'stack-the-cones': {
    title: { en: 'Cone Stacker', cs: 'Věž z kornoutů' },
    desc: { en: 'Balance a wobbling tower of cones. One shaky scoop too many and it topples.', cs: 'Balancuj kymácející se věž z kornoutů. Jeden vratký kopeček navíc a je to dole.' },
    howto: { en: '', cs: '' },
  },
  'sprinkle-flick': {
    title: { en: 'Sprinkle Flick', cs: 'Cvrnkni posyp' },
    desc: { en: 'Flick sprinkles onto the scoop with real physics. Land the combo, rack up the multiplier.', cs: 'Cvrnkej posyp na kopeček s opravdovou fyzikou. Tref kombo a nasbírej násobič.' },
    howto: { en: '', cs: '' },
  },
}

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div style={{ maxWidth: 1080, margin: '0 auto', padding: 'clamp(24px,3vw,44px) clamp(18px,5vw,64px) 90px' }}>{children}</div>
)

export default function Arcade() {
  const { t, lang } = useLang()
  const { play } = useSound()
  const qc = useQueryClient()
  const [view, setView] = useState<'hub' | 'game'>('hub')
  const [gameKey, setGameKey] = useState('catch-the-scoop')
  const [tab, setTab] = useState<'play' | 'board'>('play')
  const [phase, setPhase] = useState<Phase>('attract')
  const [score, setScore] = useState(0)
  const [lives, setLives] = useState(3)
  const [matched, setMatched] = useState(0)

  const [name, setName] = useState('')
  const [website, setWebsite] = useState('') // honeypot
  const [token, setToken] = useState('')
  const [tsReset, setTsReset] = useState(0)
  const [submit, setSubmit] = useState<SubmitState>('idle')
  const [rank, setRank] = useState<number | null>(null)

  const fieldRef = useRef<HTMLDivElement>(null)
  const handleRef = useRef<GameHandle | null>(null)

  // Run the selected game engine while phase === 'playing'.
  useEffect(() => {
    if (view !== 'game' || phase !== 'playing') return
    const field = fieldRef.current
    if (!field) return
    const commonSound = {
      playCatch: () => {
        play(660, 0.1, 'triangle', 0.16)
        setTimeout(() => play(880, 0.09, 'triangle', 0.13), 60)
      },
      playMiss: () => play(150, 0.22, 'sawtooth', 0.14),
    }
    if (gameKey === 'scoop-match') {
      handleRef.current = startScoopMatch(field, {
        onScore: setScore,
        onMatched: setMatched,
        onWin: () => {
          ;[523, 659, 784, 1047].forEach((f, i) => setTimeout(() => play(f, 0.16, 'triangle', 0.14), i * 90))
          setPhase('over')
        },
        playFlip: (id) => play(500 + id * 45, 0.08, 'triangle', 0.12),
        playMatch: commonSound.playCatch,
        playMiss: commonSound.playMiss,
      })
    } else {
      handleRef.current = startCatchTheScoop(field, {
        onScore: setScore,
        onLives: setLives,
        onOver: () => {
          ;[520, 400, 300, 220].forEach((f, i) => setTimeout(() => play(f, 0.18, 'square', 0.12), i * 130))
          setPhase('over')
        },
        ...commonSound,
      })
    }
    return () => {
      handleRef.current?.stop()
      handleRef.current = null
    }
  }, [view, phase, gameKey, play])

  function enter(key: string, toTab: 'play' | 'board' = 'play') {
    setGameKey(key)
    setTab(toTab)
    setPhase('attract')
    setScore(0)
    setLives(3)
    setMatched(0)
    setSubmit('idle')
    setToken('')
    setName('')
    setView('game')
    window.scrollTo(0, 0)
  }
  function backToHub() {
    handleRef.current?.stop()
    setView('hub')
    setPhase('attract')
    setTab('play')
  }
  function startGame() {
    setScore(0)
    setLives(3)
    setMatched(0)
    setSubmit('idle')
    setToken('')
    setPhase('playing')
  }

  async function submitScore() {
    if (!token || !name.trim() || submit !== 'idle') return
    setSubmit('sending')
    const body: ScoreSubmit = { player_name: name.trim(), score, locale: lang, website, turnstile_token: token }
    try {
      const res = await apiFetch<ScoreSubmitResult>(`/api/arcade/${gameKey}/scores`, { method: 'POST', body })
      setRank(res.rank)
      setSubmit('done')
      void qc.invalidateQueries({ queryKey: qk.leaderboard(gameKey) })
      play(880, 0.16, 'triangle', 0.14)
    } catch (err) {
      setSubmit('idle')
      // Re-arm Turnstile: the token was spent by this attempt and cannot verify again.
      setTsReset((n) => n + 1)
      const msg = err instanceof ApiError && err.status === 429 ? t.contact.errRate : err instanceof ApiError && err.status === 400 ? t.contact.errSpam : t.contact.errServer
      toast.error(msg)
    }
  }

  if (view === 'hub') return <Hub onEnter={enter} />

  const meta = GAME_META[gameKey]
  return (
    <Wrap>
      <Seo title={meta.title[lang]} description={meta.desc[lang]} path="/arcade" />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 14, marginBottom: 18, flexWrap: 'wrap' }}>
        <button onClick={backToHub} className="btn-secondary" style={{ height: 40, padding: '0 16px', fontSize: 15 }}>
          ← {t.arcade.back}
        </button>
        <h2 className="display" style={{ fontWeight: 800, fontSize: 'clamp(22px,3vw,32px)', margin: 0 }}>
          {meta.title[lang]}
        </h2>
        <div style={{ display: 'flex', gap: 8 }}>
          <Tab active={tab === 'play'} onClick={() => setTab('play')}>
            {t.arcade.playTab}
          </Tab>
          <Tab active={tab === 'board'} onClick={() => setTab('board')}>
            🏆 {t.arcade.boardTab}
          </Tab>
        </div>
      </div>

      {tab === 'play' ? (
        <div style={{ position: 'relative' }}>
          {/* HUD */}
          <div style={{ position: 'absolute', top: 14, left: 14, right: 14, zIndex: 5, display: 'flex', alignItems: 'center', justifyContent: 'space-between', pointerEvents: 'none' }}>
            <div className="mono card" style={{ padding: '8px 14px', fontWeight: 700, fontSize: 16 }}>
              {t.arcade.score} {score}
            </div>
            <div style={{ display: 'flex', gap: 5 }}>
              {gameKey === 'scoop-match' ? (
                <div className="mono card" style={{ padding: '8px 14px', fontWeight: 700, fontSize: 15 }}>🍧 {matched}/6</div>
              ) : (
                [0, 1, 2].map((i) => (
                  <span key={i} style={{ fontSize: 22, filter: i < lives ? 'none' : 'grayscale(1) opacity(.35)' }}>
                    🍦
                  </span>
                ))
              )}
            </div>
          </div>

          <div
            ref={fieldRef}
            style={{
              position: 'relative',
              width: '100%',
              height: 'clamp(380px,54vh,520px)',
              borderRadius: 24,
              overflow: 'hidden',
              border: '2px solid var(--cream-line)',
              background: 'linear-gradient(180deg,var(--sky-1),var(--sky-2))',
              boxShadow: 'var(--shadow)',
              touchAction: 'none',
            }}
          />

          {phase === 'attract' && (
            <Overlay>
              <div className="anim-bob2" style={{ fontSize: 66 }}>
                🍦
              </div>
              <h3 className="display" style={{ fontWeight: 800, fontSize: 'clamp(26px,4vw,40px)', margin: 0, color: '#fff', textShadow: '0 3px 0 rgba(0,0,0,.15)' }}>
                {meta.title[lang]}
              </h3>
              <p style={{ color: '#fff', margin: 0, maxWidth: '36ch', fontSize: 16, lineHeight: 1.5, opacity: 0.95 }}>{meta.howto[lang]}</p>
              <button onClick={startGame} className="display" style={{ marginTop: 6, padding: '16px 34px', borderRadius: 16, border: 'none', background: 'var(--yuzu)', color: '#3A2416', fontWeight: 800, fontSize: 20, cursor: 'pointer', boxShadow: '0 12px 0 rgba(0,0,0,.18)' }}>
                {t.arcade.start} ▶
              </button>
            </Overlay>
          )}

          {phase === 'over' && (
            <Overlay dark>
              <div className="card" style={{ width: 'min(440px,94%)', padding: 26 }}>
                {submit !== 'done' ? (
                  <>
                    <div style={{ textAlign: 'center' }}>
                      <div style={{ fontSize: 52 }}>{gameKey === 'scoop-match' ? '🏆' : '🫠'}</div>
                      <h3 className="display" style={{ fontWeight: 800, fontSize: 28, margin: '8px 0 2px' }}>
                        {gameKey === 'scoop-match' ? t.arcade.matchOver : t.arcade.catchOver}
                      </h3>
                      <p style={{ color: 'var(--ink-soft)', margin: 0 }}>{gameKey === 'scoop-match' ? t.arcade.finalMatch : t.arcade.finalCatch}</p>
                      <div className="display" style={{ fontWeight: 800, fontSize: 56, color: 'var(--cherry-strong)', lineHeight: 1, margin: '6px 0 4px' }}>
                        {score}
                      </div>
                      <p style={{ color: 'var(--ink-soft)', margin: '0 0 16px', fontSize: 14 }}>{t.arcade.beat}</p>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 11 }}>
                      <input className="input" maxLength={32} placeholder={t.arcade.namePh} value={name} onChange={(e) => setName(e.target.value)} />
                      <input value={website} onChange={(e) => setWebsite(e.target.value)} tabIndex={-1} autoComplete="off" aria-hidden style={{ position: 'absolute', left: -9999, width: 1, height: 1, opacity: 0 }} />
                      <Turnstile onToken={setToken} resetSignal={tsReset} />
                      <button
                        onClick={submitScore}
                        disabled={!token || !name.trim() || submit === 'sending'}
                        className="btn-primary"
                        style={{ height: 50, fontSize: 17, opacity: !token || !name.trim() ? 0.6 : 1, cursor: !token || !name.trim() ? 'not-allowed' : 'pointer' }}
                      >
                        {submit === 'sending' ? t.arcade.submitting : t.arcade.submit}
                      </button>
                      <button onClick={startGame} className="btn-secondary" style={{ height: 44, fontSize: 15 }}>
                        {t.arcade.again}
                      </button>
                    </div>
                  </>
                ) : (
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 52 }}>🏆</div>
                    <h3 className="display" style={{ fontWeight: 800, fontSize: 26, margin: '8px 0 2px' }}>
                      {t.arcade.placed}
                      {rank}
                    </h3>
                    <p style={{ color: 'var(--ink-soft)', margin: '0 0 14px' }}>{t.arcade.placedSub}</p>
                    <button onClick={startGame} className="btn-primary" style={{ width: '100%', height: 48, fontSize: 16 }}>
                      {t.arcade.again} ▶
                    </button>
                  </div>
                )}
              </div>
            </Overlay>
          )}
        </div>
      ) : (
        <Board gameKey={gameKey} onPlay={() => setTab('play')} />
      )}
    </Wrap>
  )
}

function Hub({ onEnter }: { onEnter: (key: string, tab?: 'play' | 'board') => void }) {
  const { t, lang } = useLang()
  return (
    <Wrap>
      <Seo title={t.arcade.title} description={t.arcade.sub} path="/arcade" />
      <div style={{ textAlign: 'center', maxWidth: 660, margin: '0 auto' }}>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: 9, padding: '7px 15px', borderRadius: 999, background: 'var(--surface)', border: '2px solid var(--cream-line)', fontWeight: 600, fontSize: 13, color: 'var(--ink-soft)' }}>
          <span className="anim-bob" style={{ fontSize: 18 }}>
            🕹️
          </span>
          {t.arcade.eyebrow}
        </div>
        <h1 className="display" style={{ fontWeight: 800, fontSize: 'clamp(40px,7vw,72px)', lineHeight: 1.02, margin: '16px 0 0' }}>
          {t.arcade.title}
        </h1>
        <p style={{ fontSize: 'clamp(16px,1.7vw,20px)', color: 'var(--ink-soft)', margin: '14px 0 0', lineHeight: 1.5 }}>{t.arcade.sub}</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(280px,1fr))', gap: 22, marginTop: 44 }}>
        {ROSTER.map((g) => {
          const m = GAME_META[g.key]
          return (
            <div key={g.key} className="card" style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRadius: 26 }}>
              <div style={{ position: 'relative', height: 150, background: `color-mix(in srgb, var(${g.bg}) 55%, var(--surface))`, display: 'grid', placeItems: 'center' }}>
                <div className="anim-bob2" style={{ fontSize: 60, filter: 'drop-shadow(0 8px 10px rgba(58,36,22,.25))' }}>
                  {g.icon}
                </div>
                {!g.playable && (
                  <div style={{ position: 'absolute', top: 12, right: 12, padding: '5px 12px', borderRadius: 999, background: 'var(--surface)', color: 'var(--ink-soft)', fontWeight: 800, fontSize: 11, textTransform: 'uppercase' }}>
                    {t.arcade.soon}
                  </div>
                )}
              </div>
              <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 8, flex: 1 }}>
                <h3 className="display" style={{ fontWeight: 700, fontSize: 23, margin: 0 }}>
                  {m.title[lang]}
                </h3>
                <p style={{ color: 'var(--ink-soft)', fontSize: 15, lineHeight: 1.5, margin: 0, flex: 1 }}>{m.desc[lang]}</p>
                <div style={{ display: 'flex', gap: 10, marginTop: 6 }}>
                  <button
                    onClick={() => g.playable && onEnter(g.key)}
                    disabled={!g.playable}
                    className="display"
                    style={{ flex: 1, padding: 13, borderRadius: 14, border: 'none', background: g.playable ? 'var(--cherry-strong)' : 'var(--surface-2)', color: g.playable ? '#fff' : 'var(--ink-soft)', fontWeight: 700, fontSize: 16, cursor: g.playable ? 'pointer' : 'not-allowed', opacity: g.playable ? 1 : 0.7 }}
                  >
                    {g.playable ? t.arcade.play : t.arcade.soon}
                  </button>
                  {g.playable && (
                    <button onClick={() => onEnter(g.key, 'board')} className="btn-secondary" style={{ padding: '13px 16px', fontSize: 15 }}>
                      🏆
                    </button>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </Wrap>
  )
}

function Board({ gameKey, onPlay }: { gameKey: string; onPlay: () => void }) {
  const { t, lang } = useLang()
  const { data, isLoading } = useLeaderboard(gameKey)
  const medal = (i: number) => (i === 0 ? ['var(--yuzu)', '#3A2416'] : i === 1 ? ['#D9D2C4', '#3A2416'] : i === 2 ? ['#E7B26B', '#3A2416'] : ['var(--surface-2)', 'var(--ink-soft)'])
  return (
    <div style={{ maxWidth: 640, margin: '0 auto' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
        <div className="anim-bob" style={{ fontSize: 34 }}>
          🏆
        </div>
        <div>
          <h3 className="display" style={{ fontWeight: 800, fontSize: 26, margin: 0 }}>
            {t.arcade.boardTitle}
          </h3>
          <p style={{ color: 'var(--ink-soft)', margin: '2px 0 0', fontSize: 14 }}>{t.arcade.boardSub}</p>
        </div>
      </div>
      {isLoading ? (
        <div style={{ display: 'grid', placeItems: 'center', padding: 40 }}>
          <Spinner />
        </div>
      ) : !data || data.length === 0 ? (
        <CenterState emoji="🍨" title={t.arcade.boardEmptyT} body={t.arcade.boardEmptyB} />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {data.map((r, i) => {
            const [mb, mc] = medal(i)
            return (
              <div key={i} className="card" style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '14px 16px', borderRadius: 16 }}>
                <span className="display" style={{ width: 34, height: 34, borderRadius: '50%', display: 'grid', placeItems: 'center', background: mb, color: mc, fontWeight: 800, fontSize: 15 }}>
                  {i + 1}
                </span>
                <span style={{ fontWeight: 700, fontSize: 16, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{r.player_name}</span>
                <span className="mono" style={{ fontSize: 12, color: 'var(--ink-soft)' }}>{formatDate(lang, r.created_at)}</span>
                <span className="mono" style={{ fontWeight: 700, fontSize: 18, color: 'var(--cherry-strong)', minWidth: 56, textAlign: 'right' }}>{r.score}</span>
              </div>
            )
          })}
        </div>
      )}
      <button onClick={onPlay} className="btn-primary" style={{ marginTop: 18, width: '100%', height: 50, fontSize: 17 }}>
        {t.arcade.playCta} ▶
      </button>
    </div>
  )
}

function Tab({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      style={{ height: 38, padding: '0 15px', borderRadius: 11, border: `2px solid ${active ? 'var(--cherry-strong)' : 'var(--cream-line)'}`, background: active ? 'var(--cherry-strong)' : 'var(--surface)', color: active ? '#fff' : 'var(--ink)', fontWeight: 700, fontSize: 14, cursor: 'pointer' }}
    >
      {children}
    </button>
  )
}

function Overlay({ children, dark }: { children: React.ReactNode; dark?: boolean }) {
  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        borderRadius: 24,
        background: dark ? 'rgba(27,21,48,.52)' : 'rgba(27,21,48,.34)',
        backdropFilter: 'blur(3px)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        textAlign: 'center',
        gap: 14,
        padding: 24,
        overflow: 'auto',
      }}
    >
      {children}
    </div>
  )
}
