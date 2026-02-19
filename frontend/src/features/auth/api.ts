import { API_URL } from "@/lib/api"

export interface AuthUser {
  id: number
  email: string
}

export interface AuthResponse {
  access_token: string
  user: AuthUser
}

export async function register(
  email: string,
  password: string
): Promise<AuthResponse> {
  const res = await fetch(`${API_URL}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password }),
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
  const res = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password }),
  })
  const text = await res.text()
  if (!res.ok) {
    throw new Error(text || "Invalid credentials")
  }
  return JSON.parse(text) as AuthResponse
}

export async function refresh(): Promise<{ access_token: string }> {
  const res = await fetch(`${API_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  })
  if (!res.ok) throw new Error("Session expired")
  return res.json()
}

export async function logout(): Promise<void> {
  await fetch(`${API_URL}/auth/logout`, {
    method: "POST",
    credentials: "include",
  })
}
