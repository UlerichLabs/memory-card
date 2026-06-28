ALTER TYPE "public"."dificuldade" RENAME TO "dificuldade_old";--> statement-breakpoint
CREATE TYPE "public"."dificuldade" AS ENUM('S', 'A', 'B', 'C', 'D');--> statement-breakpoint
ALTER TABLE "jogos_zerados"
  ALTER COLUMN "dificuldade" TYPE "public"."dificuldade"
  USING (
    CASE "dificuldade"::text
      WHEN 'AAA' THEN 'S'
      WHEN 'AA' THEN 'A'
      WHEN 'A' THEN 'B'
      WHEN 'B' THEN 'C'
      WHEN 'C' THEN 'D'
      ELSE 'B'
    END
  )::"public"."dificuldade";--> statement-breakpoint
DROP TYPE "public"."dificuldade_old";
