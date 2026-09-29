-- name: CriarJogoZerado :one
INSERT INTO jogos_zerados (
    usuario_id, igdb_id, nome, console, genero, tipo,
    iniciado_em, finalizado_em, tempo_jogado, nota,
    dificuldade, review, destaque, igdb_capa_url, igdb_descricao
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: ListarJogosZerados :many
SELECT * FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
  AND (sqlc.narg('genero')::text IS NULL OR unaccent(genero) ILIKE unaccent('%' || sqlc.narg('genero')::text || '%'))
  AND (sqlc.narg('tipo')::varchar IS NULL OR LOWER(tipo) = LOWER(sqlc.narg('tipo')))
  AND (sqlc.narg('nota_min')::int IS NULL OR nota >= sqlc.narg('nota_min'))
  AND (sqlc.narg('nota_max')::int IS NULL OR nota <= sqlc.narg('nota_max'))
  AND (sqlc.narg('ano')::int IS NULL OR EXTRACT(YEAR FROM finalizado_em) = sqlc.narg('ano'))
  AND (sqlc.narg('dificuldade')::varchar IS NULL OR dificuldade = sqlc.narg('dificuldade')::dificuldade)
ORDER BY finalizado_em DESC, id DESC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ContarJogosZerados :one
SELECT COUNT(*) FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('busca')::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || sqlc.narg('busca')::text || '%'))
  AND (sqlc.narg('console')::varchar IS NULL OR console = sqlc.narg('console'))
  AND (sqlc.narg('genero')::text IS NULL OR unaccent(genero) ILIKE unaccent('%' || sqlc.narg('genero')::text || '%'))
  AND (sqlc.narg('tipo')::varchar IS NULL OR LOWER(tipo) = LOWER(sqlc.narg('tipo')))
  AND (sqlc.narg('nota_min')::int IS NULL OR nota >= sqlc.narg('nota_min'))
  AND (sqlc.narg('nota_max')::int IS NULL OR nota <= sqlc.narg('nota_max'))
  AND (sqlc.narg('ano')::int IS NULL OR EXTRACT(YEAR FROM finalizado_em) = sqlc.narg('ano'))
  AND (sqlc.narg('dificuldade')::varchar IS NULL OR dificuldade = sqlc.narg('dificuldade')::dificuldade);

-- name: ObterConsolesUsuario :many
SELECT DISTINCT console
FROM jogos_zerados
WHERE usuario_id = $1 AND deleted_at IS NULL AND console != ''
ORDER BY console ASC;

-- name: ObterGenerosUsuario :many
SELECT DISTINCT genero
FROM jogos_zerados
WHERE usuario_id = $1 AND deleted_at IS NULL AND genero IS NOT NULL AND genero != '';

-- name: ObterTiposUsuario :many
SELECT DISTINCT tipo
FROM jogos_zerados
WHERE usuario_id = $1 AND deleted_at IS NULL AND tipo IS NOT NULL AND tipo != ''
ORDER BY tipo ASC;

-- name: ObterAnosUsuario :many
SELECT DISTINCT EXTRACT(YEAR FROM finalizado_em)::int AS ano
FROM jogos_zerados
WHERE usuario_id = $1 AND deleted_at IS NULL
ORDER BY ano DESC;

-- name: BuscarJogoPorID :one
SELECT * FROM jogos_zerados WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

-- name: ObterDetalhesJogoZerado :one
WITH jogos_numerados AS (
    SELECT id, usuario_id, igdb_id, nome, console, genero, tipo,
           iniciado_em, finalizado_em, tempo_jogado, nota, dificuldade,
           review, destaque, igdb_capa_url, igdb_descricao, created_at, updated_at, deleted_at,
           ROW_NUMBER() OVER (PARTITION BY usuario_id ORDER BY created_at ASC, id ASC)::int AS numero
    FROM jogos_zerados
    WHERE usuario_id = $2
)
SELECT id, usuario_id, igdb_id, nome, console, genero, tipo,
       iniciado_em, finalizado_em, tempo_jogado, nota, dificuldade,
       review, destaque, igdb_capa_url, igdb_descricao, created_at, updated_at, numero
FROM jogos_numerados
WHERE id = $1 AND deleted_at IS NULL;

-- name: AtualizarJogoZerado :one
UPDATE jogos_zerados SET
    igdb_id = $3,
    nome = $4,
    console = $5,
    genero = $6,
    tipo = $7,
    iniciado_em = $8,
    finalizado_em = $9,
    tempo_jogado = $10,
    nota = $11,
    dificuldade = $12,
    review = $13,
    destaque = $14,
    igdb_capa_url = $15,
    igdb_descricao = $16,
    updated_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ExcluirJogoZerado :execrows
UPDATE jogos_zerados
SET deleted_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;
