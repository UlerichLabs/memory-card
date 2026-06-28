ALTER TABLE "campanhas" RENAME COLUMN "criada_em" TO "created_at";--> statement-breakpoint
ALTER TABLE "conquistas" RENAME COLUMN "concluida_em" TO "data_conquista";--> statement-breakpoint
ALTER TABLE "desafio_itens" RENAME COLUMN "zerado" TO "concluido";--> statement-breakpoint
ALTER TABLE "desafios" RENAME COLUMN "criado_em" TO "created_at";--> statement-breakpoint
ALTER TABLE "favoritos" RENAME COLUMN "valor" TO "jogo_id";--> statement-breakpoint
ALTER TABLE "jogos_zerados" RENAME COLUMN "tempo_segundos" TO "tempo_jogado";--> statement-breakpoint
ALTER TABLE "campanhas" ADD COLUMN "icone" text DEFAULT null;--> statement-breakpoint
ALTER TABLE "campanhas" ADD COLUMN "concluida" boolean DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE "conquistas" ADD COLUMN "created_at" timestamp DEFAULT now() NOT NULL;--> statement-breakpoint
ALTER TABLE "favoritos" ADD COLUMN "created_at" timestamp DEFAULT now() NOT NULL;