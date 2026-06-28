CREATE TYPE "public"."tipo_desafio" AS ENUM('QUANTIDADE', 'LISTA_COMPLETA', 'FRANQUIA', 'EMPRESA', 'GENERO', 'MANUAL');--> statement-breakpoint
CREATE TYPE "public"."tipo_favorito" AS ENUM('JOGO', 'GENERO');--> statement-breakpoint
CREATE TABLE "campanhas" (
	"id" serial PRIMARY KEY NOT NULL,
	"usuario_id" integer NOT NULL,
	"nome" varchar(200) NOT NULL,
	"descricao" text,
	"criada_em" timestamp DEFAULT now() NOT NULL,
	"deleted_at" timestamp
);
--> statement-breakpoint
CREATE TABLE "conquistas" (
	"id" serial PRIMARY KEY NOT NULL,
	"usuario_id" integer NOT NULL,
	"desafio_id" integer,
	"nome" varchar(200) NOT NULL,
	"descricao" text,
	"concluida_em" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "desafio_itens" (
	"id" serial PRIMARY KEY NOT NULL,
	"desafio_id" integer NOT NULL,
	"igdb_id" integer NOT NULL,
	"nome" varchar(300) NOT NULL,
	"zerado" boolean DEFAULT false NOT NULL,
	"zeramento_id" integer
);
--> statement-breakpoint
CREATE TABLE "desafios" (
	"id" serial PRIMARY KEY NOT NULL,
	"usuario_id" integer NOT NULL,
	"campanha_id" integer,
	"nome" varchar(200) NOT NULL,
	"tipo" "tipo_desafio" NOT NULL,
	"filtro_valor" varchar(200),
	"meta" integer,
	"concluida_em" timestamp,
	"criado_em" timestamp DEFAULT now() NOT NULL,
	"deleted_at" timestamp
);
--> statement-breakpoint
CREATE TABLE "favoritos" (
	"id" serial PRIMARY KEY NOT NULL,
	"usuario_id" integer NOT NULL,
	"tipo" "tipo_favorito" NOT NULL,
	"valor" varchar(200) NOT NULL,
	"posicao" integer NOT NULL
);
--> statement-breakpoint
ALTER TABLE "jogos_zerados" ALTER COLUMN "dificuldade" SET DATA TYPE text;--> statement-breakpoint
DROP TYPE "public"."dificuldade";--> statement-breakpoint
CREATE TYPE "public"."dificuldade" AS ENUM('C', 'B', 'A', 'AA', 'AAA');--> statement-breakpoint
ALTER TABLE "jogos_zerados"
  ALTER COLUMN "dificuldade" TYPE "public"."dificuldade"
  USING (
    CASE "dificuldade"
      WHEN 'S' THEN 'AAA'
      WHEN 'A' THEN 'AA'
      WHEN 'B' THEN 'A'
      WHEN 'C' THEN 'B'
      WHEN 'D' THEN 'C'
      ELSE "dificuldade"
    END
  )::"public"."dificuldade";--> statement-breakpoint
ALTER TABLE "usuarios" ADD COLUMN "redes_sociais" json DEFAULT '{}'::json;--> statement-breakpoint
ALTER TABLE "campanhas" ADD CONSTRAINT "campanhas_usuario_id_usuarios_id_fk" FOREIGN KEY ("usuario_id") REFERENCES "public"."usuarios"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "conquistas" ADD CONSTRAINT "conquistas_usuario_id_usuarios_id_fk" FOREIGN KEY ("usuario_id") REFERENCES "public"."usuarios"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "conquistas" ADD CONSTRAINT "conquistas_desafio_id_desafios_id_fk" FOREIGN KEY ("desafio_id") REFERENCES "public"."desafios"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "desafio_itens" ADD CONSTRAINT "desafio_itens_desafio_id_desafios_id_fk" FOREIGN KEY ("desafio_id") REFERENCES "public"."desafios"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "desafio_itens" ADD CONSTRAINT "desafio_itens_zeramento_id_jogos_zerados_id_fk" FOREIGN KEY ("zeramento_id") REFERENCES "public"."jogos_zerados"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "desafios" ADD CONSTRAINT "desafios_usuario_id_usuarios_id_fk" FOREIGN KEY ("usuario_id") REFERENCES "public"."usuarios"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "desafios" ADD CONSTRAINT "desafios_campanha_id_campanhas_id_fk" FOREIGN KEY ("campanha_id") REFERENCES "public"."campanhas"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "favoritos" ADD CONSTRAINT "favoritos_usuario_id_usuarios_id_fk" FOREIGN KEY ("usuario_id") REFERENCES "public"."usuarios"("id") ON DELETE no action ON UPDATE no action;
