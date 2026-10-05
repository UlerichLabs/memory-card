import { describe, expect, it } from 'vitest'
import {
  diasDesdeInicio,
  encontrarJogoEmAndamento,
  formatarDataJogando,
  hojeIso,
  rotuloDiasDesdeInicio,
} from './jogandoUtils'
import type { JogoEmAndamento } from '@/types/jogando'

describe('jogandoUtils', () => {
  it('formata datas sem deslocamento de fuso', () => {
    expect(formatarDataJogando('2026-09-20T00:00:00Z')).toBe('20/09/2026')
    expect(formatarDataJogando(null)).toBe('sem data')
  })

  it('calcula dias por calendário local, incluindo viradas', () => {
    expect(diasDesdeInicio('2026-09-20T00:00:00Z', new Date(2026, 8, 20))).toBe(0)
    expect(diasDesdeInicio('2026-09-19T00:00:00Z', new Date(2026, 8, 20))).toBe(1)
    expect(diasDesdeInicio('2026-08-31T00:00:00Z', new Date(2026, 8, 2))).toBe(2)
    expect(diasDesdeInicio('2025-12-31T00:00:00Z', new Date(2026, 0, 1))).toBe(1)
  })

  it('produz os rótulos relativos', () => {
    const hoje = new Date(2026, 8, 20)
    expect(rotuloDiasDesdeInicio('2026-09-20T00:00:00Z', hoje)).toBe('começou hoje')
    expect(rotuloDiasDesdeInicio('2026-09-19T00:00:00Z', hoje)).toBe('há 1 dia')
    expect(rotuloDiasDesdeInicio('2026-09-17T00:00:00Z', hoje)).toBe('há 3 dias')
    expect(hojeIso(hoje)).toBe('2026-09-20')
  })

  describe('encontrarJogoEmAndamento', () => {
    const baseJogo = (override: Partial<JogoEmAndamento>): JogoEmAndamento => ({
      id: 1,
      nome: 'Chrono Trigger',
      igdb_id: null,
      igdb_capa_url: null,
      iniciado_em: '2026-09-01T00:00:00Z',
      ...override,
    })

    it('encontra por mesmo igdb_id', () => {
      const jogo = baseJogo({ id: 10, igdb_id: 42, nome: 'Nome antigo' })
      const match = encontrarJogoEmAndamento([jogo], { nome: 'Outro nome', igdbId: 42 })
      expect(match).toBe(jogo)
    })

    it('encontra por nome com acento, caixa e espaços diferentes', () => {
      const jogo = baseJogo({ id: 2, nome: '  Chrôno Trigger  ' })
      const match = encontrarJogoEmAndamento([jogo], { nome: ' chrono TRIGGER ' })
      expect(match).toBe(jogo)
    })

    it('retorna undefined para nome parecido mas diferente', () => {
      const jogo = baseJogo({ id: 3, nome: 'Kaze and the Wild Masks' })
      const match = encontrarJogoEmAndamento([jogo], { nome: 'Kaze and the Wild Masks (Steam)' })
      expect(match).toBeUndefined()
    })

    it('retorna undefined quando igdb_ids são diferentes mesmo com nome igual', () => {
      const jogo = baseJogo({ id: 4, nome: 'Chrono Trigger', igdb_id: 100 })
      const match = encontrarJogoEmAndamento([jogo], { nome: 'Chrono Trigger', igdbId: 200 })
      expect(match).toBeUndefined()
    })

    it('encontra quando um dos lados não tem igdb_id e os nomes são iguais', () => {
      const jogoSemIgdb = baseJogo({ id: 5, nome: 'Kaze', igdb_id: null })
      expect(encontrarJogoEmAndamento([jogoSemIgdb], { nome: 'Kaze', igdbId: 300 })).toBe(jogoSemIgdb)

      const jogoComIgdb = baseJogo({ id: 6, nome: 'Kaze', igdb_id: 300 })
      expect(encontrarJogoEmAndamento([jogoComIgdb], { nome: 'Kaze', igdbId: null })).toBe(jogoComIgdb)
    })

    it('prioriza match por igdb_id sobre match por nome', () => {
      const porNome = baseJogo({ id: 7, nome: 'Elden Ring', igdb_id: null, iniciado_em: '2026-01-01' })
      const porIgdb = baseJogo({ id: 8, nome: 'Elden Ring Shadow', igdb_id: 500, iniciado_em: '2026-06-01' })
      const match = encontrarJogoEmAndamento([porNome, porIgdb], { nome: 'Elden Ring', igdbId: 500 })
      expect(match).toBe(porIgdb)
    })

    it('quando ambos empatam na prioridade, seleciona o mais antigo e desempata por id', () => {
      const recente = baseJogo({ id: 9, igdb_id: 777, iniciado_em: '2026-05-10' })
      const antigo = baseJogo({ id: 10, igdb_id: 777, iniciado_em: '2026-01-10' })
      expect(encontrarJogoEmAndamento([recente, antigo], { igdbId: 777 })).toBe(antigo)

      const mesmoDiaId2 = baseJogo({ id: 20, igdb_id: 888, iniciado_em: '2026-03-01' })
      const mesmoDiaId1 = baseJogo({ id: 15, igdb_id: 888, iniciado_em: '2026-03-01' })
      expect(encontrarJogoEmAndamento([mesmoDiaId2, mesmoDiaId1], { igdbId: 888 })).toBe(mesmoDiaId1)
    })

    it('retorna undefined para lista vazia', () => {
      expect(encontrarJogoEmAndamento([], { nome: 'Zelda', igdbId: 10 })).toBeUndefined()
    })
  })
})
