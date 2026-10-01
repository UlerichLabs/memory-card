import { useState, type FormEvent } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthApiError, authService } from '@/services/authService'
import { useAuthStore } from '@/store/authStore'
import { AUTH_API_ERROR_MESSAGES, VALIDATION_MESSAGES } from './authConstants'
import { validarSenha } from './cadastroValidation'

type FormErrors = {
  senhaAtual?: string
  novaSenha?: string
  confirmacaoSenha?: string
}

export function TrocarSenhaForm() {
  const { sessao } = useAuthStore()
  const [senhaAtual, setSenhaAtual] = useState('')
  const [novaSenha, setNovaSenha] = useState('')
  const [confirmacaoSenha, setConfirmacaoSenha] = useState('')
  const [errors, setErrors] = useState<FormErrors>({})
  const [apiError, setApiError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  function validate(): FormErrors {
    const nextErrors: FormErrors = {}
    if (!validarSenha(novaSenha)) nextErrors.novaSenha = VALIDATION_MESSAGES.senhaFraca
    if (novaSenha !== confirmacaoSenha) nextErrors.confirmacaoSenha = VALIDATION_MESSAGES.senhasDivergentes
    if (senhaAtual && novaSenha === senhaAtual) {
      nextErrors.novaSenha = AUTH_API_ERROR_MESSAGES['auth.password_change.same_password']
    }
    return nextErrors
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isLoading || !sessao) return
    setApiError(null)
    setSuccess(null)
    const nextErrors = validate()
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) return

    setIsLoading(true)
    try {
      await authService.trocarSenha(sessao.access_token, { senha_atual: senhaAtual, nova_senha: novaSenha })
      setSenhaAtual('')
      setNovaSenha('')
      setConfirmacaoSenha('')
      setErrors({})
      setSuccess('Senha alterada com sucesso.')
    } catch (error: unknown) {
      if (error instanceof AuthApiError) {
        setApiError(AUTH_API_ERROR_MESSAGES[error.codigo] || error.message || AUTH_API_ERROR_MESSAGES.fallback)
      } else {
        setApiError(AUTH_API_ERROR_MESSAGES.fallback)
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate aria-busy={isLoading}>
      <div className="space-y-1.5">
        <Label htmlFor="senha-atual">Senha atual</Label>
        <Input id="senha-atual" name="senha_atual" type="password" autoComplete="current-password" required
          value={senhaAtual} onChange={(event) => setSenhaAtual(event.target.value)} disabled={isLoading}
          aria-invalid={!!errors.senhaAtual} />
        {errors.senhaAtual && <p role="alert" className="text-xs text-destructive">{errors.senhaAtual}</p>}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="nova-senha">Nova senha</Label>
        <Input id="nova-senha" name="nova_senha" type="password" autoComplete="new-password" required
          value={novaSenha} onChange={(event) => setNovaSenha(event.target.value)} disabled={isLoading}
          aria-invalid={!!errors.novaSenha} />
        {errors.novaSenha && <p role="alert" className="text-xs text-destructive">{errors.novaSenha}</p>}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="confirmacao-nova-senha">Confirmação da nova senha</Label>
        <Input id="confirmacao-nova-senha" name="confirmacao_nova_senha" type="password" autoComplete="new-password" required
          value={confirmacaoSenha} onChange={(event) => setConfirmacaoSenha(event.target.value)} disabled={isLoading}
          aria-invalid={!!errors.confirmacaoSenha} />
        {errors.confirmacaoSenha && <p role="alert" className="text-xs text-destructive">{errors.confirmacaoSenha}</p>}
      </div>
      {apiError && <p role="alert" className="text-xs text-destructive">{apiError}</p>}
      {success && <p role="status" className="text-xs text-green-400">{success}</p>}
      <Button type="submit" disabled={isLoading} className="h-11 w-full rounded-[7px] font-bold">
        {isLoading ? 'Alterando...' : 'Alterar senha'}
      </Button>
    </form>
  )
}
