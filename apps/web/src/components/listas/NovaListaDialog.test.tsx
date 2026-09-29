import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { NovaListaDialog } from './NovaListaDialog'
import { ListasContext, type ListasStore } from '@/stores/listasStore'
import { listasService } from '@/lib/services/listasService'
import { ApiError } from '@/lib/api'
import type { ListaResumo } from '@/types/listas'

function renderNovaListaDialog(storeOverrides: Partial<ListasStore> = {}) {
  const store: ListasStore = {
    listas: [],
    listaAberta: null,
    isLoading: false,
    isLoadingDetalhe: false,
    error: null,
    filtroAba: 'todos',
    isNovaListaOpen: true,
    listaEmEdicao: null,
    isExcluirListaOpen: false,
    listaParaExcluir: null,
    setFiltroAba: vi.fn(),
    limparErro: vi.fn(),
    abrirModalCriar: vi.fn(),
    abrirModalEditar: vi.fn(),
    fecharModalNovaLista: vi.fn(),
    abrirModalExcluir: vi.fn(),
    fecharModalExcluir: vi.fn(),
    carregarListas: vi.fn(),
    abrirLista: vi.fn(),
    criarLista: vi.fn().mockResolvedValue({ id: 99, nome: 'Teste' }),
    atualizarLista: vi.fn().mockResolvedValue({ id: 99, nome: 'Teste' }),
    excluirLista: vi.fn(),
    adicionarItem: vi.fn(),
    removerItem: vi.fn(),
    reordenarItens: vi.fn(),
    associarZeramento: vi.fn(),
    sincronizarFranquia: vi.fn(),
    ...storeOverrides,
  }

  return {
    ...render(
      <MemoryRouter>
        <ListasContext.Provider value={store}>
          <NovaListaDialog />
        </ListasContext.Provider>
      </MemoryRouter>
    ),
    store,
  }
}

