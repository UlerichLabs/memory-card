import { useEffect, useState } from 'react'
import { authService } from '@/services/authService'
import { estadoTokenReset, type EstadoTokenReset } from '@/components/auth/recuperacaoSenhaConstants'

export type ValidacaoTokenReset = 'carregando' | 'valido' | EstadoTokenReset | 'erro'

export function useValidacaoTokenReset(token: string): ValidacaoTokenReset {
  const [resultado, setResultado] = useState({
    token,
    estado: token ? 'carregando' as ValidacaoTokenReset : 'invalido' as ValidacaoTokenReset,
  })
  const estado = resultado.token === token
    ? resultado.estado
    : token ? 'carregando' : 'invalido'

  useEffect(() => {
    let ativo = true
    if (!token) return () => { ativo = false }
    async function validar() {
      try {
        await authService.validarTokenReset(token)
        if (ativo) setResultado({ token, estado: 'valido' })
      } catch (error: unknown) {
        if (ativo) setResultado({ token, estado: estadoTokenReset(error) || 'erro' })
      }
    }
    void validar()
    return () => { ativo = false }
  }, [token])

  return estado as ValidacaoTokenReset
}
