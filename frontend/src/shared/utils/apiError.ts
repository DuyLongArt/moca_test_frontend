type ApiError = {
  response?: {
    data?: string | { message?: unknown } | null
  }
}

export function getApiErrorMessage(error: unknown, fallback: string): string {
  const data = (error as ApiError | null)?.response?.data
  const message = typeof data === 'string' ? data : data?.message
  if (typeof message === 'string' && message.trim()) return message
  return fallback
}
