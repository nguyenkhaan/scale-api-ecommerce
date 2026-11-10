import { useContext } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'

import { AuthContext } from '@/context/AuthContext'

import { getCurrentUser, login, logout, register } from './auth.service'
import type { AuthContextValue, Role } from './auth.type'

export const authQueryKeys = {
  currentUser: ['auth', 'me'] as const,
}

export function useCurrentUserQuery() {
  return useQuery({
    queryKey: authQueryKeys.currentUser,
    queryFn: getCurrentUser,
    retry: false,
  })
}

export function useLoginMutation() {
  return useMutation({
    mutationKey: ['auth', 'login'],
    mutationFn: login,
  })
}

export function useRegisterMutation() {
  return useMutation({
    mutationKey: ['auth', 'register'],
    mutationFn: register,
  })
}

export function useLogoutMutation() {
  return useMutation({
    mutationKey: ['auth', 'logout'],
    mutationFn: logout,
  })
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext)
  if (value === null) throw new Error('useAuth must be used inside AuthProvider')
  return value
}

export const roleLabels: Record<Role, string> = {
  customer: 'Khách hàng',
  admin: 'Quản trị viên',
}

export function safeNextPath(next: string | null): string {
  if (next === null || !next.startsWith('/') || next.startsWith('//') || next.includes('\\') || hasControlCharacter(next)) {
    return '/'
  }
  return next
}

export function loginRedirectPath(pathname: string, search: string): string {
  return `/login?next=${encodeURIComponent(`${pathname}${search}`)}`
}

function hasControlCharacter(value: string): boolean {
  for (let index = 0; index < value.length; index++) {
    const code = value.charCodeAt(index)
    if (code < 0x20 || code === 0x7f) return true
  }
  return false
}
