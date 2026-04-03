const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8889"

/**
 * Typed fetch wrapper for the Go backend API.
 * Throws an Error with the server's message on non-2xx responses.
 */
export async function apiFetch<T = unknown>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Request failed" }))
    throw new Error(body.error || body.message || `API error: ${res.status}`)
  }

  return res.json() as Promise<T>
}
