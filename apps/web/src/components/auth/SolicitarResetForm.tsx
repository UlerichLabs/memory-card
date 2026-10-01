import { useState, type FormEvent } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthApiError, authService } from '@/services/authService'
import { validarEmail } from './cadastroValidation'
import { RECUPERACAO_SENHA_MENSAGENS } from './recuperacaoSenhaConstants'

export function SolicitarResetForm() {
  const [email, setEmail] = useState('')
  const [submetido, setSubmetido] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [mensagem, setMensagem] = useState<string | null>(null)
  const emailInvalido = submetido && !validarEmail(email)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isLoading) return
    setSubmetido(true)
    setMensagem(null)
    if (!validarEmail(email)) return
    setIsLoading(true)
    try {
      await authService.solicitarReset({ email: email.trim() })
      setMensagem(RECUPERACAO_SENHA_MENSAGENS.sucessoSolicitacao)
    } catch (error: unknown) {
      setMensagem(error instanceof AuthApiError && error.status === 429
        ? RECUPERACAO_SENHA_MENSAGENS.rateLimit
        : RECUPERACAO_SENHA_MENSAGENS.sucessoSolicitacao)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate aria-busy={isLoading}>
      <div className="space-y-1.5">
        <Label htmlFor="reset-email">Email</Label>
        <Input id="reset-email" name="email" type="email" autoComplete="email" required
          value={email} onChange={(event) => setEmail(event.target.value)} disabled={isLoading}
          aria-invalid={emailInvalido} aria-describedby={emailInvalido ? 'reset-email-error' : undefined}
          className="h-11 bg-secondary" />
        {emailInvalido && <p id="reset-email-error" role="alert" className="text-xs text-destructive">
          Informe um email válido.
        </p>}
      </div>
      <Button type="submit" disabled={isLoading} className="h-11 w-full rounded-[7px] font-bold">
        {isLoading ? 'Enviando...' : 'Enviar instruções'}
      </Button>
      {mensagem && <p role={mensagem === RECUPERACAO_SENHA_MENSAGENS.rateLimit ? 'alert' : 'status'}
        className={mensagem === RECUPERACAO_SENHA_MENSAGENS.rateLimit ? 'text-xs text-destructive' : 'text-xs text-primary'}>
        {mensagem}
      </p>}
    </form>
  )
}
