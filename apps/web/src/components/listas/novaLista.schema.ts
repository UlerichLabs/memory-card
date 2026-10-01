import { z } from "zod";

export const novaListaSchema = z
  .object({
    tipo: z.enum(["fila", "desafio"]),
    nome: z
      .string()
      .trim()
      .min(1, "O nome da lista é obrigatório.")
      .max(100, "O nome deve ter no máximo 100 caracteres."),
    descricao: z
      .string()
      .max(200, "A descrição deve ter no máximo 200 caracteres.")
      .optional(),
    origemTipo: z.enum(["franquia", "plataforma", "genero"]).optional(),
    origemId: z.number().positive().optional(),
    origemNome: z.string().trim().optional(),
  })
  .superRefine((valor, contexto) => {
    if (valor.tipo === "desafio") {
      if (!valor.origemTipo)
        contexto.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Escolha de onde vêm os jogos.",
          path: ["origemTipo"],
        });
      if (!valor.origemId || !valor.origemNome)
        contexto.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Escolha uma opção da lista.",
          path: ["origemNome"],
        });
    }
  });

export type NovaListaFormValues = z.infer<typeof novaListaSchema>;
