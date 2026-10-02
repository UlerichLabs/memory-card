import { afterEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GameForm } from './GameForm'
import { jogosService, JogosApiError, type IGDBJogoSugestao } from '@/lib/services/jogosService'

const sugestao: IGDBJogoSugestao = { id: 42, name: 'Chrono Trigger', first_release_date: 941500800, cover: { url: '//images.igdb.com/cover.jpg' }, platforms: [{ id: 7, name: 'PlayStation' }] }

function preencherObrigatorios() {
  fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '13/03/2026' } })
}

describe('GameForm', () => {
  afterEach(() => vi.restoreAllMocks())

  it('renderiza o redesign sem valor inicial para nota e dificuldade', () => {
    render(<GameForm onSubmit={vi.fn()} />)
    expect(screen.getByLabelText(/Nome do jogo/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Plataforma/i)).toBeInTheDocument()
    expect(screen.queryByText('Nenhum resumo disponível.')).not.toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'Nota' })).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'Dificuldade' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar registro' })).toBeInTheDocument()
  })

  it('usa a capa normalizada no preview quando o preenchimento inicial vem da IGDB', () => {
    render(<GameForm initialData={{ nome: 'Chrono Trigger', igdb_capa_url: '//images.igdb.com/t_thumb/co1abc.jpg' }} onSubmit={vi.fn()} />)

    expect(screen.getByAltText('Capa do jogo')).toHaveAttribute('src', 'https://images.igdb.com/t_cover_big/co1abc.jpg')
  })

  it('selecionar uma versão fecha a lista e não refaz a busca', async () => {
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([sugestao])
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue({ ...sugestao, summary: 'Resumo' })
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Chrono')
    await user.click(await screen.findByText('Chrono Trigger'))
    await waitFor(() => expect(jogosService.buscarIGDB).toHaveBeenCalledTimes(1))
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(screen.getByText('1999 · PlayStation')).toBeInTheDocument()
  })

  it('resposta atrasada de busca antiga não substitui a busca atual', async () => {
    let resolver: ((value: IGDBJogoSugestao[]) => void) | undefined
    vi.spyOn(jogosService, 'buscarIGDB').mockImplementation(() => new Promise((resolve) => { resolver = resolve }))
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Ch')
    await waitFor(() => expect(jogosService.buscarIGDB).toHaveBeenCalled())
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'rono')
    resolver?.([{ id: 1, name: 'Resultado antigo' }])
    await waitFor(() => expect(screen.queryByText('Resultado antigo')).not.toBeInTheDocument())
  })

  it('editar o nome depois da seleção limpa os dados do IGDB', async () => {
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([sugestao])
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue({ ...sugestao, summary: 'Resumo' })
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Chrono')
    await user.click(await screen.findByText('Chrono Trigger'))
    await waitFor(() => expect(screen.getByAltText('Capa do jogo')).toBeInTheDocument())
    await user.clear(screen.getByLabelText(/Nome do jogo/i))
    expect(screen.queryByAltText('Capa do jogo')).not.toBeInTheDocument()
    expect(screen.queryByText('1999 · PlayStation')).not.toBeInTheDocument()
    expect(screen.getByLabelText(/Plataforma/i)).toHaveValue('')
  })

  it('seleciona automaticamente uma plataforma e deixa vazio quando há várias', async () => {
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([sugestao])
    const detalhe = vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue({ ...sugestao, platforms: [{ id: 7, name: 'PlayStation' }] })
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Chrono')
    await user.click(await screen.findByText('Chrono Trigger'))
    await waitFor(() => expect(screen.getByLabelText(/Plataforma/i)).toHaveValue('PlayStation'))
    detalhe.mockResolvedValue({ ...sugestao, platforms: [{ id: 7, name: 'PlayStation' }, { id: 6, name: 'PC' }] })
    await user.clear(screen.getByLabelText(/Nome do jogo/i))
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Chrono')
    await user.click(await screen.findByText('Chrono Trigger'))
    await waitFor(() => expect(screen.getByLabelText(/Plataforma/i)).toHaveValue(''))
  })

  it('mantém gêneros inteiros dentro do limite ao preencher pelo IGDB', async () => {
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([sugestao])
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue({
      ...sugestao,
      genres: [{ id: 1, name: 'Ação' }, { id: 2, name: 'A'.repeat(150) }, { id: 3, name: 'RPG' }],
    })
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Chrono')
    await user.click(await screen.findByText('Chrono Trigger'))
    await waitFor(() => expect(screen.getByLabelText('Gênero')).toHaveValue('Ação, RPG'))
  })

  it('mascara data, valida data inválida e avança o tempo com Enter', async () => {
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    const data = screen.getByLabelText(/Finalizado em/i)
    await user.type(data, '130326')
    fireEvent.blur(data)
    expect(data).toHaveValue('13/03/2026')
    const horas = screen.getByLabelText('Horas')
    const minutos = screen.getByLabelText('Minutos')
    const segundos = screen.getByLabelText('Segundos')
    await user.type(horas, '13'); await user.keyboard('{Enter}')
    expect(minutos).toHaveFocus()
    await user.type(minutos, '2'); await user.keyboard('{Enter}')
    expect(segundos).toHaveFocus()
  })

  it('nota, dificuldade e review usam os controles novos', async () => {
    const user = userEvent.setup()
    render(<GameForm onSubmit={vi.fn()} />)
    await user.click(screen.getByRole('button', { name: '11' }))
    expect(screen.getByText('11 · Jogo da Vida')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Muito difícil' }))
    expect(screen.getByRole('button', { name: 'Muito difícil' })).toHaveAttribute('aria-pressed', 'true')
    await user.click(screen.getByRole('button', { name: '10' }))
    expect(screen.getByRole('button', { name: '10' })).not.toBeDisabled()
    expect(screen.getByRole('button', { name: '10' })).toHaveAttribute('aria-pressed', 'true')
    const review = screen.getByLabelText('Review')
    await user.type(review, 'Final difícil')
    expect(screen.getByText('13/5000')).toBeInTheDocument()
  })

  it('conflito 409 mostra mensagem junto à coroa', async () => {
    const erro = new JogosApiError('jogos.destaque_ano_conflito', 'Destaque já utilizado', 409)
    const onSubmit = vi.fn().mockRejectedValue(erro)
    const user = userEvent.setup()
    render(<GameForm initialData={{ finalizado_em: '2026-02-01' }} onSubmit={onSubmit} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Elden Ring')
    await user.type(screen.getByLabelText(/Plataforma/i), 'PC')
    preencherObrigatorios()
    await user.click(screen.getByRole('button', { name: '11' }))
    await user.click(screen.getByRole('button', { name: 'Normal' }))
    await user.click(screen.getByRole('button', { name: 'Marcar como jogo do ano' }))
    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))
    expect(onSubmit).toHaveBeenCalled()
    await waitFor(() => expect(screen.getByText('Destaque já utilizado')).toBeInTheDocument())
  })

  it('mapeia erro de review da API para o campo', async () => {
    const onSubmit = vi.fn().mockRejectedValue(new JogosApiError('jogos.review_muito_longo', 'erro interno', 400))
    const user = userEvent.setup()
    render(<GameForm initialData={{ finalizado_em: '2026-02-01' }} onSubmit={onSubmit} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Elden Ring')
    await user.type(screen.getByLabelText(/Plataforma/i), 'PC')
    preencherObrigatorios()
    await user.click(screen.getByRole('button', { name: '11' }))
    await user.click(screen.getByRole('button', { name: 'Normal' }))
    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))
    await waitFor(() => expect(screen.getByText('A review deve ter no máximo 5.000 caracteres.')).toBeInTheDocument())
  })

  it('usa mensagem genérica em português para erro desconhecido da API', async () => {
    const onSubmit = vi.fn().mockRejectedValue(new JogosApiError('erro.desconhecido', 'Internal server error', 500))
    const user = userEvent.setup()
    render(<GameForm initialData={{ finalizado_em: '2026-02-01' }} onSubmit={onSubmit} />)
    await user.type(screen.getByLabelText(/Nome do jogo/i), 'Elden Ring')
    await user.type(screen.getByLabelText(/Plataforma/i), 'PC')
    preencherObrigatorios()
    await user.click(screen.getByRole('button', { name: '11' }))
    await user.click(screen.getByRole('button', { name: 'Normal' }))
    await user.click(screen.getByRole('button', { name: 'Salvar registro' }))
    await waitFor(() => expect(screen.getByText('Não foi possível salvar o registro. Tente novamente.')).toBeInTheDocument())
    expect(screen.queryByText('Internal server error')).not.toBeInTheDocument()
  })
})
