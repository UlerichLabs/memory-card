import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HallDaFamaProvider } from '@/stores/hallDaFamaStore'
import { jogosService, JogosApiError } from '@/lib/services/jogosService'
import { EscolherGameDoAnoDialog } from './EscolherGameDoAnoDialog'
import type { JogoZeradoDTO } from '@/types/jogos'

const jogosAnoMock: JogoZeradoDTO[] = [
  {
    id: 10,
    usuario_id: 1,
    nome: 'Zelda Ocarina of Time',
    console: 'N64',
    finalizado_em: '2024-05-15T00:00:00Z',
    tempo_jogado: 120000,
    nota: 11,
    dificuldade: 'AA',
    destaque: true,
  },
  {
    id: 20,
    usuario_id: 1,
    nome: 'Super Mario 64',
    console: 'N64',
    finalizado_em: '2024-02-10T00:00:00Z',
    tempo_jogado: 40000,
    nota: 10,
    dificuldade: 'A',
    destaque: false,
  },
]

function renderDialog(props: {
  open: boolean
  ano: number
  jogoAtual: JogoZeradoDTO | null
  onOpenChange?: (open: boolean) => void
  onSuccess?: () => void
}) {
  return render(
    <HallDaFamaProvider>
      <EscolherGameDoAnoDialog
        open={props.open}
        onOpenChange={props.onOpenChange ?? vi.fn()}
        ano={props.ano}
        jogoAtual={props.jogoAtual}
        onSuccess={props.onSuccess}
      />
    </HallDaFamaProvider>
  )
}

describe('EscolherGameDoAnoDialog', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('no modo Escolher: lista jogos ordenados por nota, sem seleção inicial e botão desabilitado', async () => {
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosAnoMock,
      meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
    })

    renderDialog({ open: true, ano: 2024, jogoAtual: null })

    expect(await screen.findByText('Game do Ano de 2024')).toBeInTheDocument()
    expect(spyListar).toHaveBeenCalledWith({ ano: 2024, ordenar: 'nota', por_pagina: 100 }, undefined, expect.any(AbortSignal))

    expect(await screen.findByText('Zelda Ocarina of Time')).toBeInTheDocument()
    expect(screen.getByText('Super Mario 64')).toBeInTheDocument()

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    expect(btnConfirmar).toBeDisabled()
    expect(screen.queryByText(/deixa de ser o game do ano/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /remover game do ano/i })).not.toBeInTheDocument()
  })

  it('no modo Escolher: selecionar um jogo habilita botão e confirmar chama definirGameDoAno', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosAnoMock,
      meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
    })
    const spyDefinir = vi.spyOn(jogosService, 'definirGameDoAno').mockResolvedValue({
      ano: 2024,
      anterior_id: null,
      game_do_ano: jogosAnoMock[1],
    })
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([])

    const onOpenChange = vi.fn()
    const onSuccess = vi.fn()
    const user = userEvent.setup()

    renderDialog({ open: true, ano: 2024, jogoAtual: null, onOpenChange, onSuccess })

    const radioMario = await screen.findByText('Super Mario 64')
    await user.click(radioMario)

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    expect(btnConfirmar).toBeEnabled()

    await user.click(btnConfirmar)

    await waitFor(() => {
      expect(spyDefinir).toHaveBeenCalledWith(20, undefined)
      expect(onSuccess).toHaveBeenCalled()
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it('no modo Trocar: abre com o atual selecionado, avisa ao trocar e botão Remover Game do Ano chama DELETE', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosAnoMock,
      meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
    })
    const spyRemover = vi.spyOn(jogosService, 'removerGameDoAno').mockResolvedValue(undefined)
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([])

    const onOpenChange = vi.fn()
    const user = userEvent.setup()

    renderDialog({ open: true, ano: 2024, jogoAtual: jogosAnoMock[0], onOpenChange })

    expect(await screen.findByText('Zelda Ocarina of Time')).toBeInTheDocument()

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    expect(btnConfirmar).toBeDisabled()

    const btnRemover = screen.getByRole('button', { name: /remover game do ano/i })
    expect(btnRemover).toBeInTheDocument()

    const radioMario = screen.getByText('Super Mario 64')
    await user.click(radioMario)

    expect(
      screen.getByText((content) => content.includes('deixa de ser o Game do Ano de 2024'))
    ).toBeInTheDocument()
    expect(btnConfirmar).toBeEnabled()

    await user.click(btnRemover)

    await waitFor(() => {
      expect(spyRemover).toHaveBeenCalledWith(10, undefined)
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it.each([
    {
      tipoErro: '409 conflito',
      erro: new JogosApiError('jogos.destaque_ano_conflito', 'conflito', 409),
      mensagemEsperada: 'Este ano já possui outro Game do Ano definido.',
    },
    {
      tipoErro: 'erro genérico',
      erro: new Error('Erro inesperado de rede'),
      mensagemEsperada: 'Não foi possível definir o Game do Ano. Tente novamente.',
    },
  ])('exibe mensagem amigável e mantém modal aberto em $tipoErro', async ({ erro, mensagemEsperada }) => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosAnoMock,
      meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
    })
    vi.spyOn(jogosService, 'definirGameDoAno').mockRejectedValue(erro)

    const onOpenChange = vi.fn()
    const user = userEvent.setup()
    renderDialog({ open: true, ano: 2024, jogoAtual: null, onOpenChange })

    const radioMario = await screen.findByText('Super Mario 64')
    await user.click(radioMario)

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    await user.click(btnConfirmar)

    expect(await screen.findByText(mensagemEsperada)).toBeInTheDocument()
    expect(screen.getByText('Game do Ano de 2024')).toBeInTheDocument()
    expect(onOpenChange).not.toHaveBeenCalled()
  })

  it('permite navegação por teclado com setas no radiogroup e seleção por Espaço e confirmação por Enter', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: jogosAnoMock,
      meta: { pagina: 1, por_pagina: 100, total: 2, total_paginas: 1 },
    })
    const spyDefinir = vi.spyOn(jogosService, 'definirGameDoAno').mockResolvedValue({
      ano: 2024,
      anterior_id: null,
      game_do_ano: jogosAnoMock[1],
    })
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([])

    const onOpenChange = vi.fn()
    const onSuccess = vi.fn()
    const user = userEvent.setup()
    renderDialog({ open: true, ano: 2024, jogoAtual: null, onOpenChange, onSuccess })

    const radioZelda = await screen.findByRole('radio', { name: /zelda ocarina of time/i })
    radioZelda.focus()

    await user.keyboard('{ArrowDown}')

    const radioMario = screen.getByRole('radio', { name: /super mario 64/i })
    expect(radioMario).toHaveAttribute('aria-checked', 'true')

    await user.keyboard('{ArrowUp}')
    expect(radioZelda).toHaveAttribute('aria-checked', 'true')

    radioMario.focus()
    await user.keyboard(' ')
    expect(radioMario).toHaveAttribute('aria-checked', 'true')

    const btnConfirmar = screen.getByRole('button', { name: /definir como game do ano/i })
    btnConfirmar.focus()
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(spyDefinir).toHaveBeenCalledWith(20, undefined)
      expect(onSuccess).toHaveBeenCalled()
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })
})
