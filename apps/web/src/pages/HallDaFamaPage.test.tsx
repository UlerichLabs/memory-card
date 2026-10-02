import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { jogosService } from '@/lib/services/jogosService'
import { HallDaFamaPage } from './HallDaFamaPage'
import type { JogoZeradoDTO, ResumoGameDoAnoItem } from '@/types/jogos'

const jogoMock: JogoZeradoDTO = {
  id: 1,
  usuario_id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  finalizado_em: '2024-05-15T00:00:00Z',
  tempo_jogado: 36000,
  nota: 11,
  dificuldade: 'A',
  destaque: true,
}

const resumoMock: ResumoGameDoAnoItem[] = [
  { ano: 2024, total_jogos: 3, game_do_ano: jogoMock },
  { ano: 2023, total_jogos: 1, game_do_ano: null },
]

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/hall-da-fama']}>
      <AuthProvider>
        <JogosProvider>
          <HallDaFamaPage />
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('HallDaFamaPage', () => {
  beforeEach(() => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('atualiza o document.title para Hall da Fama', async () => {
    renderPage()

    await waitFor(() => {
      expect(document.title).toBe('Hall da Fama · Memory Card')
    })
  })

  it('exibe o estado vazio quando o usuário não tiver jogos zerados', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([])
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 100, total: 0, total_paginas: 0 },
    })

    renderPage()

    expect(await screen.findByText('Seu Hall da Fama ainda está vazio')).toBeInTheDocument()
    expect(
      screen.getByText('Registre jogos zerados para escolher seus Games do Ano.')
    ).toBeInTheDocument()
    expect(
      within(screen.getByRole('main')).getByRole('button', { name: '+ Registrar jogo' })
    ).toBeInTheDocument()
  })

  it('exibe erro de conexão e permite tentar novamente', async () => {
    const spyResumo = vi.spyOn(jogosService, 'obterResumoGameDoAno').mockRejectedValue(new Error('Erro de rede'))

    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByText('Erro ao carregar o Hall da Fama')).toBeInTheDocument()

    spyResumo.mockResolvedValue(resumoMock)
    const btnTentar = screen.getByRole('button', { name: 'Tentar novamente' })
    await user.click(btnTentar)

    expect(await screen.findByText('2024')).toBeInTheDocument()
  })

  it('renderiza os cards de Game do Ano e Games da Vida quando houver dados', async () => {
    renderPage()

    expect(await screen.findByText('2024')).toBeInTheDocument()
    expect(screen.getAllByText('Chrono Trigger').length).toBeGreaterThan(0)
    expect(screen.getByText('2023')).toBeInTheDocument()
    expect(screen.getByText('Sem Game do Ano')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Escolher Game do Ano' })).toBeInTheDocument()
    expect(screen.getByText('Games da Vida')).toBeInTheDocument()
  })

  it('abre o modal ao clicar em Escolher Game do Ano', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })

    const user = userEvent.setup()
    renderPage()

    const btnEscolher = await screen.findByRole('button', { name: 'Escolher Game do Ano' })
    await user.click(btnEscolher)

    expect(await screen.findByText('Game do Ano de 2023')).toBeInTheDocument()
  })

  it('abre o modal ao clicar em Trocar', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })

    const user = userEvent.setup()
    renderPage()

    const btnTrocar = await screen.findByRole('button', { name: 'Trocar' })
    await user.click(btnTrocar)

    expect(await screen.findByText('Game do Ano de 2024')).toBeInTheDocument()
  })

  it('exibe estado de nenhum Game da Vida quando não houver jogos com nota 11', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 100, total: 0, total_paginas: 0 },
    })

    renderPage()

    expect(await screen.findByText('Nenhum Game da Vida ainda')).toBeInTheDocument()
    expect(
      screen.getByText('Quando um jogo marcar a sua vida, dê nota 11 a ele.')
    ).toBeInTheDocument()
  })

  it('modal Trocar: atualiza card do ano e selos de Games da Vida após confirmação', async () => {
    const jogoA: JogoZeradoDTO = {
      id: 1,
      usuario_id: 1,
      nome: 'Chrono Trigger',
      console: 'SNES',
      finalizado_em: '2024-05-15T00:00:00Z',
      tempo_jogado: 36000,
      nota: 11,
      dificuldade: 'A',
      destaque: true,
    }
    const jogoB: JogoZeradoDTO = {
      id: 2,
      usuario_id: 1,
      nome: 'Super Mario 64',
      console: 'N64',
      finalizado_em: '2024-06-20T00:00:00Z',
      tempo_jogado: 40000,
      nota: 10,
      dificuldade: 'B',
      destaque: false,
    }

    const spyResumo = vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([
      { ano: 2024, total_jogos: 2, game_do_ano: jogoA },
    ])

    const spyListar = vi.spyOn(jogosService, 'listar').mockImplementation((params) => {
      if (params?.ano === 2024) {
        return Promise.resolve({
          data: [jogoA, jogoB],
          meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
        })
      }
      return Promise.resolve({
        data: [jogoA],
        meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
      })
    })

    vi.spyOn(jogosService, 'definirGameDoAno').mockResolvedValue({
      ano: 2024,
      anterior_id: 1,
      game_do_ano: { ...jogoB, destaque: true },
    })

    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByText('Jogo do ano 2024')).toBeInTheDocument()

    const btnTrocar = screen.getByRole('button', { name: 'Trocar' })
    await user.click(btnTrocar)

    expect(await screen.findByText('Game do Ano de 2024')).toBeInTheDocument()

    spyResumo.mockResolvedValue([
      { ano: 2024, total_jogos: 2, game_do_ano: { ...jogoB, destaque: true } },
    ])
    spyListar.mockImplementation((params) => {
      if (params?.ano === 2024) {
        return Promise.resolve({
          data: [{ ...jogoB, destaque: true }, { ...jogoA, destaque: false }],
          meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
        })
      }
      return Promise.resolve({
        data: [{ ...jogoA, destaque: false }],
        meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
      })
    })

    const radioMario = await screen.findByText('Super Mario 64')
    await user.click(radioMario)

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    await user.click(btnConfirmar)

    await waitFor(() => {
      expect(screen.queryByText('Game do Ano de 2024')).not.toBeInTheDocument()
    })

    expect(await screen.findByText('Super Mario 64')).toBeInTheDocument()
    expect(screen.queryByText('Jogo do ano 2024')).not.toBeInTheDocument()
  })
})
