import type { Lang } from './plural'

// Centralised UX copy in both languages (real Czech + English, from the design).
// Content strings (projects, about, etc.) come bilingual from the API; this is
// the site chrome and playful microcopy.
export interface Strings {
  nav: { projects: string; about: string; skills: string; arcade: string; links: string; contact: string }
  home: {
    kicker: string
    name: string
    tag: string
    sub: string
    ctaPrimary: string
    ctaSecondary: string
    flingHint: string
    paths: { title: string; desc: string; emoji: string; tint: string; href: string }[]
  }
  projects: {
    title: string
    sub: string
    all: string
    featured: string
    empty: string
    error: string
    back: string
    gallery: string
    links: string
    notFound: string
  }
  linktree: { title: string; sub: string; empty: string }
  about: { title: string }
  skills: { title: string; sub: string; empty: string }
  arcade: {
    eyebrow: string
    title: string
    sub: string
    play: string
    soon: string
    best: string
    back: string
    playTab: string
    boardTab: string
    score: string
    start: string
    again: string
    name: string
    namePh: string
    submit: string
    submitting: string
    ts: string
    tsChecking: string
    tsOk: string
    placed: string
    placedSub: string
    boardTitle: string
    boardSub: string
    boardEmptyT: string
    boardEmptyB: string
    playCta: string
    lives: string
    catchOver: string
    matchOver: string
    finalCatch: string
    finalMatch: string
    beat: string
  }
  contact: {
    title: string
    sub: string
    name: string
    email: string
    subject: string
    message: string
    send: string
    sending: string
    successT: string
    successB: string
    errValidation: string
    errSpam: string
    errRate: string
    errServer: string
    ts: string
    tsOk: string
  }
  legal: { title: string; sub: string; ico: string; dic: string; seat: string; register: string; email: string; phone: string }
  notFound: { code: string; title: string; body: string; home: string; projects: string; arcade: string; egg: string }
  footer: { legal: string; madeWith: string }
  common: { loading: string; retry: string; language: string; theme: string; sound: string }
}

