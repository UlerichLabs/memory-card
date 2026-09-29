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

-- name: ListarJogosZeradosPorNota :many
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
ORDER BY nota DESC, finalizado_em DESC, id DESC
LIMIT sqlc.arg('limite')::int OFFSET sqlc.arg('offset_val')::int;

-- name: ObterResumoGameDoAno :many
WITH anos AS (
    SELECT
        EXTRACT(YEAR FROM j.finalizado_em)::int AS ano,
        COUNT(*)::bigint AS total_jogos
    FROM jogos_zerados j
    WHERE j.usuario_id = sqlc.arg('usuario_id')::int AND j.deleted_at IS NULL
    GROUP BY EXTRACT(YEAR FROM j.finalizado_em)::int
),
destaques AS (
    SELECT
        EXTRACT(YEAR FROM d.finalizado_em)::int AS ano,
        d.id, d.usuario_id, d.igdb_id, d.nome, d.console, d.genero, d.tipo,
        d.iniciado_em, d.finalizado_em, d.tempo_jogado, d.nota, d.dificuldade,
        d.review, d.destaque, d.igdb_capa_url, d.igdb_descricao, d.created_at, d.updated_at
    FROM jogos_zerados d
    WHERE d.usuario_id = sqlc.arg('usuario_id')::int AND d.deleted_at IS NULL AND d.destaque = true
)
SELECT
    a.ano,
    a.total_jogos,
    d.id AS destaque_id,
    d.usuario_id AS destaque_usuario_id,
    d.igdb_id AS destaque_igdb_id,
    d.nome AS destaque_nome,
    d.console AS destaque_console,
    d.genero AS destaque_genero,
    d.tipo AS destaque_tipo,
    d.iniciado_em AS destaque_iniciado_em,
    d.finalizado_em AS destaque_finalizado_em,
    d.tempo_jogado AS destaque_tempo_jogado,
    d.nota AS destaque_nota,
    d.dificuldade AS destaque_dificuldade,
    d.review AS destaque_review,
    d.destaque AS destaque_destaque,
    d.igdb_capa_url AS destaque_igdb_capa_url,
    d.created_at AS destaque_created_at,
    d.updated_at AS destaque_updated_at
FROM anos a
LEFT JOIN destaques d ON a.ano = d.ano
ORDER BY a.ano DESC;

-- name: BuscarJogoPorIDParaUpdate :one
SELECT * FROM jogos_zerados
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL
FOR UPDATE;

-- name: DesmarcarGameDoAnoAtual :one
UPDATE jogos_zerados
SET destaque = false, updated_at = now()
WHERE usuario_id = sqlc.arg('usuario_id')
  AND EXTRACT(YEAR FROM finalizado_em)::int = sqlc.arg('ano')::int
  AND destaque = true
  AND deleted_at IS NULL
RETURNING id;

-- name: MarcarGameDoAno :one
UPDATE jogos_zerados
SET destaque = true, updated_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DesmarcarGameDoAnoPorID :execrows
UPDATE jogos_zerados
SET destaque = false, updated_at = now()
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

