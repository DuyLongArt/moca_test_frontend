import { api } from '../lib/axios'
import { isMockApi } from '../lib/mockApi'
import type { HealthResponse } from '../types/health'

export async function fetchHealth(): Promise<HealthResponse> {
  if (isMockApi()) return { status: 'UP', service: 'moca-platform' }
  const { data } = await api.get<HealthResponse>('/api/v1/health')
  return data
}
