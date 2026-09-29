import { z } from 'zod'

export const novaListaSchema = z.object({
  tipo: z.enum(['fila', 'desafio']),
  nome: z.string().trim().min(1, 'O nome da lista é obrigatório.').max(100, 'O nome deve ter no máximo 100 caracteres.'),
  descricao: z.string().max(200, 'A descrição deve ter no máximo 200 caracteres.').optional(),
  regraTipo: z.enum(['franquia', 'plataforma', 'genero', 'manual']).optional(),
  regraValor: z.string().trim().optional(),
  igdbId: z.number().nullable().optional(),
  meta: z.union([z.number().int().min(1, 'A meta deve ser no mínimo 1.').max(10000, 'A meta deve ser no máximo 10.000.'), z.nan(), z.null()]).optional(),
}).superRefine((val, ctx) => {
  if (val.tipo === 'desafio') {
    if (!val.regraTipo) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Selecione uma regra para o desafio.', path: ['regraTipo'] })
      return
    }
    if (val.regraTipo === 'franquia') {
      if (!val.igdbId) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Selecione uma franquia da lista.', path: ['regraValor'] })
      }
    }
    if (val.regraTipo === 'plataforma') {
      if (!val.regraValor || val.regraValor.trim().length === 0) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Informe a plataforma.', path: ['regraValor'] })
      } else if (val.regraValor.trim().length > 150) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'A plataforma deve ter no máximo 150 caracteres.', path: ['regraValor'] })
      }
      if (typeof val.meta !== 'number' || Number.isNaN(val.meta) || val.meta < 1) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'A meta é obrigatória para este desafio.', path: ['meta'] })
      }
    }
    if (val.regraTipo === 'genero') {
      if (!val.regraValor || val.regraValor.trim().length === 0) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Informe o gênero.', path: ['regraValor'] })
      } else if (val.regraValor.trim().length > 150) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'O gênero deve ter no máximo 150 caracteres.', path: ['regraValor'] })
      }
      if (typeof val.meta !== 'number' || Number.isNaN(val.meta) || val.meta < 1) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'A meta é obrigatória para este desafio.', path: ['meta'] })
      }
    }
  }
})

export type NovaListaFormValues = z.infer<typeof novaListaSchema>
