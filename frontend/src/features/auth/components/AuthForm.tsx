import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import { login, register } from "../api"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function AuthForm() {
  const [isLogin, setIsLogin] = useState(true)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const { login: setAuth } = useAuth()
  const navigate = useNavigate()

  const mutation = useMutation({
    mutationFn: () =>
      isLogin ? login(email, password) : register(email, password),
    onSuccess: (data) => {
      setAuth(data.access_token, data.user)
      navigate("/", { replace: true })
    },
  })

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    mutation.mutate()
  }

  return (
    <div className="mx-auto max-w-sm rounded-xl border border-border bg-card p-6 shadow-sm">
      <h2 className="mb-6 text-2xl font-semibold">
        {isLogin ? "Login" : "Register"}
      </h2>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="grid gap-2">
          <Label htmlFor="auth-email">Email</Label>
          <Input
            id="auth-email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="auth-password">Password</Label>
          <Input
            id="auth-password"
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete={isLogin ? "current-password" : "new-password"}
          />
        </div>
        <Button type="submit" disabled={mutation.isPending} className="w-full">
          {mutation.isPending
            ? "Loading…"
            : isLogin
              ? "Login"
              : "Create account"}
        </Button>

        {mutation.isError && (
          <p className="text-destructive text-sm" role="alert">
            {mutation.error.message}
          </p>
        )}
      </form>

      <button
        type="button"
        onClick={() => {
          setIsLogin(!isLogin)
          mutation.reset()
        }}
        className="mt-4 text-sm text-muted-foreground underline hover:text-foreground"
      >
        {isLogin
          ? "Don't have an account? Register"
          : "Already have an account? Login"}
      </button>
    </div>
  )
}
