import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PerfilJogador } from './PerfilJogador'

const resumo = { total_jogos: 8, total_segundos: 7200, media_segundos_por_jogo: 900, nota_media: 9.5, jogos_no_ano_atual: 2, primeiro_zeramento_em: '2020-01-01T00:00:00Z', dias_desde_primeiro: 2000, anos_desde_primeiro: 5 }

describe('PerfilJogador', () => {
  it('renderiza identidade, ano e métricas', () => {
    render(<PerfilJogador nome="Ana Silva" resumo={resumo} totalAbandonados={2} />)
    expect(screen.getByText('AS')).toBeInTheDocument()
    expect(screen.getByText('Jogando desde 2020')).toBeInTheDocument()
    expect(screen.getByText('Jogos zerados')).toBeInTheDocument()
    expect(screen.getByText('9.5')).toBeInTheDocument()
  })

  it('omite a jornada de início quando o dado está ausente', () => {
    render(<PerfilJogador nome="Ana" resumo={{ ...resumo, primeiro_zeramento_em: null }} totalAbandonados={0} />)
    expect(screen.queryByText(/Jogando desde/)).not.toBeInTheDocument()
  })
})
