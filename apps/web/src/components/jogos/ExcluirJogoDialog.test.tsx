import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ExcluirJogoDialog } from './ExcluirJogoDialog'

describe('ExcluirJogoDialog', () => {
  it('renderiza título, aviso de irreversibilidade e nome do jogo quando aberto', () => {
    render(
      <ExcluirJogoDialog
        open={true}
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        jogoNome="Super Metroid"
      />
    )

    expect(screen.getByRole('heading', { name: 'Excluir registro' })).toBeInTheDocument()
    expect(screen.getByText(/Tem certeza que deseja excluir o registro de "Super Metroid"\?/)).toBeInTheDocument()
    expect(screen.getByText(/Esta ação não pode ser desfeita\./)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Excluir' })).toBeInTheDocument()
  })

  it('não chama onConfirm ao clicar em Cancelar e fecha o modal', async () => {
    const user = userEvent.setup()
    const handleOpenChange = vi.fn()
    const handleConfirm = vi.fn()

    render(
      <ExcluirJogoDialog
        open={true}
        onOpenChange={handleOpenChange}
        onConfirm={handleConfirm}
        jogoNome="Super Metroid"
      />
    )

    await user.click(screen.getByRole('button', { name: 'Cancelar' }))

    expect(handleConfirm).not.toHaveBeenCalled()
    expect(handleOpenChange).toHaveBeenCalledWith(false)
  })

  it('chama onConfirm ao clicar em Excluir', async () => {
    const user = userEvent.setup()
    const handleOpenChange = vi.fn()
    const handleConfirm = vi.fn()

    render(
      <ExcluirJogoDialog
        open={true}
        onOpenChange={handleOpenChange}
        onConfirm={handleConfirm}
        jogoNome="Super Metroid"
      />
    )

    await user.click(screen.getByRole('button', { name: 'Excluir' }))

    expect(handleConfirm).toHaveBeenCalledTimes(1)
  })

  it('desabilita botões e exibe estado de loading durante exclusão', () => {
    render(
      <ExcluirJogoDialog
        open={true}
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        jogoNome="Super Metroid"
        isLoading={true}
      />
    )

    expect(screen.getByRole('button', { name: 'Excluindo...' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeDisabled()
  })
})
