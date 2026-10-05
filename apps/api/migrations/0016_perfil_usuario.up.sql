ALTER TABLE usuarios
    ADD COLUMN IF NOT EXISTS jogo_favorito_id INTEGER REFERENCES jogos_zerados(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS console_favorito VARCHAR(100),
    ADD COLUMN IF NOT EXISTS jogando_desde SMALLINT;

ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_username_key;

WITH duplicados AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY lower(username) ORDER BY id ASC) as rn
    FROM usuarios
    WHERE username IS NOT NULL
)
UPDATE usuarios
SET username = NULL
WHERE id IN (SELECT id FROM duplicados WHERE rn > 1);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usuarios_username_lower ON usuarios (lower(username)) WHERE username IS NOT NULL;
