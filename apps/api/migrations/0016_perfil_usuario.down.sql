DROP INDEX IF EXISTS idx_usuarios_username_lower;

ALTER TABLE usuarios
    DROP COLUMN IF EXISTS jogo_favorito_id,
    DROP COLUMN IF EXISTS console_favorito,
    DROP COLUMN IF EXISTS jogando_desde;

ALTER TABLE usuarios ADD CONSTRAINT usuarios_username_key UNIQUE (username);
