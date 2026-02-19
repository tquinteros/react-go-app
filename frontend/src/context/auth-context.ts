import { createContext } from "react"

export interface User {
  id: number
  email: string
}

export interface AuthContextType {
  user: User | null
  accessToken: string | null
  login: (token: string, user: User) => void
  logout: () => Promise<void>
  isAuthenticated: boolean
  isInitialized: boolean
}

export const AuthContext = createContext<AuthContextType | null>(null)
