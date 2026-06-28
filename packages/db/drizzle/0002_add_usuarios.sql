CREATE TABLE "usuarios" (
	"id" serial PRIMARY KEY NOT NULL,
	"nome" varchar(100) NOT NULL,
	"username" varchar(50) NOT NULL,
	"email" varchar(255) NOT NULL,
	"senha_hash" varchar(255) NOT NULL,
	"avatar" varchar(500),
	"bio" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	"deleted_at" timestamp,
	CONSTRAINT "usuarios_username_unique" UNIQUE("username"),
	CONSTRAINT "usuarios_email_unique" UNIQUE("email")
);--> statement-breakpoint
ALTER TABLE "jogos_zerados" ADD COLUMN "usuario_id" integer;--> statement-breakpoint
DO $$
DECLARE
	system_user_id integer;
BEGIN
	IF EXISTS (SELECT 1 FROM "jogos_zerados" WHERE "usuario_id" IS NULL) THEN
		INSERT INTO "usuarios" ("nome", "username", "email", "senha_hash")
		VALUES ('Sistema', 'sistema', 'sistema@memory-card.local', 'migration-placeholder')
		ON CONFLICT ("email") DO UPDATE SET "email" = EXCLUDED."email"
		RETURNING "id" INTO system_user_id;

		UPDATE "jogos_zerados"
		SET "usuario_id" = system_user_id
		WHERE "usuario_id" IS NULL;
	END IF;
END $$;--> statement-breakpoint
ALTER TABLE "jogos_zerados" ALTER COLUMN "usuario_id" SET NOT NULL;--> statement-breakpoint
ALTER TABLE "jogos_zerados" ADD CONSTRAINT "jogos_zerados_usuario_id_usuarios_id_fk" FOREIGN KEY ("usuario_id") REFERENCES "public"."usuarios"("id") ON DELETE no action ON UPDATE no action;
