import { z } from 'zod'

function parseDate(value: string): Date | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return null
  const ano = Number(match[1])
  const mes = Number(match[2]) - 1
  const dia = Number(match[3])
  const d = new Date(ano, mes, dia)
  if (d.getFullYear() !== ano || d.getMonth() !== mes || d.getDate() !== dia) {
    return null
  }
  return d
}

export const TETO_HORAS_ABANDONADO = 100_000
export const TETO_SEGUNDOS_ABANDONADO = 360_000_000

export const abandonarJogoSchema = z
  .object({
    igdb_id: z.number().nullable().optional(),
    nome: z.string().trim().min(1, 'Nome é obrigatório').max(200, 'Nome deve ter no máximo 200 caracteres'),
    console: z
      .string()
      .trim()
      .min(1, 'Plataforma é obrigatória')
      .max(100, 'Plataforma deve ter no máximo 100 caracteres'),
    abandonado_em: z.string().trim().min(1, 'Data de abandono é obrigatória'),
    tempo_jogado_horas: z
      .number()
      .int()
      .min(0, 'Horas devem ser maiores ou iguais a 0')
      .max(TETO_HORAS_ABANDONADO, 'Tempo jogado não pode ultrapassar 100.000 horas'),
    tempo_jogado_minutos: z.number().int().min(0).max(59, 'Minutos devem ser entre 0 e 59'),
    tempo_jogado_segundos: z.number().int().min(0).max(59, 'Segundos devem ser entre 0 e 59'),
    motivo: z.string().max(500, 'Motivo deve ter no máximo 500 caracteres').nullable().optional(),
    igdb_capa_url: z.string().optional(),
    igdb_descricao: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    const hoje = new Date()
    hoje.setHours(23, 59, 59, 999)
    const dataAbandono = parseDate(data.abandonado_em)
    if (!dataAbandono) {
      ctx.addIssue({ code: 'custom', path: ['abandonado_em'], message: 'Data de abandono inválida' })
    } else if (dataAbandono > hoje) {
      ctx.addIssue({ code: 'custom', path: ['abandonado_em'], message: 'Data de abandono não pode ser futura' })
    }
    const totalSegundos = data.tempo_jogado_horas * 3600 + data.tempo_jogado_minutos * 60 + data.tempo_jogado_segundos
    if (totalSegundos > TETO_SEGUNDOS_ABANDONADO) {
      ctx.addIssue({
        code: 'custom',
        path: ['tempo_jogado_horas'],
        message: 'Tempo total não pode ultrapassar 100.000 horas',
      })
    }
  })

export type AbandonarJogoFormData = z.infer<typeof abandonarJogoSchema>
