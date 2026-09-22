import { z } from 'zod'

export const gameFormSchema = z.object({
  igdb_id: z.number().nullable().optional(),
  nome: z.string().trim().min(1, 'Nome é obrigatório'),
  console: z.string().trim().min(1, 'Console é obrigatório'),
  genero: z.string().optional(),
  tipo: z.string().optional(),
  iniciado_em: z.string().optional(),
  finalizado_em: z.string().trim().min(1, 'Data de finalização é obrigatória'),
  tempo_jogado_horas: z.coerce.number().int().min(0, 'Horas devem ser maiores ou iguais a 0').default(0),
  tempo_jogado_minutos: z.coerce.number().int().min(0, 'Minutos devem ser entre 0 e 59').max(59, 'Minutos devem ser entre 0 e 59').default(0),
  tempo_jogado_segundos: z.coerce.number().int().min(0, 'Segundos devem ser entre 0 e 59').max(59, 'Segundos devem ser entre 0 e 59').default(0),
  nota: z.coerce.number().int().min(1, 'Nota deve ser entre 1 e 11').max(11, 'Nota deve ser entre 1 e 11'),
  dificuldade: z.enum(['C', 'B', 'A', 'AA', 'AAA']),
  condicao_zeramento: z.string().max(500, 'Condição de zeramento deve ter no máximo 500 caracteres').optional(),
  destaque: z.boolean().default(false),
  igdb_capa_url: z.string().optional(),
  igdb_descricao: z.string().optional(),
})

export type GameFormData = z.infer<typeof gameFormSchema>
