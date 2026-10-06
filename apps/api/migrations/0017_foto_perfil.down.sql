ALTER TABLE usuarios
    DROP CONSTRAINT IF EXISTS check_usuarios_avatar_mutuamente_exclusivo;

ALTER TABLE usuarios
    DROP COLUMN IF EXISTS avatar_jogo_id;
