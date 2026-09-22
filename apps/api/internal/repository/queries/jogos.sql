-- name: CriarJogoZerado :one
INSERT INTO jogos_zerados (
    usuario_id, igdb_id, nome, console, genero, tipo,
    iniciado_em, finalizado_em, tempo_jogado, nota,
    dificuldade, condicao_zeramento, destaque, igdb_capa_url, igdb_descricao
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: ListarJogosZerados :many
SELECT * FROM jogos_zerados
WHERE usuario_id = $1
  AND deleted_at IS NULL
  AND ($2::varchar IS NULL OR console = $2)
  AND ($3::varchar IS NULL OR genero = $3)
  AND ($4::int IS NULL OR nota = $4)
  AND ($5::text IS NULL OR nome ILIKE '%' || $5 || '%')
ORDER BY finalizado_em DESC
LIMIT $6 OFFSET $7;

-- name: BuscarJogoPorID :one
SELECT * FROM jogos_zerados WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

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
    condicao_zeramento = $13,
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

