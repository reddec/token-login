/**
 * Extract a human-readable error message from whatever @hey-api/client-fetch throws.
 *
 * The error could be:
 * - A string (server returned plain text body)
 * - A JSON object with .error_message, .error, or .message fields (ogen backend)
 * - An Error instance (network failure)
 * - An empty object {}
 */
export function getErrorMessage(e: unknown, fallback = 'An unexpected error occurred'): string {
  if (!e) return fallback
  if (typeof e === 'string') return e.trim() || fallback
  if (e instanceof Error) return e.message || fallback
  if (typeof e === 'object') {
    const obj = e as Record<string, unknown>
    if (typeof obj.error_message === 'string') return obj.error_message
    if (typeof obj.error === 'string') return obj.error
    if (typeof obj.message === 'string') return obj.message
    // try to stringify the object
    try {
      const s = JSON.stringify(e)
      if (s !== '{}') return s
    } catch { /* ignore */ }
  }
  return fallback
}

/**
 * Safely extract an HTTP status from an error thrown by @hey-api/client-fetch.
 * Returns undefined when the error has no recognizable response shape.
 */
export function getResponseStatus(e: unknown): number | undefined {
  if (!(e && typeof e === 'object' && 'response' in e)) {
    return undefined
  }
  const response = e.response
  if (!(response && typeof response === 'object' && 'status' in response)) {
    return undefined
  }
  return typeof response.status === 'number' ? response.status : undefined
}
