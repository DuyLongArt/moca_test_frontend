export function isMockApi(): boolean {
  if (import.meta.env.VITE_USE_MOCK === 'true') return true
  if (import.meta.env.VITE_USE_MOCK === 'false') return false
  return import.meta.env.DEV && import.meta.env.VITE_DEV_BYPASS_AUTH === 'true'
}
