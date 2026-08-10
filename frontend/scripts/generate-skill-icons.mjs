// Generates src/lib/skillIcons.data.ts from two monochrome icon sets, tinted at
// render time to the design palette (so a TypeScript logo shows in --sky, not its
// brand blue — keeping the ice-cream vibe). Nothing is hand-drawn: we bake only
// the curated subset below into a small local data file so there is no runtime
// icon dependency and the bundle stays tiny.
//
//   • Simple Icons (simple-icons)  → language / framework / tool brand logos
//   • Lucide       (lucide-static) → generic "other" glyphs + anything logo-less
//
// To add or change an icon: edit CATALOG, then run `npm run gen:icons`.
// `accent` is a palette flavour (see src/index.css tokens): rendered as var(--<accent>).
// `key` is the stable string stored in the DB `icon` column.

import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const FRONT = resolve(__dirname, '..')
const SI = resolve(FRONT, 'node_modules/simple-icons/icons')
const LU = resolve(FRONT, 'node_modules/lucide-static/icons')
const OUT = resolve(FRONT, 'src/lib/skillIcons.data.ts')

// group ∈ lang | framework | tool | other. `s` = source: 'si' (Simple Icons) | 'lu' (Lucide).
// For si, `key` is also the svg filename; for lu, `file` overrides the filename.
const CATALOG = [
  // ---------------- languages ----------------
  ['si', 'typescript', 'TypeScript', 'lang', 'sky'],
  ['si', 'javascript', 'JavaScript', 'lang', 'yuzu'],
  ['si', 'python', 'Python', 'lang', 'pistachio'],
  ['si', 'go', 'Go', 'lang', 'sky'],
  ['si', 'rust', 'Rust', 'lang', 'tanger'],
  ['si', 'c', 'C', 'lang', 'poppy'],
  ['si', 'cplusplus', 'C++', 'lang', 'sky'],
  ['si', 'php', 'PHP', 'lang', 'poppy'],
  ['si', 'ruby', 'Ruby', 'lang', 'cherry'],
  ['si', 'swift', 'Swift', 'lang', 'tanger'],
  ['si', 'kotlin', 'Kotlin', 'lang', 'poppy'],
  ['si', 'dart', 'Dart', 'lang', 'sky'],
  ['si', 'elixir', 'Elixir', 'lang', 'poppy'],
  ['si', 'scala', 'Scala', 'lang', 'cherry'],
  ['si', 'html5', 'HTML5', 'lang', 'tanger'],
  ['si', 'css', 'CSS', 'lang', 'sky'],
  ['si', 'sass', 'Sass', 'lang', 'cherry'],
  ['si', 'lua', 'Lua', 'lang', 'sky'],
  ['si', 'haskell', 'Haskell', 'lang', 'poppy'],
  ['si', 'r', 'R', 'lang', 'sky'],
  ['si', 'gnubash', 'Bash', 'lang', 'pistachio'],
  ['si', 'clojure', 'Clojure', 'lang', 'pistachio'],
  ['si', 'perl', 'Perl', 'lang', 'poppy'],

  // ---------------- frameworks / runtimes ----------------
  ['si', 'react', 'React', 'framework', 'sky'],
  ['si', 'vuedotjs', 'Vue', 'framework', 'pistachio'],
  ['si', 'angular', 'Angular', 'framework', 'cherry'],
  ['si', 'svelte', 'Svelte', 'framework', 'tanger'],
  ['si', 'solid', 'Solid', 'framework', 'sky'],
  ['si', 'nextdotjs', 'Next.js', 'framework', 'poppy'],
  ['si', 'nodedotjs', 'Node.js', 'framework', 'pistachio'],
  ['si', 'deno', 'Deno', 'framework', 'sky'],
  ['si', 'bun', 'Bun', 'framework', 'mango'],
  ['si', 'express', 'Express', 'framework', 'pistachio'],
  ['si', 'fastapi', 'FastAPI', 'framework', 'pistachio'],
  ['si', 'django', 'Django', 'framework', 'pistachio'],
  ['si', 'flask', 'Flask', 'framework', 'poppy'],
  ['si', 'spring', 'Spring', 'framework', 'pistachio'],
  ['si', 'dotnet', '.NET', 'framework', 'poppy'],
  ['si', 'laravel', 'Laravel', 'framework', 'cherry'],
  ['si', 'redux', 'Redux', 'framework', 'poppy'],
  ['si', 'threedotjs', 'Three.js', 'framework', 'poppy'],

  // ---------------- tools / infra / db / design ----------------
  ['si', 'vite', 'Vite', 'tool', 'poppy'],
  ['si', 'webpack', 'Webpack', 'tool', 'sky'],
  ['si', 'tailwindcss', 'Tailwind', 'tool', 'sky'],
  ['si', 'docker', 'Docker', 'tool', 'sky'],
  ['si', 'kubernetes', 'Kubernetes', 'tool', 'sky'],
  ['si', 'git', 'Git', 'tool', 'tanger'],
  ['si', 'github', 'GitHub', 'tool', 'poppy'],
  ['si', 'gitlab', 'GitLab', 'tool', 'tanger'],
  ['si', 'linux', 'Linux', 'tool', 'mango'],
  ['si', 'nginx', 'Nginx', 'tool', 'pistachio'],
  ['si', 'graphql', 'GraphQL', 'tool', 'cherry'],
  ['si', 'figma', 'Figma', 'tool', 'tanger'],
  ['si', 'postgresql', 'PostgreSQL', 'tool', 'sky'],
  ['si', 'mysql', 'MySQL', 'tool', 'sky'],
  ['si', 'sqlite', 'SQLite', 'tool', 'sky'],
  ['si', 'mongodb', 'MongoDB', 'tool', 'pistachio'],
  ['si', 'redis', 'Redis', 'tool', 'cherry'],
  ['si', 'firebase', 'Firebase', 'tool', 'mango'],
  ['si', 'supabase', 'Supabase', 'tool', 'pistachio'],
  ['si', 'vercel', 'Vercel', 'tool', 'poppy'],
  ['si', 'prisma', 'Prisma', 'tool', 'poppy'],
  ['si', 'jest', 'Jest', 'tool', 'cherry'],
  ['si', 'vitest', 'Vitest', 'tool', 'yuzu'],
  ['si', 'storybook', 'Storybook', 'tool', 'cherry'],

  // ---------------- other / generic / craft (Lucide) ----------------
  ['lu', 'database', 'Database', 'other', 'sky'],
  ['lu', 'terminal', 'Terminal', 'other', 'pistachio'],
  ['lu', 'code', 'Code', 'other', 'poppy', 'code-xml'],
  ['lu', 'braces', 'Braces', 'other', 'mango'],
  ['lu', 'server', 'Server', 'other', 'sky'],
  ['lu', 'cpu', 'Chip', 'other', 'poppy'],
  ['lu', 'cloud', 'Cloud', 'other', 'sky'],
  ['lu', 'branch', 'Branch', 'other', 'tanger', 'git-branch'],
  ['lu', 'package', 'Package', 'other', 'mango'],
  ['lu', 'layers', 'Layers', 'other', 'poppy'],
  ['lu', 'globe', 'Web', 'other', 'sky'],
  ['lu', 'mail', 'Mail', 'other', 'cherry'],
  ['lu', 'palette', 'Palette', 'other', 'cherry'],
  ['lu', 'brush', 'Brush', 'other', 'tanger'],
  ['lu', 'camera', 'Camera', 'other', 'poppy'],
  ['lu', 'scissors', 'Scissors', 'other', 'cherry'],
  ['lu', 'hammer', 'Hammer', 'other', 'mango'],
  ['lu', 'wrench', 'Wrench', 'other', 'sky'],
  ['lu', 'book', 'Book', 'other', 'mango', 'book-open'],
  ['lu', 'pen', 'Pen', 'other', 'poppy', 'pen-tool'],
  ['lu', 'music', 'Music', 'other', 'poppy'],
  ['lu', 'game', 'Games', 'other', 'cherry', 'gamepad-2'],
  ['lu', 'rocket', 'Rocket', 'other', 'tanger'],
  ['lu', 'coffee', 'Coffee', 'other', 'mango'],
  ['lu', 'sparkles', 'Sparkles', 'other', 'yuzu'],
  ['lu', 'star', 'Star', 'other', 'yuzu'],
  ['lu', 'heart', 'Heart', 'other', 'cherry'],
]

