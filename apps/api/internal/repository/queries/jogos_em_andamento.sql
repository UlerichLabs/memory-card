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
