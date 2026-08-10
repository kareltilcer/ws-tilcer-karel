// The API fetch wrapper (Mode B). Same-origin: paths are relative ("/api/…"), so
// no CORS; in dev Vite proxies /api to the Go backend. Admin auth is the session
// cookie (sent automatically) + a double-submit CSRF token read from the JS-
// readable `csrf` cookie. On 401 the admin app routes to the login screen.

let onUnauthorized: (() => void) | null = null
let authRedirected = false

export function setUnauthorizedHandler(fn: (() => void) | null): void {
  onUnauthorized = fn
}

// Fire the 401 handler at most once until the next successful response. A burst
// of concurrent requests that all 401 (e.g. a batch upload after the session
// expires) then triggers a single redirect instead of one per in-flight request.
function handleUnauthorized(): void {
  if (authRedirected) return
  authRedirected = true
  onUnauthorized?.()
}

function csrfToken(): string {
  const m = document.cookie.match(/(?:^|;\s*)csrf=([^;]+)/)
  return m ? decodeURIComponent(m[1]) : ''
}

const SAFE = new Set(['GET', 'HEAD', 'OPTIONS'])

export class ApiError extends Error {
  status: number
  code: string
  detail?: string
  constructor(status: number, code: string, detail?: string) {
    super(detail ? `${code}: ${detail}` : code)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.detail = detail
  }
}

export interface RequestOptions {
  method?: string
  body?: unknown
  signal?: AbortSignal
  skipAuthRedirect?: boolean
}

export async function apiFetch<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const method = opts.method ?? 'GET'
  const headers: Record<string, string> = {}
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json'
  if (!SAFE.has(method)) {
    const t = csrfToken()
    if (t) headers['X-CSRF-Token'] = t
  }

  const res = await fetch(path, {
    method,
    credentials: 'include',
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    signal: opts.signal,
  })

  if (res.status === 401 && !opts.skipAuthRedirect) {
    handleUnauthorized()
    throw new ApiError(401, 'unauthorized')
  }
  if (!res.ok) {
    let code = 'error'
    let detail: string | undefined
    try {
      const parsed = (await res.json()) as { error?: string; detail?: string }
      code = parsed.error ?? code
      detail = parsed.detail
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(res.status, code, detail)
  }
  authRedirected = false // a successful response means auth is valid again
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

/** postForm sends multipart/form-data (media upload) with the CSRF header. */
export async function postForm<T>(path: string, form: FormData): Promise<T> {
  const headers: Record<string, string> = {}
  const t = csrfToken()
  if (t) headers['X-CSRF-Token'] = t
  const res = await fetch(path, { method: 'POST', credentials: 'include', headers, body: form })
  if (res.status === 401) {
    handleUnauthorized()
    throw new ApiError(401, 'unauthorized')
  }
  if (!res.ok) {
    let code = 'error'
    let detail: string | undefined
    try {
      const parsed = (await res.json()) as { error?: string; detail?: string }
      code = parsed.error ?? code
      detail = parsed.detail
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, code, detail)
  }
  authRedirected = false // a successful response means auth is valid again
  return (await res.json()) as T
}
