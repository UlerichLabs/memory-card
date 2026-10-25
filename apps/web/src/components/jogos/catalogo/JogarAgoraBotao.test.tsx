import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { JogarAgoraBotao } from './JogarAgoraBotao'

describe('JogarAgoraBotao', () => {
  it('renderiza o botao no estado padrao e responde ao clique', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()
    render(<JogarAgoraBotao onClick={onClick} isIniciando={false} emAndamento={false} />)

    const botao = screen.getByRole('button', { name: 'Jogar agora' })
    expect(botao).toBeEnabled()
    await user.click(botao)
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('fica desabilitado e exibe Iniciando... quando isIniciando e true', () => {
    render(<JogarAgoraBotao onClick={vi.fn()} isIniciando={true} emAndamento={false} />)

    const botao = screen.getByRole('button', { name: 'Iniciando...' })
    expect(botao).toBeDisabled()
  })

  it('fica desabilitado e exibe Em andamento quando emAndamento e true', () => {
    render(<JogarAgoraBotao onClick={vi.fn()} isIniciando={false} emAndamento={true} />)

    const botao = screen.getByRole('button', { name: 'Em andamento' })
    expect(botao).toBeDisabled()
  })
})
