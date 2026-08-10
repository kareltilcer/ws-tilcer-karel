import { useQuery } from '@tanstack/react-query'
import { apiFetch, postForm } from '../api/client'
import { queryClient } from '../api/queryClient'
import { qk } from '../api/keys'
import type {
  BusinessInfo,
  Category,
  ContactMessage,
  Link,
  Media,
  MessageStatus,
  Page,
  ProjectDetail,
  Section,
  Skill,
  SessionUser,
} from '../api/types'

// Invalidate both the admin key and the matching public key after a mutation.
function invalidate(keys: readonly unknown[][]) {
  keys.forEach((k) => void queryClient.invalidateQueries({ queryKey: k }))
}

// ---- session / auth ----
export function useSession() {
  return useQuery({
    queryKey: qk.session(),
    queryFn: () => apiFetch<SessionUser>('/api/auth/session', { skipAuthRedirect: true }),
    retry: false,
    staleTime: 0,
  })
}
export async function login(email: string, password: string) {
  const r = await apiFetch<SessionUser>('/api/auth/login', { method: 'POST', body: { email, password }, skipAuthRedirect: true })
  invalidate([[...qk.session()]])
  return r
}
export async function logout() {
  await apiFetch<void>('/api/auth/logout', { method: 'POST', skipAuthRedirect: true })
  invalidate([[...qk.session()]])
}

// ---- projects ----
export function useAdminProjects(status?: string) {
  const qs = status ? `?status=${status}` : ''
  return useQuery({
    queryKey: qk.adminProjects({ status }),
    queryFn: () => apiFetch<Page<ProjectDetail>>(`/api/admin/projects${qs}`),
  })
}
const projInval = [[...qk.adminProjects()], ['admin', 'projects'], ['projects'], ['project']]
export const createProject = async (body: unknown) => {
  const r = await apiFetch<ProjectDetail>('/api/admin/projects', { method: 'POST', body })
  invalidate(projInval)
  return r
}
export const updateProject = async (id: number, body: unknown) => {
  const r = await apiFetch<ProjectDetail>(`/api/admin/projects/${id}`, { method: 'PATCH', body })
  invalidate(projInval)
  return r
}
export const deleteProject = async (id: number) => {
  await apiFetch<void>(`/api/admin/projects/${id}`, { method: 'DELETE' })
  invalidate(projInval)
}
export const reorderProjects = async (ids: number[]) => {
  await apiFetch<void>('/api/admin/projects/reorder', { method: 'POST', body: { ids } })
  invalidate(projInval)
}

// ---- categories ----
export function useAdminCategories() {
  return useQuery({ queryKey: qk.adminCategories(), queryFn: () => apiFetch<Category[]>('/api/admin/categories') })
}
const catInval = [[...qk.adminCategories()], ['categories']]
export const createCategory = async (body: unknown) => {
  const r = await apiFetch<Category>('/api/admin/categories', { method: 'POST', body })
  invalidate(catInval)
  return r
}
export const updateCategory = async (id: number, body: unknown) => {
  const r = await apiFetch<Category>(`/api/admin/categories/${id}`, { method: 'PATCH', body })
  invalidate(catInval)
  return r
}
export const deleteCategory = async (id: number) => {
  await apiFetch<void>(`/api/admin/categories/${id}`, { method: 'DELETE' })
  invalidate(catInval)
}
export const reorderCategories = async (ids: number[]) => {
  await apiFetch<void>('/api/admin/categories/reorder', { method: 'POST', body: { ids } })
  invalidate(catInval)
}

// ---- links ----
export function useAdminLinks() {
  return useQuery({ queryKey: qk.adminLinks(), queryFn: () => apiFetch<Link[]>('/api/admin/links') })
}
const linkInval = [[...qk.adminLinks()], ['links']]
export const createLink = async (body: unknown) => {
  const r = await apiFetch<Link>('/api/admin/links', { method: 'POST', body })
  invalidate(linkInval)
  return r
}
export const updateLink = async (id: number, body: unknown) => {
  const r = await apiFetch<Link>(`/api/admin/links/${id}`, { method: 'PATCH', body })
  invalidate(linkInval)
  return r
}
export const deleteLink = async (id: number) => {
  await apiFetch<void>(`/api/admin/links/${id}`, { method: 'DELETE' })
  invalidate(linkInval)
}
export const reorderLinks = async (ids: number[]) => {
  await apiFetch<void>('/api/admin/links/reorder', { method: 'POST', body: { ids } })
  invalidate(linkInval)
}

