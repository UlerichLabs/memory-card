import { z } from 'zod'

function parseDate(value: string): Date | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return null
  const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
  if (date.getFullYear() !== Number(match[1]) || date.getMonth() !== Number(match[2]) - 1 || date.getDate() !== Number(match[3])) return null
  return date
}

export const gameFormSchema = z.object({
  igdb_id: z.number().nullable().optional(),
  nome: z.string().trim().min(1, 'Nome é obrigatório').max(200, 'Nome deve ter no máximo 200 caracteres'),
  console: z.string().trim().min(1, 'Plataforma é obrigatória').max(100, 'Plataforma deve ter no máximo 100 caracteres'),
  genero: z.string().max(150, 'Gênero deve ter no máximo 150 caracteres').optional(),
  tipo: z.string().max(50, 'Tipo deve ter no máximo 50 caracteres').optional(),
  iniciado_em: z.string().optional(),
  finalizado_em: z.string().trim().min(1, 'Data de finalização é obrigatória'),
  tempo_jogado_horas: z.number().int().min(0, 'Horas devem ser maiores ou iguais a 0'),
  tempo_jogado_minutos: z.number().int().min(0).max(59, 'Minutos devem ser entre 0 e 59'),
  tempo_jogado_segundos: z.number().int().min(0).max(59, 'Segundos devem ser entre 0 e 59'),
  nota: z.number().int().min(1, 'Nota deve ser entre 1 e 11').max(11, 'Nota deve ser entre 1 e 11'),
  dificuldade: z.enum(['C', 'B', 'A', 'AA', 'AAA']),
  review: z.string().max(5000, 'Review deve ter no máximo 5.000 caracteres').nullable().optional(),
  destaque: z.boolean().default(false),
  igdb_capa_url: z.string().optional(),
  igdb_descricao: z.string().optional(),
}).superRefine((data, ctx) => {
  const hoje = new Date()
  hoje.setHours(23, 59, 59, 999)
  const iniciado = data.iniciado_em ? parseDate(data.iniciado_em) : null
  const finalizado = parseDate(data.finalizado_em)
  if (data.iniciado_em && !iniciado) ctx.addIssue({ code: 'custom', path: ['iniciado_em'], message: 'Data de início inválida' })
  if (!finalizado) ctx.addIssue({ code: 'custom', path: ['finalizado_em'], message: 'Data de finalização inválida' })
  if (iniciado && iniciado > hoje) ctx.addIssue({ code: 'custom', path: ['iniciado_em'], message: 'Data de início não pode ser futura' })
  if (finalizado && finalizado > hoje) ctx.addIssue({ code: 'custom', path: ['finalizado_em'], message: 'Data de finalização não pode ser futura' })
  if (iniciado && finalizado && iniciado > finalizado) ctx.addIssue({ code: 'custom', path: ['iniciado_em'], message: 'Início deve ser anterior à finalização' })
})

export type GameFormData = z.infer<typeof gameFormSchema>
