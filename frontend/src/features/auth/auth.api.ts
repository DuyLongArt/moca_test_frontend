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

type ApiEnvelope<T> = {
  status?: string
  code?: number
  message?: string
  data?: T
}

type ClinicAuthData = {
  id?: string | number
  email?: string
  phoneNumber?: string
  phone_number?: string
  fullName?: string
  full_name?: string
  role?: UserRole
  token?: string
  accessToken?: string
}

function mapClinicAuth(
  envelope: ApiEnvelope<ClinicAuthData>,
  fallback: Partial<AuthUser>,
): { user: AuthUser; accessToken: string } {
  const data = envelope.data ?? {}
  return {
    accessToken: data.accessToken ?? data.token ?? '',
    user: {
      id: String(data.id ?? fallback.id ?? ''),
      email: data.email ?? fallback.email ?? '',
      phoneNumber: data.phoneNumber ?? data.phone_number ?? fallback.phoneNumber ?? '',
      fullName: data.fullName ?? data.full_name ?? fallback.fullName ?? '',
      role: data.role ?? fallback.role ?? 'DOCTOR',
    },
  }
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

export async function patientLogin(
  phone: string,
  password: string,
): Promise<{ user: AuthUser; accessToken: string }> {
  if (isMockApi()) {
    return { accessToken: 'dev-bypass', user: { ...getDevUser('PATIENT'), phoneNumber: phone } }
  }
  const { data } = await api.post<ApiEnvelope<ClinicAuthData>>('/api/v1/auth/patient/login', {
    phone_number: phone,
    password,
  })
  return mapClinicAuth(data, {
    phoneNumber: phone,
    role: 'PATIENT',
    fullName: '',
  })
}

export async function patientAuthStatus(phone: string): Promise<{ exists: boolean; hasPassword: boolean }> {
  if (isMockApi()) return { exists: false, hasPassword: false }
  const { data } = await api.post<
    ApiEnvelope<{ exists?: boolean; has_password?: boolean }>
  >('/api/v1/auth/patient/status', { phone_number: phone })
  return {
    exists: data.data?.exists ?? false,
    hasPassword: data.data?.has_password ?? false,
  }
}

export async function patientSignup(payload: {
  phone: string
  email: string
  fullName: string
  password: string
}): Promise<{ user: AuthUser; accessToken: string }> {
  if (isMockApi()) {
    return {
      accessToken: 'dev-bypass',
      user: { ...getDevUser('PATIENT'), phoneNumber: payload.phone, email: payload.email, fullName: payload.fullName },
    }
  }
  const { data } = await api.post<ApiEnvelope<ClinicAuthData>>('/api/v1/auth/patient/signup', {
    phone_number: payload.phone,
    email: payload.email,
    full_name: payload.fullName,
    password: payload.password,
  })
  return mapClinicAuth(data, {
    phoneNumber: payload.phone,
    email: payload.email,
    fullName: payload.fullName,
    role: 'PATIENT',
  })
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
  const { data } = await api.post<ApiEnvelope<ClinicAuthData>>('/api/v1/auth/doctor/login', {
    email,
    password,
  })
  return mapClinicAuth(data, {
    email,
    role: email.includes('admin') ? 'ADMIN' : 'DOCTOR',
    fullName: '',
  })
}

export type PasswordResetRole = 'doctor' | 'patient'

export async function requestPasswordReset(email: string, role: PasswordResetRole = 'doctor'): Promise<void> {
  if (isMockApi()) return
  await api.post('/api/v1/auth/forgot-password', { email, role })
}

export async function resetPassword(
  token: string,
  password: string,
  role: PasswordResetRole = 'doctor',
): Promise<void> {
  if (isMockApi()) return
  await api.post('/api/v1/auth/reset-password', { token, password, role })
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
  const { data } = await api.post<ApiEnvelope<ClinicAuthData>>('/api/v1/auth/doctor/signup', payload)
  return mapClinicAuth(data, {
    email: payload.email,
    fullName: payload.fullName,
    role: 'DOCTOR',
  })
}
