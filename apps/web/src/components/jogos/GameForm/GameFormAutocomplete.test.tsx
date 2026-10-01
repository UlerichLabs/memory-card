import { afterEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { useState } from 'react'
import { jogosService } from '@/lib/services/jogosService'
import { GameFormAutocomplete } from './GameFormAutocomplete'

describe('GameFormAutocomplete', () => {
  afterEach(() => vi.restoreAllMocks())

  it('não busca nem abre sugestões com nome pré-preenchido', async () => {
    const buscar = vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([{ id: 1, name: 'Super Metroid' }])
    render(<GameFormAutocomplete nome="Super Metroid" onChangeNome={vi.fn()} onSelectSugestao={vi.fn()} />)
    await new Promise((resolve) => setTimeout(resolve, 400))
    expect(buscar).not.toHaveBeenCalled()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('começa a buscar depois que o usuário digita', async () => {
    const buscar = vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue([{ id: 1, name: 'Super Metroid' }])
    function Formulario() {
      const [nome, setNome] = useState('')
      return <GameFormAutocomplete nome={nome} onChangeNome={setNome} onSelectSugestao={vi.fn()} />
    }
    render(<Formulario />)
    fireEvent.change(screen.getByLabelText('Nome do jogo *'), { target: { value: 'Metroid' } })
    await waitFor(() => expect(buscar).toHaveBeenCalledWith('Metroid', undefined, expect.any(AbortSignal)), { timeout: 1000 })
  })
})
