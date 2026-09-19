import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthApiError, authService } from '@/services/authService'
import { validarSenha } from './cadastroValidation'
import { estadoTokenReset, RECUPERACAO_SENHA_MENSAGENS, type EstadoTokenReset } from './recuperacaoSenhaConstants'

type RedefinirSenhaFormProps = {
  token: string
  onTokenInvalido: (estado: EstadoTokenReset) => void
}

export function RedefinirSenhaForm({ token, onTokenInvalido }: RedefinirSenhaFormProps) {
  const navigate = useNavigate()
  const [senha, setSenha] = useState('')
  const [confirmacaoSenha, setConfirmacaoSenha] = useState('')
  const [submetido, setSubmetido] = useState(false)
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const senhaInvalida = submetido && !validarSenha(senha)
  const confirmacaoInvalida = submetido && senha !== confirmacaoSenha

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isLoading) return
    setSubmetido(true)
    setApiError(null)
    if (!validarSenha(senha) || senha !== confirmacaoSenha) return
    setIsLoading(true)
    try {
      await authService.redefinirSenha({ token, senha })
      navigate('/login', { replace: true, state: { mensagem: RECUPERACAO_SENHA_MENSAGENS.sucessoRedefinicao } })
    } catch (error: unknown) {
      if (error instanceof AuthApiError && error.codigo === 'auth.password_reset.weak_password') {
        setApiError(error.message || RECUPERACAO_SENHA_MENSAGENS.senhaFraca)
      } else {
        const estado = estadoTokenReset(error)
        if (estado) onTokenInvalido(estado)
        else setApiError(RECUPERACAO_SENHA_MENSAGENS.falhaRedefinicao)
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate aria-busy={isLoading}>
      <div className="space-y-1.5">
        <Label htmlFor="nova-senha">Nova senha</Label>
        <Input id="nova-senha" name="senha" type="password" autoComplete="new-password" required
          value={senha} onChange={(event) => setSenha(event.target.value)} disabled={isLoading}
          aria-invalid={senhaInvalida} className="h-11 bg-secondary" />
        {senhaInvalida && <p role="alert" className="text-xs text-destructive">{RECUPERACAO_SENHA_MENSAGENS.senhaFraca}</p>}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="confirmacao-nova-senha">Confirmação de senha</Label>
        <Input id="confirmacao-nova-senha" name="confirmacaoSenha" type="password" autoComplete="new-password" required
          value={confirmacaoSenha} onChange={(event) => setConfirmacaoSenha(event.target.value)} disabled={isLoading}
          aria-invalid={confirmacaoInvalida} className="h-11 bg-secondary" />
        {confirmacaoInvalida && <p role="alert" className="text-xs text-destructive">Senha e confirmação não conferem.</p>}
      </div>
      {apiError && <p role="alert" className="text-xs text-destructive">{apiError}</p>}
      <Button type="submit" disabled={isLoading} className="h-11 w-full rounded-[7px] font-bold">
        {isLoading ? 'Redefinindo...' : 'Redefinir senha'}
      </Button>
    </form>
  )
}
