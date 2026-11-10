import { useMemo, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'

import { authQueryKeys, useCurrentUserQuery, useLoginMutation, useLogoutMutation, useRegisterMutation } from '@/features/auth/auth.hook'
import { getCurrentUser, isAuthApiError } from '@/features/auth/auth.service'
import type { AuthContextValue, AuthState } from '@/features/auth/auth.type'
import { AuthContext } from './AuthContext'

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const currentUserQuery = useCurrentUserQuery()
  const loginMutation = useLoginMutation()
  const registerMutation = useRegisterMutation()
  const logoutMutation = useLogoutMutation()

  const state: AuthState = useMemo(() => {
    if (currentUserQuery.isPending) return { status: 'loading' }
    if (currentUserQuery.data !== null && currentUserQuery.data !== undefined) {
      return { status: 'authenticated', user: currentUserQuery.data }
    }
    return { status: 'anonymous' }
  }, [currentUserQuery.data, currentUserQuery.isPending])

  const value: AuthContextValue = useMemo(
    () => ({
      state,
      async refresh() {
        await queryClient.invalidateQueries({ queryKey: authQueryKeys.currentUser })
        try {
          await queryClient.fetchQuery({
            queryKey: authQueryKeys.currentUser,
            queryFn: getCurrentUser,
            retry: false,
          })
        } catch {
          // A failed session refresh is rendered as anonymous, matching the previous auth flow.
        }
      },
      async login(credentials) {
        await queryClient.cancelQueries({ queryKey: authQueryKeys.currentUser })
        const result = await loginMutation.mutateAsync(credentials)
        queryClient.setQueryData(authQueryKeys.currentUser, result.user)
        return result.user
      },
      async register(credentials) {
        await queryClient.cancelQueries({ queryKey: authQueryKeys.currentUser })
        const result = await registerMutation.mutateAsync(credentials)
        queryClient.setQueryData(authQueryKeys.currentUser, result.user)
        return result.user
      },
      async logout() {
        try {
          await logoutMutation.mutateAsync()
          queryClient.setQueryData(authQueryKeys.currentUser, null)
        } catch (error) {
          if (isAuthApiError(error) && error.status === 401) {
            queryClient.setQueryData(authQueryKeys.currentUser, null)
            return
          }

          try {
            const user = await getCurrentUser()
            if (user === null) {
              queryClient.setQueryData(authQueryKeys.currentUser, null)
              return
            }
          } catch {
            // Keep the original logout error when the session cannot be rechecked.
          }
          throw error
        }
      },
    }),
    [loginMutation, logoutMutation, queryClient, registerMutation, state],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
