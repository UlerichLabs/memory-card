CREATE TYPE "public"."dificuldade" AS ENUM('C', 'B', 'A', 'AA', 'AAA');--> statement-breakpoint
CREATE TABLE "jogos_zerados" (
	"id" serial PRIMARY KEY NOT NULL,
	"igdb_id" integer,
	"nome" varchar(200) NOT NULL,
	"console" varchar(50) NOT NULL,
	"genero" varchar(50),
	"tipo" varchar(50),
	"iniciado_em" timestamp,
	"finalizado_em" timestamp NOT NULL,
	"tempo_segundos" integer NOT NULL,
	"nota" integer NOT NULL,
	"dificuldade" "dificuldade" NOT NULL,
	"condicao_zeramento" text,
	"destaque" boolean DEFAULT false NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "jogos_zerados_nota_range" CHECK (nota between 1 and 11)
);
