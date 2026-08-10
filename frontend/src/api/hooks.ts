import { useQuery } from '@tanstack/react-query'
import { apiFetch } from './client'
import { qk } from './keys'
import type {
  BusinessInfo,
  Category,
  Link,
  ProjectDetail,
  ProjectSummary,
  ScoreEntry,
  Section,
  Skill,
} from './types'

// ---- public read hooks ----

export function useProjects(filters?: { category?: string; featured?: boolean }) {
  const params = new URLSearchParams()
  if (filters?.category) params.set('category', filters.category)
  if (filters?.featured) params.set('featured', 'true')
  const qs = params.toString()
  return useQuery({
    queryKey: qk.projects(filters),
    queryFn: () => apiFetch<ProjectSummary[]>(`/api/projects${qs ? `?${qs}` : ''}`),
  })
}

export function useProject(slug: string) {
  return useQuery({
    queryKey: qk.project(slug),
    queryFn: () => apiFetch<ProjectDetail>(`/api/projects/${encodeURIComponent(slug)}`),
    enabled: !!slug,
  })
}

export function useCategories() {
  return useQuery({ queryKey: qk.categories(), queryFn: () => apiFetch<Category[]>('/api/categories') })
}

export function useLinks() {
  return useQuery({ queryKey: qk.links(), queryFn: () => apiFetch<Link[]>('/api/links') })
}

export function useSection(key: string) {
  return useQuery({
    queryKey: qk.section(key),
    queryFn: () => apiFetch<Section>(`/api/sections/${encodeURIComponent(key)}`),
  })
}

export function useSkills() {
  return useQuery({ queryKey: qk.skills(), queryFn: () => apiFetch<Skill[]>('/api/skills') })
}

export function useBusinessInfo() {
  return useQuery({ queryKey: qk.businessInfo(), queryFn: () => apiFetch<BusinessInfo>('/api/business-info') })
}

export function useLeaderboard(gameKey: string, limit = 10) {
  return useQuery({
    queryKey: qk.leaderboard(gameKey),
    queryFn: () => apiFetch<ScoreEntry[]>(`/api/arcade/${encodeURIComponent(gameKey)}/scores?limit=${limit}`),
    enabled: !!gameKey,
  })
}

// ---- public writes (fire-and-forget / imperative) ----

/** registerClick fires the best-effort link counter; navigation never waits on it. */
export function registerClick(id: number): void {
  void apiFetch<void>(`/api/links/${id}/click`, { method: 'POST' }).catch(() => {})
}
