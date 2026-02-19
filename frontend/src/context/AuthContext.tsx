import { useCallback, useEffect, useState, type ReactNode } from "react"
import { refresh, logout as logoutApi } from "@/features/auth/api"
import { AuthContext } from "./auth-context"

const SESSION_KEY = "auth_session"

function decodeJwtPayload(token: string): { user_id?: number; email?: string } {
  try {
    const base64 = token.split(".")[1]
    if (!base64) return {}
    const json = atob(base64.replace(/-/g, "+").replace(/_/g, "/"))
    return JSON.parse(json) as { user_id?: number; email?: string }
  } catch {
    return {}
  }
}

function loadSession(): { token: string; user: { id: number; email: string } } | null {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const data = JSON.parse(raw) as { access_token: string; user: { id: number; email: string } }
    if (!data?.access_token || !data?.user) return null
    return { token: data.access_token, user: data.user }
  } catch {
    return null
  }
}

function saveSession(token: string, user: { id: number; email: string }) {
  localStorage.setItem(SESSION_KEY, JSON.stringify({ access_token: token, user }))
}

function clearSession() {
  localStorage.removeItem(SESSION_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<{ id: number; email: string } | null>(() =>
    loadSession()?.user ?? null
  )
  const [accessToken, setAccessToken] = useState<string | null>(() =>
    loadSession()?.token ?? null
  )
  const [isInitialized, setIsInitialized] = useState(false)

  const login = useCallback((token: string, u: { id: number; email: string }) => {
    setAccessToken(token)
    setUser(u)
    saveSession(token, u)
  }, [])

  const logout = useCallback(async () => {
    await logoutApi()
    setAccessToken(null)
    setUser(null)
    clearSession()
  }, [])

  useEffect(() => {
    refresh()
      .then((data) => {
        const payload = decodeJwtPayload(data.access_token)
        setAccessToken(data.access_token)
        const u = {
          id: payload.user_id ?? 0,
          email: payload.email ?? "",
        }
        setUser(u)
        saveSession(data.access_token, u)
      })
      .catch(() => {
        const session = loadSession()
        if (session) {
          setAccessToken(session.token)
          setUser(session.user)
        } else {
          setAccessToken(null)
          setUser(null)
        }
      })
      .finally(() => setIsInitialized(true))
  }, [])

  return (
    <AuthContext.Provider
      value={{
        user,
        accessToken,
        login,
        logout,
        isAuthenticated: !!accessToken,
        isInitialized,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