const en: Strings = {
  nav: { projects: 'Projects', about: 'About', skills: 'Skills', arcade: 'Arcade', links: 'Links', contact: 'Contact' },
  home: {
    kicker: 'Developer · Game-maker · Ice-cream enthusiast',
    name: 'Ahoj, I’m Karel.',
    tag: 'I build playful apps, games & the odd perfect scoop.',
    sub: 'Former Czechitas teacher, full-time developer, aspiring game-maker. Right now I’m building YarnLog and a home-management app — and chasing the perfect mango scoop. Welcome to my sunny corner of the internet.',
    ctaPrimary: 'See my work',
    ctaSecondary: 'Enter the arcade',
    flingHint: 'Grab a scoop — give it a fling.',
    paths: [
      { title: 'Projects', desc: 'Software & hand-made, side by side.', emoji: '🧶', tint: 'var(--pistachio)', href: '/projects' },
      { title: 'Arcade', desc: 'Summer minigames + high scores.', emoji: '🕹️', tint: 'var(--sky)', href: '/arcade' },
      { title: 'Links', desc: 'Everywhere else you’ll find me.', emoji: '🍦', tint: 'var(--yuzu)', href: '/links' },
      { title: 'Contact', desc: 'Say ahoj — I reply.', emoji: '✉️', tint: 'var(--poppy)', href: '/contact' },
    ],
  },
  projects: {
    title: 'Projects',
    sub: 'Software and hand-made work, side by side.',
    all: 'All',
    featured: 'Featured',
    empty: 'No projects yet — check back soon.',
    error: 'Could not load projects.',
    back: 'Back to projects',
    gallery: 'Gallery',
    links: 'Links',
    notFound: 'That project melted away.',
  },
  linktree: { title: 'Find me elsewhere', sub: 'The fast path to everywhere.', empty: 'No links yet.' },
  about: { title: 'About me' },
  skills: { title: 'Skills', sub: 'What I reach for.', empty: 'No skills listed yet.' },
  arcade: {
    eyebrow: 'The Arcade',
    title: 'Play something sweet.',
    sub: 'Tiny summer games, real leaderboards, no account needed. Grab a cone and go.',
    play: 'Play',
    soon: 'Soon',
    best: 'Best',
    back: 'Arcade',
    playTab: 'Play',
    boardTab: 'Leaderboard',
    score: 'Score',
    start: 'Play',
    again: 'Play again',
    name: 'Your name',
    namePh: 'e.g. Mango Master',
    submit: 'Submit score',
    submitting: 'Sending…',
    ts: 'I’m not a melting robot',
    tsChecking: 'Checking…',
    tsOk: 'Verified — you’re human',
    placed: 'You placed #',
    placedSub: 'Nice scooping. Here’s the fresh board:',
    boardTitle: 'Leaderboard',
    boardSub: 'Top scores — top 10, all time.',
    boardEmptyT: 'No scores yet',
    boardEmptyB: 'Be the first to make the board.',
    playCta: 'Play now',
    lives: 'Lives',
    catchOver: 'Splat!',
    matchOver: 'Board cleared!',
    finalCatch: 'You caught',
    finalMatch: 'You scored',
    beat: 'Add your name to the board below.',
  },
  contact: {
    title: 'Say ahoj',
    sub: 'Got a project, a question, or just want to chat? Drop me a line.',
    name: 'Your name',
    email: 'Your email',
    subject: 'Subject (optional)',
    message: 'Message',
    send: 'Send message',
    sending: 'Sending…',
    successT: 'Message sent! 🍦',
    successB: 'Thanks — I’ll get back to you soon.',
    errValidation: 'Please check the fields and try again.',
    errSpam: 'Spam check failed — please retry the challenge.',
    errRate: 'Too many messages — please wait a moment.',
    errServer: 'Something went wrong. Please try again.',
    ts: 'I’m not a melting robot',
    tsOk: 'Verified — you’re human',
  },
  legal: {
    title: 'Legal notice',
    sub: 'Mandatory disclosure of a Czech sole trader (OSVČ).',
    ico: 'Business ID (IČO)',
    dic: 'VAT ID (DIČ)',
    seat: 'Place of business',
    register: 'Trade register',
    email: 'Email',
    phone: 'Phone',
  },
  notFound: {
    code: '404',
    title: 'This scoop rolled away.',
    body: 'The page you’re after isn’t here — but there’s plenty more to taste.',
    home: 'Back home',
    projects: 'See projects',
    arcade: 'Enter the arcade',
    egg: 'Psst — the dropped scoop is still delicious.',
  },
  footer: { legal: 'Legal notice', madeWith: 'Made with sunshine & sprinkles.' },
  common: { loading: 'Loading…', retry: 'Retry', language: 'Language', theme: 'Theme', sound: 'Sound' },
}

