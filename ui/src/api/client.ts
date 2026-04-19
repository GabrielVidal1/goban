export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

const TOKEN_STORAGE_KEY = 'kanban_auth_token'

export function getAuthToken(): string {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem(TOKEN_STORAGE_KEY) ?? ''
}

export function setAuthToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_STORAGE_KEY, token)
  } catch {
    // ignore storage errors (e.g. private browsing)
  }
}

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getAuthToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }

  if (token) {
    headers['X-Auth-Token'] = token
  }

  const res = await fetch(path, {
    headers,
    ...options,
  })

  if (!res.ok) {
    let message = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body.error) message = body.error
    } catch {
      // ignore parse error
    }

    if (res.status === 401) {
      window.dispatchEvent(new CustomEvent('kanban:unauthorized', { detail: message }))
    }

    throw new ApiError(res.status, message)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
