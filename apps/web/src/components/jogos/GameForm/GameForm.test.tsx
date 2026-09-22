import { afterEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthContext } from '@/store/authStore'
import { GameForm } from './GameForm'
import { jogosService, JogosApiError, type IGDBJogoSugestao } from '@/lib/services/jogosService'

describe('GameForm', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renderiza todos os campos obrigatórios', () => {
    render(<GameForm onSubmit={vi.fn()} />)

    expect(screen.getByLabelText(/Nome do jogo/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Console/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Finalizado em/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Horas/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Minutos/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Segundos/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Nota \(1 a 11\)/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Dificuldade/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar registro' })).toBeInTheDocument()
  })

  it('autocomplete IGDB busca após debounce e preenche os campos ao selecionar', async () => {
    const sugestoes: IGDBJogoSugestao[] = [
      {
        id: 42,
        name: 'Chrono Trigger',
        cover: { id: 1, url: '//images.igdb.com/cover.jpg' },
        summary: 'Um RPG clássico sobre viagens no tempo.',
      },
    ]
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue(sugestoes)

    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)

    const nomeInput = screen.getByLabelText(/Nome do jogo/i)
    await user.type(nomeInput, 'Chrono')

    await waitFor(() => {
      expect(jogosService.buscarIGDB).toHaveBeenCalledWith('Chrono', undefined, expect.any(AbortSignal))
    })

    const opcao = await screen.findByText('Chrono Trigger')
    expect(opcao).toBeInTheDocument()

    await user.click(opcao)

    expect(nomeInput).toHaveValue('Chrono Trigger')
    const capa = screen.getByAltText('Capa do jogo')
    expect(capa).toHaveAttribute('src', 'https://images.igdb.com/cover.jpg')
  })

  it('autocomplete envia token de autenticacao quando usuario autenticado', async () => {
    const sugestoes: IGDBJogoSugestao[] = [
      { id: 10, name: 'God of War', cover: { url: '//images.igdb.com/gow.jpg' } },
    ]
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue(sugestoes)

    const mockAuthValue = {
      sessao: {
        access_token: 'token-jwt-valido',
        refresh_token: 'r-token',
        usuario: { id: 1, nome: 'Test', email: 't@example.com', idioma: 'pt-BR', created_at: '' },
      },
      login: vi.fn(),
      request: vi.fn(),
      refresh: vi.fn(),
      logout: vi.fn(),
    }

    const user = userEvent.setup()
    render(
      <AuthContext.Provider value={mockAuthValue}>
        <GameForm onSubmit={vi.fn()} />
      </AuthContext.Provider>
    )

    const nomeInput = screen.getByLabelText(/Nome do jogo/i)
    await user.type(nomeInput, 'God')

    await waitFor(() => {
      expect(jogosService.buscarIGDB).toHaveBeenCalledWith('God', 'token-jwt-valido', expect.any(AbortSignal))
    })
  })

  it('detalhes do jogo preenchem plataformas sugeridas e genero', async () => {
    const sugestoes: IGDBJogoSugestao[] = [
      { id: 99, name: 'God of War Ragnarok' },
    ]
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue(sugestoes)
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue({
      id: 99,
      name: 'God of War Ragnarok',
      genres: [{ id: 1, name: 'Ação' }, { id: 2, name: 'Aventura' }],
      platforms: [{ id: 10, name: 'PlayStation 5' }, { id: 11, name: 'PlayStation 4' }],
      summary: 'Jornada mitológica nórdica.',
    })

    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)

    const nomeInput = screen.getByLabelText(/Nome do jogo/i)
    await user.type(nomeInput, 'God')
    const opcao = await screen.findByText('God of War Ragnarok')
    await user.click(opcao)

    await waitFor(() => {
      expect(screen.getByLabelText(/Gênero/i)).toHaveValue('Ação, Aventura')
      expect(screen.getByLabelText(/Console/i)).toHaveValue('PlayStation 5')
    })

    await user.click(screen.getByLabelText(/Console/i))
    const opcaoPs4 = screen.getByRole('option', { name: 'PlayStation 4' })
    await user.click(opcaoPs4)
    expect(screen.getByLabelText(/Console/i)).toHaveValue('PlayStation 4')
  })

  it('campos preenchidos via IGDB continuam editáveis manualmente', async () => {
    const sugestoes: IGDBJogoSugestao[] = [
      {
        id: 42,
        name: 'Chrono Trigger',
        cover: { id: 1, url: 'https://images.igdb.com/cover.jpg' },
        summary: 'Descrição',
      },
    ]
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue(sugestoes)

    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)

    const nomeInput = screen.getByLabelText(/Nome do jogo/i)
    await user.type(nomeInput, 'Chrono')

    const opcao = await screen.findByText('Chrono Trigger')
    await user.click(opcao)

    expect(nomeInput).toHaveValue('Chrono Trigger')

    await user.type(nomeInput, ' - Edição SNES')
    expect(nomeInput).toHaveValue('Chrono Trigger - Edição SNES')
  })

  it('validação: nota fora de 1-11 exibe erro', async () => {
    const user = userEvent.setup()
    const handleSubmit = vi.fn()
    render(<GameForm onSubmit={handleSubmit} />)

    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Super Mario')
    await user.click(screen.getByLabelText(/Console/i))
    await user.click(screen.getByRole('option', { name: 'Super Nintendo' }))
    fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '01/02/2026' } })

    const notaInput = screen.getByLabelText(/Nota \(1 a 11\)/i)
    await user.clear(notaInput)
    await user.type(notaInput, '15')

    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))

    expect(handleSubmit).not.toHaveBeenCalled()
    expect(screen.getByText('Nota deve ser entre 1 e 11')).toBeInTheDocument()
  })

  it('validação: minutos/segundos > 59 exibe erro', async () => {
    const user = userEvent.setup()
    const handleSubmit = vi.fn()
    render(<GameForm onSubmit={handleSubmit} />)

    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Super Mario')
    await user.click(screen.getByLabelText(/Console/i))
    await user.click(screen.getByRole('option', { name: 'Super Nintendo' }))
    fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '01/02/2026' } })

    const minutosInput = screen.getByLabelText(/Minutos/i)
    await user.clear(minutosInput)
    await user.type(minutosInput, '75')

    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))

    expect(handleSubmit).not.toHaveBeenCalled()
    expect(screen.getByText('Minutos devem ser entre 0 e 59')).toBeInTheDocument()
  })

  it('contador de caracteres atualiza e bloqueia além de 500', async () => {
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)

    const textarea = screen.getByPlaceholderText(/Ex: 100% de conquistas/i)
    expect(screen.getByText('0/500')).toBeInTheDocument()
    expect(textarea).toHaveAttribute('maxLength', '500')

    await user.type(textarea, 'Zerado com 100%')
    expect(screen.getByText('15/500')).toBeInTheDocument()
  })

  it('conflito 409 no destaque exibe mensagem e reverte o toggle', async () => {
    const user = userEvent.setup()
    const erro409 = new JogosApiError('jogos.destaque_ano_conflito', 'já existe um destaque para este ano', 409)
    const handleSubmit = vi.fn().mockRejectedValue(erro409)

    render(<GameForm onSubmit={handleSubmit} />)

    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Elden Ring')
    await user.click(screen.getByLabelText(/Console/i))
    await user.click(screen.getByRole('option', { name: 'PC' }))
    fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '01/02/2026' } })

    const destaqueSwitch = screen.getByRole('switch', { name: /Marcar como jogo destaque do ano/i })
    await user.click(destaqueSwitch)
    expect(destaqueSwitch).toBeChecked()

    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))

    await waitFor(() => {
      expect(screen.getByText('já existe um destaque para este ano')).toBeInTheDocument()
    })
    expect(destaqueSwitch).not.toBeChecked()
  })
})
