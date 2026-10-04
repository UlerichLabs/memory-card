-- name: CriarJogoEmAndamento :one
INSERT INTO jogos_em_andamento (
    usuario_id, igdb_id, igdb_capa_url, nome, iniciado_em
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListarJogosEmAndamento :many
SELECT * FROM jogos_em_andamento
WHERE usuario_id = $1 AND deleted_at IS NULL
ORDER BY iniciado_em DESC, id DESC;

-- name: ExcluirJogoEmAndamento :execrows
UPDATE jogos_em_andamento
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

-- name: DarBaixaJogoEmAndamento :execrows
UPDATE jogos_em_andamento
SET deleted_at = now(), updated_at = now()
WHERE id = (
    SELECT candidato.id
    FROM jogos_em_andamento AS candidato
    WHERE candidato.usuario_id = sqlc.arg('usuario_id')::int
      AND candidato.deleted_at IS NULL
      AND (
          (candidato.igdb_id IS NOT NULL AND sqlc.narg('igdb_id')::int IS NOT NULL AND candidato.igdb_id = sqlc.narg('igdb_id')::int)
          OR (
              unaccent(lower(btrim(candidato.nome))) = unaccent(lower(btrim(sqlc.arg('nome')::text)))
              AND NOT (
                  candidato.igdb_id IS NOT NULL
                  AND sqlc.narg('igdb_id')::int IS NOT NULL
                  AND candidato.igdb_id <> sqlc.narg('igdb_id')::int
              )
          )
      )
    ORDER BY CASE WHEN candidato.igdb_id = sqlc.narg('igdb_id')::int AND sqlc.narg('igdb_id')::int IS NOT NULL THEN 0 ELSE 1 END,
             candidato.iniciado_em ASC,
             candidato.id ASC
    LIMIT 1
)
AND usuario_id = sqlc.arg('usuario_id')::int
AND deleted_at IS NULL;
