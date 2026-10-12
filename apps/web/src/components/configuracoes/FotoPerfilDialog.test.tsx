import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FotoPerfilDialog } from './FotoPerfilDialog'
import { PerfilContext, type PerfilStore } from '@/stores/perfilStore'
import { jogosService } from '@/lib/services/jogosService'
import type { PerfilDTO } from '@/types/perfil'

const mockPerfil: PerfilDTO = {
  id: 1,
  email: 'lucas@example.com',
  nome: 'Lucas Tester',
  username: 'lucas',
  bio: 'Gamer',
  avatar_url: null,
  foto_url: null,
  avatar_tipo: 'iniciais',
  jogo_favorito: null,
  console_favorito: null,
  jogando_desde: null,
  jogando_desde_ano: null,
  jogando_desde_efetivo: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}


function renderDialog(storeOverrides: Partial<PerfilStore> = {}, props: { open?: boolean; onOpenChange?: (open: boolean) => void; initialTab?: 'upload' | 'capa' } = {}) {
  const store: PerfilStore = {
    perfil: mockPerfil,
    isLoading: false,
    isSalvando: false,
    error: null,
    carregar: vi.fn().mockResolvedValue(mockPerfil),
    salvar: vi.fn().mockResolvedValue(mockPerfil),
    enviarFoto: vi.fn().mockResolvedValue(mockPerfil),
    usarCapa: vi.fn().mockResolvedValue(mockPerfil),
    removerFoto: vi.fn().mockResolvedValue(mockPerfil),
    limparErro: vi.fn(),
    ...storeOverrides,
  }

  const onOpenChange = props.onOpenChange ?? vi.fn()

  const utils = render(
    <PerfilContext.Provider value={store}>
      <FotoPerfilDialog open={props.open ?? true} onOpenChange={onOpenChange} initialTab={props.initialTab ?? 'upload'} />
    </PerfilContext.Provider>
  )

  return { ...utils, store, onOpenChange }
}

describe('FotoPerfilDialog', () => {
  beforeEach(() => {
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:mock-preview-url'),
      revokeObjectURL: vi.fn(),
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('renderiza com aba upload aberta por padrão e valida arquivo no cliente', async () => {
    const user = userEvent.setup()
    const { store } = renderDialog()

    expect(screen.getByRole('heading', { name: 'Foto de perfil' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Enviar foto' })).toHaveAttribute('aria-selected', 'true')

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    expect(input).toBeInTheDocument()

    const arquivoGrande = new File([new Uint8Array(3 * 1024 * 1024)], 'foto-grande.jpg', { type: 'image/jpeg' })
    await user.upload(input, arquivoGrande)

    expect(screen.getByRole('alert')).toHaveTextContent('A foto deve ter no máximo 2 MB.')
    expect(screen.getByRole('button', { name: 'Usar como foto' })).toBeDisabled()

    const dropZone = screen.getByLabelText('Selecionar imagem de perfil')
    const arquivoInvalido = new File(['conteudo'], 'documento.gif', { type: 'image/gif' })
    fireEvent.drop(dropZone, { dataTransfer: { files: [arquivoInvalido] } })

    expect(screen.getByRole('alert')).toHaveTextContent('Apenas JPG, PNG ou WebP.')

    const arquivoValido = new File(['conteudo-valido'], 'avatar.png', { type: 'image/png' })
    await user.upload(input, arquivoValido)

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    const btnSalvar = screen.getByRole('button', { name: 'Usar como foto' })
    expect(btnSalvar).toBeEnabled()

    await user.click(btnSalvar)
    expect(store.enviarFoto).toHaveBeenCalledWith(arquivoValido)
  })

  it('permite alternar para aba capa, buscar jogo e definir capa', async () => {
    const jogosMock = [
      { id: 10, usuario_id: 1, nome: 'Chrono Trigger', console: 'SNES', finalizado_em: '2026-03-01', tempo_jogado: 1200, nota: 10, dificuldade: 'A' as const, destaque: false, igdb_capa_url: 'https://images.igdb.com/chrono.jpg' },
      { id: 20, usuario_id: 1, nome: 'Super Mario World', console: 'SNES', finalizado_em: '2026-03-02', tempo_jogado: 600, nota: 9, dificuldade: 'B' as const, destaque: false, igdb_capa_url: '' },
    ]
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosMock,
      meta: { pagina: 1, por_pagina: 12, total: 2, total_paginas: 1 },
    })

    const user = userEvent.setup()
    const { store, onOpenChange } = renderDialog({}, { initialTab: 'capa' })

    expect(screen.getByRole('tab', { name: 'Capa de um jogo' })).toHaveAttribute('aria-selected', 'true')

    await waitFor(() => {
      expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
      expect(screen.getByText('Super Mario World')).toBeInTheDocument()
    })

    const marioBotao = screen.getByText('Super Mario World').closest('button')
    expect(marioBotao).toBeDisabled()

    const chronoBotao = screen.getByText('Chrono Trigger').closest('button')
    expect(chronoBotao).toBeEnabled()
    await user.click(chronoBotao!)

    expect(screen.getByText('Selecionado')).toBeInTheDocument()

    const btnSalvar = screen.getByRole('button', { name: 'Usar como foto' })
    await user.click(btnSalvar)

    expect(store.usarCapa).toHaveBeenCalledWith(10)
    await waitFor(() => {
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it('fecha o modal ao clicar em Cancelar', async () => {
    const user = userEvent.setup()
    const { onOpenChange } = renderDialog()

    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
