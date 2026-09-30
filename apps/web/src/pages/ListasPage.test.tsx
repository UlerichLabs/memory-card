import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { ListasPage } from './ListasPage'
import { listasService } from '@/lib/services/listasService'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { ApiError } from '@/lib/api'
import type { ListaResumo, ListaDetalhada } from '@/types/listas'

const mockFila: ListaResumo = {
  id: 1,
  tipo: 'fila',
  nome: 'Fila Backlog',
  descricao: 'Jogos para jogar',
  origem: null,
  total_itens: 1,
  itens_pendentes: 1,
  progresso: null,
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-09-01T00:00:00Z',
}

const mockFilaDetalhe: ListaDetalhada = {
  ...mockFila,
  itens: [
    {
      id: 10,
      igdb_id: 100,
      nome: 'Chrono Trigger',
      console: 'SNES',
      igdb_capa_url: null,
      ano_lancamento: 1995,
      posicao: 1,
      origem: 'item',
      zerado: false,
      jogo_zerado: null,
    },
  ],
}

function renderListasPage(initialPath = '/listas') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <JogosProvider>
          <Routes>
            <Route path="/listas" element={<ListasPage />} />
            <Route path="/listas/:id" element={<ListasPage />} />
          </Routes>
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('ListasPage', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('exibe estado vazio quando não há nenhuma lista cadastrada', async () => {
    vi.spyOn(listasService, 'listar').mockResolvedValue([])

    renderListasPage('/listas')

    expect(await screen.findByText('Nenhuma lista ou desafio ainda')).toBeInTheDocument()
    expect(
      screen.getByText('Crie uma fila do que jogar em seguida ou um desafio como zerar uma franquia inteira.')
    ).toBeInTheDocument()
  })

  it('carrega e seleciona lista via rota /listas/:id', async () => {
    vi.spyOn(listasService, 'listar').mockResolvedValue([mockFila])
    vi.spyOn(listasService, 'obterPorId').mockResolvedValue(mockFilaDetalhe)

    renderListasPage('/listas/1')

    expect(await screen.findByRole('heading', { name: 'Fila Backlog' })).toBeInTheDocument()
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
  })

  it('exibe tela de Lista não encontrada quando o id não existe (404)', async () => {
    vi.spyOn(listasService, 'listar').mockResolvedValue([])
    vi.spyOn(listasService, 'obterPorId').mockRejectedValue(
      new ApiError('listas.nao_encontrada', 'Lista não encontrada', 404)
    )

    renderListasPage('/listas/999')

    expect(await screen.findByText('Lista não encontrada')).toBeInTheDocument()
    expect(screen.getByText('Voltar para Listas e Desafios')).toBeInTheDocument()
  })
})
