import { createContext } from 'react'

import type { AuthContextValue } from '@/features/auth/auth.type'

export const AuthContext = createContext<AuthContextValue | null>(null)
