import { z } from 'zod'

const anoAtual = new Date().getFullYear()

export const perfilSchema = z.object({
  nome: z
    .string()
    .trim()
    .min(1, 'O nome é obrigatório.')
    .max(100, 'O nome deve ter no máximo 100 caracteres.'),
  username: z
    .string()
    .trim()
    .refine(
      (val) => !val || /^[A-Za-z0-9_]{3,30}$/.test(val),
      'O nome de usuário deve ter entre 3 e 30 caracteres e conter apenas letras, números e sublinhado (_).'
    )
    .optional()
    .nullable(),
  bio: z
    .string()
    .max(280, 'A bio deve ter no máximo 280 caracteres.')
    .optional()
    .nullable(),
  jogo_favorito_id: z.number().nullable().optional(),
  console_favorito: z
    .string()
    .max(100, 'O console favorito deve ter no máximo 100 caracteres.')
    .optional()
    .nullable(),
  jogando_desde: z
    .preprocess(
      (val) => (val === '' || val === null || val === undefined ? null : Number(val)),
      z
        .number()
        .int('O ano deve ser um número inteiro.')
        .min(1970, `O ano deve ser entre 1970 e ${anoAtual}.`)
        .max(anoAtual, `O ano deve ser entre 1970 e ${anoAtual}.`)
        .nullable()
    )
    .optional(),

})

export type PerfilFormInput = z.infer<typeof perfilSchema>

export { PERFIL_BACKEND_ERROS, FOTO_ERRO_MENSAGENS, ERRO_GENERICO_PERFIL, ERRO_GENERICO_FOTO } from './perfilErros'