const cs: Strings = {
  nav: { projects: 'Projekty', about: 'O mně', skills: 'Dovednosti', arcade: 'Arkáda', links: 'Odkazy', contact: 'Kontakt' },
  home: {
    kicker: 'Vývojář · Tvůrce her · Nadšenec do zmrzliny',
    name: 'Ahoj, já jsem Karel.',
    tag: 'Tvořím hravé appky, hry a občas dokonalý kopeček.',
    sub: 'Bývalý lektor v Czechitas, developer na plný úvazek, začínající tvůrce her. Zrovna teď stavím YarnLog a appku na správu domácnosti — a lovím dokonalý mangový kopeček. Vítej v mém slunečním koutku internetu.',
    ctaPrimary: 'Prohlédni si moji práci',
    ctaSecondary: 'Vstup do arkády',
    flingHint: 'Chyť kopeček — a mrskni s ním.',
    paths: [
      { title: 'Projekty', desc: 'Software i ruční práce vedle sebe.', emoji: '🧶', tint: 'var(--pistachio)', href: '/projects' },
      { title: 'Arkáda', desc: 'Letní hry + žebříčky.', emoji: '🕹️', tint: 'var(--sky)', href: '/arcade' },
      { title: 'Odkazy', desc: 'Všude jinde, kde mě najdeš.', emoji: '🍦', tint: 'var(--yuzu)', href: '/links' },
      { title: 'Kontakt', desc: 'Řekni ahoj — odpovím.', emoji: '✉️', tint: 'var(--poppy)', href: '/contact' },
    ],
  },
  projects: {
    title: 'Projekty',
    sub: 'Software i ruční práce vedle sebe.',
    all: 'Vše',
    featured: 'Vybrané',
    empty: 'Zatím žádné projekty — zkus to brzy znovu.',
    error: 'Projekty se nepodařilo načíst.',
    back: 'Zpět na projekty',
    gallery: 'Galerie',
    links: 'Odkazy',
    notFound: 'Tenhle projekt se rozpustil.',
  },
  linktree: { title: 'Najdi mě jinde', sub: 'Nejrychlejší cesta kamkoliv.', empty: 'Zatím žádné odkazy.' },
  about: { title: 'O mně' },
  skills: { title: 'Dovednosti', sub: 'Po čem sahám.', empty: 'Zatím žádné dovednosti.' },
  arcade: {
    eyebrow: 'Arkáda',
    title: 'Zahraj si něco sladkého.',
    sub: 'Malé letní hry, opravdové žebříčky, žádný účet. Chyť kornout a jedu.',
    play: 'Hrát',
    soon: 'Brzy',
    best: 'Nej',
    back: 'Arkáda',
    playTab: 'Hra',
    boardTab: 'Žebříček',
    score: 'Skóre',
    start: 'Hrát',
    again: 'Hrát znovu',
    name: 'Tvé jméno',
    namePh: 'např. Mistr Mango',
    submit: 'Odeslat skóre',
    submitting: 'Odesílám…',
    ts: 'Nejsem tající robot',
    tsChecking: 'Ověřuji…',
    tsOk: 'Ověřeno — jsi člověk',
    placed: 'Jsi na místě #',
    placedSub: 'Pěkné nabírání. Tady je čerstvý žebříček:',
    boardTitle: 'Žebříček',
    boardSub: 'Nejlepší skóre — top 10, za celou dobu.',
    boardEmptyT: 'Zatím žádné skóre',
    boardEmptyB: 'Buď první na desce.',
    playCta: 'Hrát teď',
    lives: 'Životy',
    catchOver: 'Žuch!',
    matchOver: 'Deska čistá!',
    finalCatch: 'Chytil jsi',
    finalMatch: 'Nasbíral jsi',
    beat: 'Přidej své jméno do žebříčku níže.',
  },
  contact: {
    title: 'Řekni ahoj',
    sub: 'Máš projekt, otázku, nebo si chceš jen popovídat? Napiš mi.',
    name: 'Tvé jméno',
    email: 'Tvůj e-mail',
    subject: 'Předmět (nepovinné)',
    message: 'Zpráva',
    send: 'Odeslat zprávu',
    sending: 'Odesílám…',
    successT: 'Zpráva odeslána! 🍦',
    successB: 'Díky — brzy se ozvu.',
    errValidation: 'Zkontroluj prosím pole a zkus to znovu.',
    errSpam: 'Kontrola proti spamu selhala — zkus výzvu znovu.',
    errRate: 'Příliš mnoho zpráv — chvíli počkej.',
    errServer: 'Něco se pokazilo. Zkus to prosím znovu.',
    ts: 'Nejsem tající robot',
    tsOk: 'Ověřeno — jsi člověk',
  },
  legal: {
    title: 'Právní informace',
    sub: 'Povinně zveřejněné informace o podnikateli (OSVČ).',
    ico: 'IČO',
    dic: 'DIČ',
    seat: 'Místo podnikání',
    register: 'Živnostenský rejstřík',
    email: 'E-mail',
    phone: 'Telefon',
  },
  notFound: {
    code: '404',
    title: 'Tenhle kopeček se odkutálel.',
    body: 'Stránka, kterou hledáš, tu není — ale je tu spousta dalšího k ochutnání.',
    home: 'Zpět domů',
    projects: 'Projekty',
    arcade: 'Do arkády',
    egg: 'Pssst — spadlý kopeček je pořád výborný.',
  },
  footer: { legal: 'Právní informace', madeWith: 'Vyrobeno se sluncem a posypem.' },
  common: { loading: 'Načítám…', retry: 'Zkusit znovu', language: 'Jazyk', theme: 'Motiv', sound: 'Zvuk' },
}

export const STRINGS: Record<Lang, Strings> = { cs, en }
