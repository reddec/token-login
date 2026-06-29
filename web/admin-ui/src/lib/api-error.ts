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
