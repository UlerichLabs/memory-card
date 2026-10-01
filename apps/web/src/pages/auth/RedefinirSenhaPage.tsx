import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { RedefinirSenhaForm } from '@/components/auth/RedefinirSenhaForm'
import { mensagemTokenReset, RECUPERACAO_SENHA_MENSAGENS, type EstadoTokenReset } from '@/components/auth/recuperacaoSenhaConstants'
import { useValidacaoTokenReset } from '@/hooks/useValidacaoTokenReset'
import { authTheme } from './authTheme'

function ErroTokenReset({ estado }: { estado: EstadoTokenReset | 'erro' }) {
  const mensagem = estado === 'erro' ? RECUPERACAO_SENHA_MENSAGENS.falhaValidacao : mensagemTokenReset(estado)
  return (
    <section className="w-full max-w-md space-y-6 rounded-xl border bg-card p-5 text-center">
      <h1 className="text-[22px] font-bold">Não foi possível redefinir sua senha</h1>
      <p role="alert" className="text-[13px] text-destructive">{mensagem}</p>
      <Link to="/esqueci-senha" className="font-medium text-primary underline-offset-4 hover:underline">Solicitar novo link</Link>
    </section>
  )
}

export function RedefinirSenhaPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') || ''
  const [estadoSubmissao, setEstadoSubmissao] = useState<EstadoTokenReset | null>(null)
  const estadoInicial = useValidacaoTokenReset(token)
  const estado = estadoSubmissao || estadoInicial

  return (
    <main style={authTheme} className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      {estado === 'carregando' && <p role="status" className="text-[13px] text-muted-foreground">Validando link...</p>}
      {estado === 'valido' && <section className="w-full max-w-md space-y-6 rounded-xl border bg-card p-5">
        <header className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Memory Card</p>
          <h1 className="text-[22px] font-bold">Redefinir senha</h1>
          <p className="text-[13px] text-muted-foreground">Crie uma nova senha para sua conta.</p>
        </header>
        <RedefinirSenhaForm token={token} onTokenInvalido={setEstadoSubmissao} />
      </section>}
      {(['invalido', 'expirado', 'utilizado', 'erro'] as const).includes(estado as EstadoTokenReset | 'erro') &&
        <ErroTokenReset estado={estado as EstadoTokenReset | 'erro'} />}
    </main>
  )
}
