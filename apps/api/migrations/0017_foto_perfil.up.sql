UPDATE usuarios
SET avatar_url = NULL
WHERE avatar_url IS NOT NULL
  AND avatar_url !~ '^[a-f0-9]{32}\.jpg$';

ALTER TABLE usuarios
    ADD COLUMN IF NOT EXISTS avatar_jogo_id INTEGER REFERENCES jogos_zerados(id) ON DELETE SET NULL;

ALTER TABLE usuarios
    ADD CONSTRAINT check_usuarios_avatar_mutuamente_exclusivo
    CHECK (NOT (avatar_url IS NOT NULL AND avatar_jogo_id IS NOT NULL));
