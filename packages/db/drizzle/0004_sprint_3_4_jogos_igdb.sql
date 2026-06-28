ALTER TABLE "jogos_zerados" ALTER COLUMN "nome" SET DATA TYPE varchar(300);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ALTER COLUMN "console" SET DATA TYPE varchar(100);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ALTER COLUMN "genero" SET DATA TYPE varchar(100);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ALTER COLUMN "tipo" SET DATA TYPE varchar(100);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ADD COLUMN "igdb_capa_url" varchar(500);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ADD COLUMN "igdb_descricao" text;--> statement-breakpoint
ALTER TABLE "jogos_zerados" ADD COLUMN "deleted_at" timestamp;