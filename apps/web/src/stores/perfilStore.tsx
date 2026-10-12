import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { perfilService } from '@/lib/services/perfilService'
import { AuthContext } from '@/store/authStore'
import type { AtualizarPerfilPayload, PerfilDTO } from '@/types/perfil'

export interface PerfilStore {
  perfil: PerfilDTO | null
  isLoading: boolean
  isSalvando: boolean
  error: string | null
  carregar: (signal?: AbortSignal) => Promise<PerfilDTO | null>
  salvar: (payload: AtualizarPerfilPayload) => Promise<PerfilDTO>
  enviarFoto: (arquivo: File) => Promise<PerfilDTO>
  usarCapa: (jogoId: number) => Promise<PerfilDTO>
  removerFoto: () => Promise<PerfilDTO>
  limparErro: () => void
}

export const PerfilContext = createContext<PerfilStore | null>(null)

export function PerfilProvider({ children, token: tokenProp }: { children: ReactNode; token?: string }) {
  const auth = useContext(AuthContext)
  const token = tokenProp ?? auth?.sessao?.access_token

  const [perfil, setPerfil] = useState<PerfilDTO | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isSalvando, setIsSalvando] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function carregar(signal?: AbortSignal): Promise<PerfilDTO | null> {
    if (!token) return null
    setIsLoading(true)
    setError(null)
    try {
      const dados = await perfilService.obter(token, signal)
      setPerfil(dados)
      return dados
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return null
      const msg = err instanceof Error ? err.message : 'Erro ao carregar perfil'
      setError(msg)
      return null
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (token) {
      const controller = new AbortController()
      void carregar(controller.signal)
      return () => controller.abort()
    } else {
      setPerfil(null)
    }
  }, [token])

  async function salvar(payload: AtualizarPerfilPayload): Promise<PerfilDTO> {
    setIsSalvando(true)
    setError(null)
    try {
      const dados = await perfilService.atualizar(payload, token)
      setPerfil(dados)
      return dados
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Erro ao salvar perfil'
      setError(msg)
      throw err
    } finally {
      setIsSalvando(false)
    }
  }

  async function enviarFoto(arquivo: File): Promise<PerfilDTO> {
    setIsSalvando(true)
    setError(null)
    try {
      const dados = await perfilService.enviarFoto(arquivo, token)
      setPerfil(dados)
      return dados
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Erro ao enviar foto'
      setError(msg)
      throw err
    } finally {
      setIsSalvando(false)
    }
  }

  async function usarCapa(jogoId: number): Promise<PerfilDTO> {
    setIsSalvando(true)
    setError(null)
    try {
      const dados = await perfilService.usarCapa(jogoId, token)
      setPerfil(dados)
      return dados
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Erro ao definir capa como foto'
      setError(msg)
      throw err
    } finally {
      setIsSalvando(false)
    }
  }

  async function removerFoto(): Promise<PerfilDTO> {
    setIsSalvando(true)
    setError(null)
    try {
      const dados = await perfilService.removerFoto(token)
      setPerfil(dados)
      return dados
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Erro ao remover foto'
      setError(msg)
      throw err
    } finally {
      setIsSalvando(false)
    }
  }

  function limparErro() {
    setError(null)
  }

  const value: PerfilStore = {
    perfil, isLoading, isSalvando, error,
    carregar, salvar, enviarFoto, usarCapa, removerFoto, limparErro,
  }

  return <PerfilContext.Provider value={value}>{children}</PerfilContext.Provider>
}

const defaultPerfilStore: PerfilStore = {
  perfil: null, isLoading: false, isSalvando: false, error: null,
  carregar: async () => null,
  salvar: async () => { throw new Error('PerfilProvider não encontrado') },
  enviarFoto: async () => { throw new Error('PerfilProvider não encontrado') },
  usarCapa: async () => { throw new Error('PerfilProvider não encontrado') },
  removerFoto: async () => { throw new Error('PerfilProvider não encontrado') },
  limparErro: () => {},
}

export function usePerfilStore(): PerfilStore {
  return useContext(PerfilContext) ?? defaultPerfilStore
}

