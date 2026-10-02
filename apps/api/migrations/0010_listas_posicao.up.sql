ALTER TABLE listas ADD COLUMN posicao INTEGER;

WITH ordenadas AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY usuario_id, tipo
            ORDER BY created_at DESC, id DESC
        )::INTEGER AS nova_posicao
    FROM listas
)
UPDATE listas
SET posicao = ordenadas.nova_posicao
FROM ordenadas
WHERE listas.id = ordenadas.id;

ALTER TABLE listas ALTER COLUMN posicao SET NOT NULL;

CREATE INDEX idx_listas_usuario_tipo_posicao
    ON listas (usuario_id, tipo, posicao);