const GROUP_LABELS = {
  lang: { cs: 'Jazyky', en: 'Languages' },
  framework: { cs: 'Frameworky', en: 'Frameworks' },
  tool: { cs: 'Nástroje', en: 'Tools' },
  other: { cs: 'Ostatní', en: 'Other' },
}
const GROUP_ORDER = ['lang', 'framework', 'tool', 'other']

// The children of the root <svg> as a single minified string. Keeps every
// element with all its attributes, so multi-path icons and paths carrying
// fill-rule/clip-rule (holes/cutouts) survive intact.
function innerBody(svg, dropTitle) {
  svg = svg.replace(/<!--[\s\S]*?-->/g, '') // license comment
  let inner = svg.replace(/<svg[\s\S]*?>/, '').replace(/<\/svg>\s*$/, '')
  if (dropTitle) inner = inner.replace(/<title>[\s\S]*?<\/title>/g, '')
  return inner.replace(/\s*\n\s*/g, '').replace(/\s{2,}/g, ' ').trim()
}

function siBody(file) {
  const body = innerBody(readFileSync(resolve(SI, `${file}.svg`), 'utf8'), true)
  if (!body.includes('<path')) throw new Error(`no <path> in simple-icons/${file}.svg`)
  return body
}

function luBody(file) {
  return innerBody(readFileSync(resolve(LU, `${file}.svg`), 'utf8'), false)
}