// ---- sections ----
export function useSectionAdmin(key: string) {
  return useQuery({ queryKey: qk.section(key), queryFn: () => apiFetch<Section>(`/api/sections/${key}`) })
}
export const saveSection = async (key: string, body: unknown) => {
  const r = await apiFetch<Section>(`/api/admin/sections/${key}`, { method: 'PUT', body })
  invalidate([['section', key]])
  return r
}

// ---- skills ----
export function useAdminSkills() {
  return useQuery({ queryKey: ['admin', 'skills'], queryFn: () => apiFetch<Skill[]>('/api/admin/skills') })
}
const skillInval = [['admin', 'skills'], ['skills']]
export const createSkill = async (body: unknown) => {
  const r = await apiFetch<Skill>('/api/admin/skills', { method: 'POST', body })
  invalidate(skillInval)
  return r
}
export const updateSkill = async (id: number, body: unknown) => {
  const r = await apiFetch<Skill>(`/api/admin/skills/${id}`, { method: 'PATCH', body })
  invalidate(skillInval)
  return r
}
export const deleteSkill = async (id: number) => {
  await apiFetch<void>(`/api/admin/skills/${id}`, { method: 'DELETE' })
  invalidate(skillInval)
}
export const reorderSkills = async (ids: number[]) => {
  await apiFetch<void>('/api/admin/skills/reorder', { method: 'POST', body: { ids } })
  invalidate(skillInval)
}

// ---- business info ----
export function useBusinessAdmin() {
  return useQuery({ queryKey: qk.businessInfo(), queryFn: () => apiFetch<BusinessInfo>('/api/business-info') })
}
export const saveBusiness = async (body: unknown) => {
  const r = await apiFetch<BusinessInfo>('/api/admin/business-info', { method: 'PUT', body })
  invalidate([[...qk.businessInfo()]])
  return r
}

// ---- media ----
export function useAdminMedia(cursor = '') {
  const qs = cursor ? `?cursor=${cursor}` : ''
  return useQuery({ queryKey: qk.adminMedia(cursor), queryFn: () => apiFetch<Page<Media>>(`/api/admin/media${qs}`) })
}
export const uploadMedia = async (form: FormData) => {
  const r = await postForm<Media>('/api/admin/media', form)
  invalidate([['admin', 'media']])
  return r
}
export const deleteMedia = async (id: number) => {
  await apiFetch<void>(`/api/admin/media/${id}`, { method: 'DELETE' })
  invalidate([['admin', 'media']])
}

// ---- contact inbox ----
export function useAdminContact(status?: string) {
  const qs = status ? `?status=${status}` : ''
  return useQuery({ queryKey: qk.adminContact({ status }), queryFn: () => apiFetch<Page<ContactMessage>>(`/api/admin/contact${qs}`) })
}
export const setMessageStatus = async (id: number, status: MessageStatus) => {
  const r = await apiFetch<ContactMessage>(`/api/admin/contact/${id}`, { method: 'PATCH', body: { status } })
  invalidate([['admin', 'contact']])
  return r
}
export const deleteMessage = async (id: number) => {
  await apiFetch<void>(`/api/admin/contact/${id}`, { method: 'DELETE' })
  invalidate([['admin', 'contact']])
}

// ---- arcade moderation ----
export interface AdminScore {
  id: number
  player_name: string
  score: number
  created_at: string
}
export function useAdminScores(gameKey: string) {
  return useQuery({ queryKey: ['admin', 'scores', gameKey], queryFn: () => apiFetch<AdminScore[]>(`/api/admin/arcade/${gameKey}/scores?limit=100`) })
}
export const deleteScore = async (id: number, gameKey: string) => {
  await apiFetch<void>(`/api/admin/arcade/scores/${id}`, { method: 'DELETE' })
  invalidate([['admin', 'scores', gameKey], ['leaderboard', gameKey]])
}
export const resetBoard = async (gameKey: string) => {
  await apiFetch<void>(`/api/admin/arcade/${gameKey}/reset`, { method: 'POST' })
  invalidate([['admin', 'scores', gameKey], ['leaderboard', gameKey]])
}