describe('NovaListaDialog', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renderiza o formulário no modo padrão (Fila)', () => {
    renderNovaListaDialog()
    expect(screen.getByRole('heading', { name: 'Nova lista ou desafio' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: /Fila/i })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('button', { name: 'Criar fila' })).toBeInTheDocument()
  })

  it('alterna entre Fila e Desafio ao clicar no card', async () => {
    const user = userEvent.setup()
    renderNovaListaDialog()

    const btnDesafio = screen.getByRole('radio', { name: /Desafio/i })
    await user.click(btnDesafio)

    expect(btnDesafio).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByText('Quais jogos contam?')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Criar desafio' })).toBeInTheDocument()
  })

  it('cada regra mostra o campo certo e o texto de ajuda certo', async () => {
    const user = userEvent.setup()
    renderNovaListaDialog()

    await user.click(screen.getByRole('radio', { name: /Desafio/i }))

    expect(screen.getByLabelText(/Franquia \*/i)).toBeInTheDocument()
    expect(screen.getByText(/Os jogos da franquia são trazidos do IGDB/i)).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: /Plataforma/i }))
    expect(screen.getByLabelText(/Plataforma \*/i)).toBeInTheDocument()
    expect(screen.getByText(/Não tem lista fixa/i)).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: /Gênero/i }))
    expect(screen.getByLabelText(/Gênero \*/i)).toBeInTheDocument()
    expect(screen.getByText(/Todo zeramento com esse gênero conta/i)).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: /Eu escolho os jogos/i }))
    expect(screen.queryByLabelText(/Plataforma \*/i)).not.toBeInTheDocument()
    expect(screen.getByText(/Você adiciona os jogos depois de criar/i)).toBeInTheDocument()
  })

  it('exibe erros Zod por campo ao tentar submeter inválido', async () => {
    const user = userEvent.setup()
    renderNovaListaDialog()

    await user.click(screen.getByRole('button', { name: 'Criar fila' }))
    expect(await screen.findByText('O nome da lista é obrigatório.')).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: /Desafio/i }))
    await user.type(screen.getByLabelText(/Nome \*/i), 'Desafio Teste')

    await user.click(screen.getByRole('button', { name: 'Criar desafio' }))
    expect(await screen.findByText('Selecione uma franquia da lista.')).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: /Plataforma/i }))
    await user.click(screen.getByRole('button', { name: 'Criar desafio' }))
    expect(await screen.findByText('Informe a plataforma.')).toBeInTheDocument()
    expect(await screen.findByText('A meta é obrigatória para este desafio.')).toBeInTheDocument()
  })

  it('envia payload correto para criação de fila', async () => {
    const user = userEvent.setup()
    const { store } = renderNovaListaDialog()

    await user.type(screen.getByLabelText(/Nome \*/i), 'Minha Fila')
    await user.type(screen.getByLabelText(/Descrição \(opcional\)/i), 'Uma descrição legal')
    await user.click(screen.getByRole('button', { name: 'Criar fila' }))

    expect(store.criarLista).toHaveBeenCalledWith({
      tipo: 'fila',
      nome: 'Minha Fila',
      descricao: 'Uma descrição legal',
      regra: null,
      meta: null,
    })
  })

  it('envia payload correto para criação de desafio de franquia', async () => {
    vi.spyOn(listasService, 'buscarFranquiasIGDB').mockResolvedValue([{ id: 10, name: 'Mario' }])
    const user = userEvent.setup()
    const { store } = renderNovaListaDialog()

    await user.click(screen.getByRole('radio', { name: /Desafio/i }))
    await user.type(screen.getByLabelText(/Nome \*/i), 'Desafio Mario')

    const inputFranquia = screen.getByLabelText(/Franquia \*/i)
    await user.type(inputFranquia, 'Mario')

    const option = await screen.findByRole('option', { name: 'Mario' })
    await user.click(option)

    await user.click(screen.getByRole('button', { name: 'Criar desafio' }))

    expect(store.criarLista).toHaveBeenCalledWith({
      tipo: 'desafio',
      nome: 'Desafio Mario',
      descricao: null,
      regra: { tipo: 'franquia', valor: 'Mario', igdb_id: 10 },
      meta: null,
    })
  })

  it('envia payload correto para desafio de plataforma', async () => {
    const user = userEvent.setup()
    const { store } = renderNovaListaDialog()

    await user.click(screen.getByRole('radio', { name: /Desafio/i }))
    await user.click(screen.getByRole('radio', { name: /Plataforma/i }))
    await user.type(screen.getByLabelText(/Nome \*/i), 'Desafio SNES')
    await user.type(screen.getByLabelText(/Plataforma \*/i), 'SNES')
    await user.type(screen.getByLabelText(/Meta \*/i), '10')

    await user.click(screen.getByRole('button', { name: 'Criar desafio' }))

    expect(store.criarLista).toHaveBeenCalledWith({
      tipo: 'desafio',
      nome: 'Desafio SNES',
      descricao: null,
      regra: { tipo: 'plataforma', valor: 'SNES', igdb_id: null },
      meta: 10,
    })
  })

  it('exibe erro 400 da API mapeado para o campo correspondente', async () => {
    const user = userEvent.setup()
    const criarMock = vi.fn().mockRejectedValue(
      new ApiError('listas.nome_invalido', 'Nome inválido', 400)
    )
    renderNovaListaDialog({ criarLista: criarMock })

    await user.type(screen.getByLabelText(/Nome \*/i), 'Fila')
    await user.click(screen.getByRole('button', { name: 'Criar fila' }))

    expect(await screen.findByText('Nome inválido.')).toBeInTheDocument()
  })

  it('no modo edição, exibe título Editar e bloqueia tipo e regras', () => {
    const listaEdicao: ListaResumo = {
      id: 5,
      tipo: 'desafio',
      nome: 'Zelda Franquia',
      descricao: 'Desc',
      regra: { tipo: 'franquia', valor: 'Zelda', igdb_id: 10 },
      meta: 5,
      total_itens: 2,
      itens_pendentes: 1,
      progresso: null,
      created_at: '2026-09-01T00:00:00Z',
      updated_at: '2026-09-01T00:00:00Z',
    }

    renderNovaListaDialog({ listaEmEdicao: listaEdicao })

    expect(screen.getByRole('heading', { name: 'Editar desafio' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: /Fila/i })).toBeDisabled()
    expect(screen.getByRole('radio', { name: /Desafio/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Salvar' })).toBeInTheDocument()
  })
})