const seen = new Set()
const rows = []
for (const [s, key, name, group, accent, fileOverride] of CATALOG) {
  if (seen.has(key)) throw new Error(`duplicate icon key: ${key}`)
  seen.add(key)
  const file = fileOverride || key
  const type = s === 'si' ? 'fill' : 'stroke'
  const body = s === 'si' ? siBody(file) : luBody(file)
  rows.push({ key, name, group, accent, type, body })
}

const esc = (v) => v.replace(/\\/g, '\\\\').replace(/`/g, '\\`').replace(/\$\{/g, '\\${')
const entries = rows
  .map((r) => `  ${JSON.stringify(r.key)}: { name: ${JSON.stringify(r.name)}, group: '${r.group}', accent: '${r.accent}', type: '${r.type}', body: \`${esc(r.body)}\` },`)
  .join('\n')

const out = `// AUTO-GENERATED by scripts/generate-skill-icons.mjs — do not edit by hand.
// Run \`npm run gen:icons\` to regenerate. Icons are drawn from Simple Icons
// (brand logos, single <path fill>) and Lucide (generic glyphs, stroked), and
// are tinted to a design-palette flavour (\`accent\` → var(--<accent>)) at render.

export type IconGroup = 'lang' | 'framework' | 'tool' | 'other'

export interface SkillIconDef {
  name: string
  group: IconGroup
  accent: string
  type: 'fill' | 'stroke'
  body: string
}

export const ICON_GROUP_ORDER: IconGroup[] = ${JSON.stringify(GROUP_ORDER)}

export const ICON_GROUP_LABELS: Record<IconGroup, { cs: string; en: string }> = ${JSON.stringify(GROUP_LABELS, null, 2)}

export const SKILL_ICONS: Record<string, SkillIconDef> = {
${entries}
}
`

if (!existsSync(dirname(OUT))) mkdirSync(dirname(OUT), { recursive: true })
writeFileSync(OUT, out, 'utf8')
console.log(`✓ wrote ${rows.length} icons → ${OUT.replace(FRONT, '.')}`)
