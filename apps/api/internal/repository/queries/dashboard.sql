-- name: ObterResumoDashboard :one
SELECT
    COUNT(*)::bigint AS total_jogos,
    COALESCE(SUM(tempo_jogado), 0)::bigint AS total_segundos,
    COALESCE(SUM(nota), 0)::bigint AS soma_notas,
    COUNT(nota)::bigint AS total_avaliados,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM finalizado_em) = sqlc.arg('ano_atual')::int)::bigint AS jogos_no_ano_atual,
    MIN(finalizado_em)::timestamp AS primeiro_zeramento_em
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL;

-- name: ListarEstatisticasPorAno :many
WITH por_ano AS (
    SELECT
        EXTRACT(YEAR FROM j.finalizado_em)::int AS ano,
        COUNT(*)::bigint AS total_jogos,
        COALESCE(SUM(j.tempo_jogado), 0)::bigint AS total_segundos
    FROM jogos_zerados j
    WHERE j.usuario_id = sqlc.arg('usuario_id') AND j.deleted_at IS NULL
    GROUP BY EXTRACT(YEAR FROM j.finalizado_em)::int
),
destaques AS (
    SELECT
        EXTRACT(YEAR FROM d.finalizado_em)::int AS ano,
        d.id,
        d.nome,
        d.console,
        COALESCE(d.igdb_capa_url, '')::text AS igdb_capa_url,
        d.nota
    FROM jogos_zerados d
    WHERE d.usuario_id = sqlc.arg('usuario_id') AND d.deleted_at IS NULL AND d.destaque = true
)
SELECT
    pa.ano,
    pa.total_jogos,
    pa.total_segundos,
    d.id AS destaque_id,
    d.nome AS destaque_nome,
    d.console AS destaque_console,
    d.igdb_capa_url AS destaque_igdb_capa_url,
    d.nota AS destaque_nota
FROM por_ano pa
LEFT JOIN destaques d ON pa.ano = d.ano
ORDER BY pa.ano ASC;

-- name: ObterRankingPlataformas :many
SELECT
    console,
    COUNT(*)::bigint AS total_jogos,
    COALESCE(SUM(tempo_jogado), 0)::bigint AS total_segundos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND console IS NOT NULL
  AND console != ''
GROUP BY console
ORDER BY total_jogos DESC, total_segundos DESC, console ASC;

-- name: ObterRankingGeneros :many
SELECT
    genero,
    COUNT(*)::bigint AS total_jogos,
    COALESCE(SUM(tempo_jogado), 0)::bigint AS total_segundos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND genero IS NOT NULL
  AND genero != ''
GROUP BY genero
ORDER BY total_jogos DESC, total_segundos DESC, genero ASC;

-- name: ObterBreakdownTipo :many
SELECT
    tipo,
    COUNT(*)::bigint AS total_jogos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND tipo IS NOT NULL
  AND tipo != ''
  AND genero IS NOT NULL
  AND unaccent(LOWER(genero)) = unaccent(LOWER(sqlc.arg('genero')::text))
GROUP BY tipo
ORDER BY total_jogos DESC, tipo ASC;

-- name: ObterRecordeMaisLongo :one
SELECT
    id,
    nome,
    console,
    EXTRACT(YEAR FROM finalizado_em)::int AS ano,
    COALESCE(igdb_capa_url, '')::text AS igdb_capa_url,
    tempo_jogado AS tempo_jogado_segundos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
ORDER BY tempo_jogado DESC, finalizado_em DESC, id DESC
LIMIT 1;

-- name: ObterRecordeMaisCurto :one
SELECT
    id,
    nome,
    console,
    EXTRACT(YEAR FROM finalizado_em)::int AS ano,
    COALESCE(igdb_capa_url, '')::text AS igdb_capa_url,
    tempo_jogado AS tempo_jogado_segundos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND tempo_jogado > 0
ORDER BY tempo_jogado ASC, finalizado_em DESC, id DESC
LIMIT 1;

-- name: ObterDistribuicaoNotas :many
SELECT
    nota,
    COUNT(*)::bigint AS total
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND nota IS NOT NULL
GROUP BY nota
ORDER BY nota ASC;

-- name: ObterDistribuicaoDificuldade :many
SELECT
    dificuldade,
    COUNT(*)::bigint AS total_jogos
FROM jogos_zerados
WHERE usuario_id = sqlc.arg('usuario_id')
  AND deleted_at IS NULL
  AND dificuldade IS NOT NULL
GROUP BY dificuldade;
