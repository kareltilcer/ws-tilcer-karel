// TanStack Query keys (PRD §7.4). Public reads are cache-friendly (long stale);
// each admin mutation invalidates its public + admin keys.
export const qk = {
  projects: (filters?: { category?: string; featured?: boolean }) => ['projects', filters ?? {}] as const,
  project: (slug: string) => ['project', slug] as const,
  categories: () => ['categories'] as const,
  links: () => ['links'] as const,
  section: (key: string) => ['section', key] as const,
  skills: () => ['skills'] as const,
  businessInfo: () => ['business-info'] as const,
  leaderboard: (gameKey: string) => ['leaderboard', gameKey] as const,
  // admin
  adminProjects: (filters?: unknown) => ['admin', 'projects', filters ?? {}] as const,
  adminCategories: () => ['admin', 'categories'] as const,
  adminLinks: () => ['admin', 'links'] as const,
  adminMedia: (cursor?: string) => ['admin', 'media', cursor ?? ''] as const,
  adminContact: (filters?: unknown) => ['admin', 'contact', filters ?? {}] as const,
  session: () => ['session'] as const,
} as const
