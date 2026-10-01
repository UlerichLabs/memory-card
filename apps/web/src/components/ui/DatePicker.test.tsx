import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { DatePicker } from './DatePicker'

describe('DatePicker', () => {
  it('renderiza com placeholder padrão dd/mm/aaaa quando vazio', () => {
    render(<DatePicker onChange={vi.fn()} ariaLabel="Data" />)
    const trigger = screen.getByRole('button', { name: 'Data' })
    expect(trigger).toHaveTextContent('dd/mm/aaaa')
  })

  it('exibe data formatada como dd/mm/aaaa quando valor fornecido em ISO', () => {
    render(<DatePicker value="2026-12-25" onChange={vi.fn()} ariaLabel="Data" />)
    const trigger = screen.getByRole('button', { name: 'Data' })
    expect(trigger).toHaveTextContent('25/12/2026')
  })

  it('abre o popover e permite selecionar uma data', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()
    render(<DatePicker onChange={handleChange} ariaLabel="Data" />)

    const trigger = screen.getByRole('button', { name: 'Data' })
    await user.click(trigger)

    const dayButton = await screen.findByRole('button', { name: /15/ })
    await user.click(dayButton)

    expect(handleChange).toHaveBeenCalled()
  })
})
