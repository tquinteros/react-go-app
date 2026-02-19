import { API_URL } from "@/lib/api"

export interface AuthUser {
  id: number
  email: string
}

export interface AuthResponse {
  access_token: string
  user: AuthUser
}

async function authFetch(
  path: string,
  options: RequestInit & { json?: object }
): Promise<Response> {
  const { json, ...rest } = options
  return fetch(`${API_URL}${path}`, {
    ...rest,
    headers: {
      "Content-Type": "application/json",
      ...(rest.headers as HeadersInit),
    },
    credentials: "include",
    body: json ? JSON.stringify(json) : undefined,
  })
}

export async function register(
  email: string,
  password: string
): Promise<AuthResponse> {
  const res = await authFetch("/auth/register", {
    method: "POST",
    json: { email, password },
  })
  const text = await res.text()
  if (!res.ok) {
    throw new Error(
      res.status === 409 ? "Email already exists" : text || "Error registering"
    )
  }
  return JSON.parse(text) as AuthResponse
}

export async function login(
  email: string,
  password: string
): Promise<AuthResponse> {
  const res = await authFetch("/auth/login", {
    method: "POST",
    json: { email, password },
  })
  const text = await res.text()
  if (!res.ok) {
    const message =
      res.status === 401 ? text || "Invalid credentials" : text || "Login failed"
    throw new Error(message)
  }
  return JSON.parse(text) as AuthResponse
}

export async function refresh(): Promise<{ access_token: string }> {
  const res = await authFetch("/auth/refresh", { method: "POST" })
  if (!res.ok) throw new Error("Session expired")
  return res.json()
}

export async function logout(): Promise<void> {
  await authFetch("/auth/logout", { method: "POST" })
}
