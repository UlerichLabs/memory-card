-- name: CriarJogoAbandonado :one
INSERT INTO jogos_abandonados (
    usuario_id, igdb_id, igdb_capa_url, nome, console,
    tempo_jogado, motivo, abandonado_em, iniciado_em
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: BuscarJogoAbandonadoPorID :one
SELECT * FROM jogos_abandonados
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

-- name: AtualizarJogoAbandonado :one
UPDATE jogos_abandonados SET
    igdb_id = $3,
    igdb_capa_url = $4,
    nome = $5,
    console = $6,
    tempo_jogado = $7,
    motivo = $8,
    abandonado_em = $9,
    iniciado_em = $10,
    updated_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ExcluirJogoAbandonado :execrows
UPDATE jogos_abandonados
SET deleted_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

-- name: ContarJogosAbandonados :one
SELECT COUNT(*) FROM jogos_abandonados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'));

-- name: ListarJogosAbandonadosRecentes :many
SELECT * FROM jogos_abandonados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
ORDER BY abandonado_em DESC, id DESC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ListarJogosAbandonadosAntigos :many
SELECT * FROM jogos_abandonados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
ORDER BY abandonado_em ASC, id ASC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ListarJogosAbandonadosNome :many
SELECT * FROM jogos_abandonados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
ORDER BY unaccent(nome) ASC, id ASC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ListarJogosAbandonadosTempo :many
SELECT * FROM jogos_abandonados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
ORDER BY tempo_jogado DESC, id DESC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ObterConsolesAbandonadosUsuario :many
SELECT DISTINCT console
FROM jogos_abandonados
WHERE usuario_id = $1 AND deleted_at IS NULL AND console != ''
ORDER BY console ASC;

-- name: ObterTotalAbandonadosUsuario :one
SELECT COUNT(*) FROM jogos_abandonados
WHERE usuario_id = $1 AND deleted_at IS NULL;
