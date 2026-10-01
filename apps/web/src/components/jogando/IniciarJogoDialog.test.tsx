import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { JogandoContext, type JogandoStore } from '@/stores/jogandoStore'
import { IniciarJogoDialog } from './IniciarJogoDialog'

vi.mock('@/components/jogos/GameForm/GameFormAutocomplete', () => ({
  GameFormAutocomplete: ({ onChangeNome, nome, error }: { onChangeNome: (value: string) => void; nome: string; error?: string }) => (
    <label>Jogo *<input aria-label="Jogo" value={nome} onChange={(event) => onChangeNome(event.target.value)} />{error && <span>{error}</span>}</label>
  ),
}))

function renderDialog(overrides: Partial<JogandoStore> = {}) {
  const store = {
    jogos: [], isLoading: false, carregado: true, error: null, aviso: null, isModalOpen: true,
    carregar: vi.fn(), criar: vi.fn().mockResolvedValue({}), remover: vi.fn(), abrirModalIniciar: vi.fn(), fecharModal: vi.fn(), definirAviso: vi.fn(), limparAviso: vi.fn(),
    ...overrides,
  } as unknown as JogandoStore
  return { store, ...render(<JogandoContext.Provider value={store}><IniciarJogoDialog /></JogandoContext.Provider>) }
}

describe('IniciarJogoDialog', () => {
  it('desabilita o envio sem nome e envia Hoje no payload', async () => {
    const { store } = renderDialog()
    expect(screen.getByRole('button', { name: 'Iniciar jogo' })).toBeDisabled()
    fireEvent.change(screen.getByRole('textbox', { name: 'Jogo' }), { target: { value: 'Hades' } })
    expect(screen.getByRole('button', { name: 'Iniciar jogo' })).toBeEnabled()
    fireEvent.click(screen.getByRole('button', { name: 'Iniciar jogo' }))
    await waitFor(() => expect(store.criar).toHaveBeenCalledWith(expect.objectContaining({ nome: 'Hades', iniciado_em: expect.any(String) })))
  })

  it('altera o início para Ontem e permite cancelar', async () => {
    const { store } = renderDialog()
    fireEvent.change(screen.getByRole('textbox', { name: 'Jogo' }), { target: { value: 'Celeste' } })
    fireEvent.click(screen.getByRole('button', { name: 'Ontem' }))
    fireEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(store.fecharModal).toHaveBeenCalled()
  })
})
