const baseUrl = (import.meta.env.VITE_API_URL || 'http://localhost:18080').replace(/\/+$/, '')

export async function apiGet<T>(path: string, signal?: AbortSignal): Promise<T> {
  const response = await fetch(`${baseUrl}/${path.replace(/^\/+/, '')}`, {
    signal,
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    throw new Error(`A API respondeu com HTTP ${response.status}.`)
  }

  return response.json() as Promise<T>
}

export function getHealth(signal?: AbortSignal) {
  return apiGet<{ status: string }>('/api/v1/health', signal)
}
