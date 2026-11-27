import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import {
  doctorLogin as doctorLoginApi,
  doctorSignup as doctorSignupApi,
  patientLogin as patientLoginApi,
  patientSignup as patientSignupApi,
} from '../features/auth/auth.api'
import { queryClient } from '../shared/lib/queryClient'

export type UserRole = 'PATIENT' | 'DOCTOR' | 'ADMIN'

export type AuthUser = {
  id: string
  fullName: string
  email: string
  phoneNumber: string
  role: UserRole
  gender?: string
  dateOfBirth?: string | null
  educationYears?: number | null
}

type AuthState = {
  user: AuthUser | null
  token: string | null
  /** Patient: phone + password → POST /api/v1/auth/patient/login */
  patientLogin: (phone: string, password: string) => Promise<void>
  patientSignup: (payload: {
    phone: string
    email: string
    fullName: string
    password: string
  }) => Promise<void>
  /** Clinician / admin: email + password → POST /api/v1/auth/doctor/login */
  doctorLogin: (email: string, password: string) => Promise<void>
  /** Clinician signup → POST /api/v1/auth/doctor/signup */
  doctorSignup: (payload: {
    email: string
    password: string
    fullName: string
    specialty?: string
    licenseNumber?: string
  }) => Promise<void>
  updateUser: (user: AuthUser) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => {
      const authenticate = (user: AuthUser, token: string) => {
        set({ user, token })
        queryClient.clear()
      }

      return {
        user: null,
        token: null,
        patientLogin: async (phone, password) => {
          const { user, accessToken } = await patientLoginApi(phone, password)
          authenticate(user, accessToken)
        },
        patientSignup: async (payload) => {
          const { user, accessToken } = await patientSignupApi(payload)
          authenticate(user, accessToken)
        },
        doctorLogin: async (email, password) => {
          const { user, accessToken } = await doctorLoginApi(email, password)
          authenticate(user, accessToken)
        },
        doctorSignup: async (payload) => {
          const { user, accessToken } = await doctorSignupApi(payload)
          const token = accessToken || (await doctorLoginApi(payload.email, payload.password)).accessToken
          authenticate(user, token)
        },
        updateUser: (user) => set({ user }),
        logout: () => {
          set({ user: null, token: null })
          queryClient.clear()
        },
      }
    },
    { name: 'moca-auth' },
  ),
)
