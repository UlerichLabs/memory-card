import { describe, expect, it, vi, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PerfilSecao } from './PerfilSecao'
import { PerfilContext, type PerfilStore } from '@/stores/perfilStore'
import type { PerfilDTO } from '@/types/perfil'

const mockPerfil: PerfilDTO = {
  id: 1,
  email: 'lucas@example.com',
  nome: 'Lucas Tester',
  username: 'lucas_tester',
  bio: 'Colecionador de cartuchos',
  avatar_url: 'https://cdn.example.com/foto.jpg',
  foto_url: 'https://cdn.example.com/foto.jpg',
  avatar_tipo: 'upload',
  jogo_favorito: {
    id: 15,
    nome: 'Super Metroid',
    capa_url: 'https://cdn.example.com/metroid.jpg',
  },
  console_favorito: 'Super Nintendo',
  jogando_desde: 1994,
  jogando_desde_ano: 1994,
  jogando_desde_efetivo: 1994,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}


function renderPerfilSecao(storeOverrides: Partial<PerfilStore> = {}) {
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

  const utils = render(
    <PerfilContext.Provider value={store}>
      <PerfilSecao />
    </PerfilContext.Provider>
  )

  return { ...utils, store }
}

describe('PerfilSecao', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renderiza os cards e preenche os campos com os dados do perfil', () => {
    renderPerfilSecao()

    expect(screen.getByRole('heading', { name: 'Foto de perfil' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Informações' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Em destaque no perfil' })).toBeInTheDocument()

    expect(screen.getByLabelText(/Nome de exibição/)).toHaveValue('Lucas Tester')
    expect(screen.getByLabelText(/Nome de usuário/)).toHaveValue('lucas_tester')
    expect(screen.getByLabelText(/Sobre você/)).toHaveValue('Colecionador de cartuchos')
    expect(screen.getByText('Super Metroid')).toBeInTheDocument()
    expect(screen.getByLabelText(/Jogando desde/)).toHaveValue(1994)

    expect(screen.getByRole('button', { name: 'Descartar' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Salvar alterações' })).toBeDisabled()
  })

  it('habilita botões ao alterar dados e restaura ao clicar em Descartar', async () => {
    const user = userEvent.setup()
    renderPerfilSecao()

    const nomeInput = screen.getByLabelText(/Nome de exibição/)
    await user.clear(nomeInput)
    await user.type(nomeInput, 'Novo Nome')

    const btnDescartar = screen.getByRole('button', { name: 'Descartar' })
    const btnSalvar = screen.getByRole('button', { name: 'Salvar alterações' })
    expect(btnDescartar).toBeEnabled()
    expect(btnSalvar).toBeEnabled()

    await user.click(btnDescartar)
    expect(nomeInput).toHaveValue('Lucas Tester')
    expect(btnDescartar).toBeDisabled()
    expect(btnSalvar).toBeDisabled()
  })

  it('exibe erro de validação e bloqueia envio para ano inválido', async () => {
    const user = userEvent.setup()
    const { store } = renderPerfilSecao()

    const anoInput = screen.getByLabelText(/Jogando desde/)
    await user.clear(anoInput)
    await user.type(anoInput, '1960')

    await user.click(screen.getByRole('button', { name: 'Salvar alterações' }))

    expect(await screen.findByText(/O ano deve ser entre 1970/)).toBeInTheDocument()
    expect(store.salvar).not.toHaveBeenCalled()
  })

  it('chama salvar com os dados corretos ao submeter formulário válido', async () => {
    const user = userEvent.setup()
    const { store } = renderPerfilSecao()

    const bioInput = screen.getByLabelText(/Sobre você/)
    await user.clear(bioInput)
    await user.type(bioInput, 'Nova bio atualizada')

    await user.click(screen.getByRole('button', { name: 'Salvar alterações' }))

    await waitFor(() => {
      expect(store.salvar).toHaveBeenCalledWith({
        nome: 'Lucas Tester',
        username: 'lucas_tester',
        bio: 'Nova bio atualizada',
        jogo_favorito_id: 15,
        console_favorito: 'Super Nintendo',
        jogando_desde: 1994,
      })
    })
  })

  it('chama removerFoto ao clicar no botão Remover do card de foto', async () => {
    const user = userEvent.setup()
    const { store } = renderPerfilSecao()

    const btnRemover = screen.getByRole('button', { name: 'Remover' })
    expect(btnRemover).toBeInTheDocument()

    await user.click(btnRemover)
    expect(store.removerFoto).toHaveBeenCalled()
  })

  it('renderiza indicador de carregamento quando isLoading e perfil nulo', () => {
    renderPerfilSecao({ isLoading: true, perfil: null })

    expect(screen.getByRole('status', { name: 'Carregando perfil' })).toBeInTheDocument()
  })
})
