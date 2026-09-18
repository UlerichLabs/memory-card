import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthApiError } from '@/services/authService'
import { useAuthStore } from '@/store/authStore'
import { AUTH_API_ERROR_MESSAGES, VALIDATION_MESSAGES } from './authConstants'

export function LoginForm() {
  const navigate = useNavigate()
  const { login } = useAuthStore()
  const [email, setEmail] = useState('')
  const [senha, setSenha] = useState('')
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const emailInvalido = !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())
  const emailError = submitted && emailInvalido
  const senhaError = submitted && !senha

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isLoading) return
    setSubmitted(true)
    setApiError(null)
    if (emailInvalido || !senha) return
    setIsLoading(true)
    try {
      await login({ email, senha })
      navigate('/', { replace: true })
    } catch (error: unknown) {
      setApiError(error instanceof AuthApiError && error.codigo !== 'fallback'
        ? AUTH_API_ERROR_MESSAGES[error.codigo] || AUTH_API_ERROR_MESSAGES.loginFallback
        : AUTH_API_ERROR_MESSAGES.loginFallback)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate aria-busy={isLoading}>
      <div className="space-y-1.5">
        <Label htmlFor="login-email">Email</Label>
        <Input id="login-email" name="email" type="email" autoComplete="username" required
          value={email} onChange={(event) => setEmail(event.target.value)} disabled={isLoading}
          aria-invalid={emailError} aria-describedby={emailError ? 'login-email-error' : undefined}
          className="h-11 bg-secondary" />
        {emailError && <p id="login-email-error" role="alert" className="text-xs text-destructive">
          {VALIDATION_MESSAGES.emailInvalido}
        </p>}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="login-senha">Senha</Label>
        <Input id="login-senha" name="senha" type="password" autoComplete="current-password" required
          value={senha} onChange={(event) => setSenha(event.target.value)} disabled={isLoading}
          aria-invalid={senhaError} aria-describedby={senhaError ? 'login-senha-error' : undefined}
          className="h-11 bg-secondary" />
        {senhaError && <p id="login-senha-error" role="alert" className="text-xs text-destructive">
          {VALIDATION_MESSAGES.senhaObrigatoria}
        </p>}
      </div>
      <Button type="submit" disabled={isLoading} className="h-11 w-full rounded-[7px] font-bold">
        {isLoading ? 'Entrando...' : 'Entrar'}
      </Button>
      {apiError && <p role="alert" className="text-xs text-destructive">{apiError}</p>}
    </form>
  )
}
