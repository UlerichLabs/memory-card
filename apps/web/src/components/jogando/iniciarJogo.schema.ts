import { z } from 'zod'

function dataValida(value: string): boolean {
  const partes = value.split('-').map(Number)
  if (partes.length !== 3 || partes.some(Number.isNaN)) return false
  const data = new Date(partes[0], partes[1] - 1, partes[2])
  return data.getFullYear() === partes[0] && data.getMonth() === partes[1] - 1 && data.getDate() === partes[2]
}

export const iniciarJogoSchema = z.object({
  nome: z.string().trim().min(1, 'Nome é obrigatório').max(200, 'Nome deve ter no máximo 200 caracteres'),
  iniciado_em: z.string().min(1, 'Data de início é obrigatória').refine(dataValida, 'Data de início inválida'),
  igdb_id: z.number().nullable().optional(),
  igdb_capa_url: z.string().nullable().optional(),
}).superRefine((data, ctx) => {
  const hoje = new Date()
  const partes = data.iniciado_em.split('-').map(Number)
  const inicio = new Date(partes[0], partes[1] - 1, partes[2])
  hoje.setHours(23, 59, 59, 999)
  if (dataValida(data.iniciado_em) && inicio > hoje) {
    ctx.addIssue({ code: 'custom', path: ['iniciado_em'], message: 'Data de início não pode ser futura' })
  }
})

export type IniciarJogoFormData = z.infer<typeof iniciarJogoSchema>
