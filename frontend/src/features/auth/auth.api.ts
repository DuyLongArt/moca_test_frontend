import { api } from '../../shared/lib/axios'
import { getDevUser } from '../../shared/lib/devFixtures'
import { isMockApi } from '../../shared/lib/mockApi'
import type { AuthUser, UserRole } from '../../stores/authStore'

export type AuthUserDto = {
  id: string
  email: string
  phoneNumber: string
  fullName: string
  role: UserRole
  gender?: string
  dateOfBirth?: string | null
  educationYears?: number | null
}

type LoginResponse = {
  accessToken: string
  user: AuthUserDto
}

export function mapAuthUserDto(data: AuthUserDto): AuthUser {
  return {
    id: data.id,
    email: data.email,
    phoneNumber: data.phoneNumber,
    fullName: data.fullName,
    role: data.role,
    gender: data.gender || undefined,
    dateOfBirth: data.dateOfBirth ?? null,
    educationYears: data.educationYears ?? null,
  }
}

function mapLogin(data: LoginResponse): { user: AuthUser; accessToken: string } {
  return {
    accessToken: data.accessToken,
    user: mapAuthUserDto(data.user),
  }
}

/** Patient — phone only, no password. */
export async function patientLogin(
  phone: string,
): Promise<{ user: AuthUser; accessToken: string }> {
  if (isMockApi()) {
    return { accessToken: 'dev-bypass', user: { ...getDevUser('PATIENT'), phoneNumber: phone } }
  }
  const { data } = await api.post<LoginResponse>('/api/v1/auth/patient/login', {
    phone_number: phone,
  })
  return mapLogin(data)
}

/** Doctor / admin — email + password. */
export async function doctorLogin(
  email: string,
  password: string,
): Promise<{ user: AuthUser; accessToken: string }> {
  if (isMockApi()) {
    const role: UserRole = email.includes('admin') ? 'ADMIN' : 'DOCTOR'
    return { accessToken: 'dev-bypass', user: { ...getDevUser(role), email } }
  }
  const { data } = await api.post<LoginResponse>('/api/v1/auth/staff/login', {
    email,
    password,
  })
  return mapLogin(data)
}

export async function requestPasswordReset(email: string): Promise<void> {
  if (isMockApi()) return
  await api.post('/api/v1/auth/forgot-password', { email })
}

export async function resetPassword(token: string, password: string): Promise<void> {
  if (isMockApi()) return
  await api.post('/api/v1/auth/reset-password', { token, password })
}

/** Doctor signup — creates account + returns JWT. */
export async function doctorSignup(payload: {
  email: string
  password: string
  fullName: string
  specialty?: string
  licenseNumber?: string
}): Promise<{ user: AuthUser; accessToken: string }> {
  if (isMockApi()) {
    return {
      accessToken: 'dev-bypass',
      user: { ...getDevUser('DOCTOR'), email: payload.email, fullName: payload.fullName },
    }
  }
  const { data } = await api.post<LoginResponse>('/api/v1/auth/staff/signup', payload)
  return mapLogin(data)
}
