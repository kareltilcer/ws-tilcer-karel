// TS mirror of openapi.yaml response/request shapes. Bilingual fields arrive as
// *_cs / *_en pairs so the UI switches language with no refetch.

export type PublishStatus = 'draft' | 'published'
export type MessageStatus = 'new' | 'read' | 'archived' | 'spam'

export interface MediaRef {
  id: number
  public_url: string
  alt_cs: string | null
  alt_en: string | null
  width: number | null
  height: number | null
}

export interface Media {
  id: number
  s3_key: string
  public_url: string
  mime: string
  width: number | null
  height: number | null
  size_bytes: number
  alt_cs: string | null
  alt_en: string | null
  created_at: string
}

export interface Category {
  id: number
  slug: string
  name_cs: string
  name_en: string
  sort_order: number
  visible: boolean
}

export interface ProjectLink {
  label: string
  url: string
  type?: string
}

export interface ProjectSummary {
  id: number
  slug: string
  category: Category
  title_cs: string
  title_en: string
  summary_cs: string | null
  summary_en: string | null
  cover: MediaRef | null
  featured: boolean
  status: PublishStatus
  project_date: string | null
  sort_order: number
}

export interface ProjectDetail extends ProjectSummary {
  body_cs: string | null
  body_en: string | null
  gallery: MediaRef[]
  links: ProjectLink[]
  created_at: string
  updated_at: string
}

export interface Link {
  id: number
  label_cs: string
  label_en: string
  url: string
  icon: string | null
  visible: boolean
  sort_order: number
  click_count: number
}

export interface Section {
  key: string
  heading_cs: string | null
  heading_en: string | null
  body_cs: string | null
  body_en: string | null
  media: MediaRef | null
  updated_at: string
}

export interface Skill {
  id: number
  name_cs: string
  name_en: string
  category_cs: string
  category_en: string
  icon: string | null
  level: number | null
  visible: boolean
  sort_order: number
}

export interface BusinessInfo {
  legal_name: string
  ico: string | null
  dic: string | null
  address_lines: string | null
  city: string | null
  zip: string | null
  country: string | null
  register_note_cs: string | null
  register_note_en: string | null
  contact_email: string | null
  contact_phone: string | null
  data_note_cs: string | null
  data_note_en: string | null
  updated_at: string
}

export interface ContactSubmit {
  name: string
  email: string
  subject?: string
  message: string
  locale?: 'cs' | 'en'
  website?: string // honeypot
  turnstile_token?: string
}

export interface ScoreEntry {
  player_name: string
  score: number
  created_at: string
}

export interface ScoreSubmit {
  player_name: string
  score: number
  locale?: 'cs' | 'en'
  website?: string
  turnstile_token?: string
}

export interface ScoreSubmitResult {
  rank: number
  top: ScoreEntry[]
}

export interface ContactMessage {
  id: number
  name: string
  email: string
  subject: string | null
  message: string
  locale: string | null
  ip: string | null
  user_agent: string | null
  status: MessageStatus
  created_at: string
}

export interface Page<T> {
  items: T[]
  next_cursor: string | null
}

export interface SessionUser {
  user: { id: string; email: string; display_name: string | null; roles: string[] }
}
