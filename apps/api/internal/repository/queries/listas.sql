-- name: CriarLista :one
INSERT INTO listas (
    usuario_id, tipo, nome, descricao, regra_tipo, regra_valor, regra_igdb_id, meta
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: BuscarListaPorID :one
SELECT * FROM listas
WHERE id = $1 AND usuario_id = $2;

-- name: BuscarListaPorIDParaUpdate :one
SELECT * FROM listas
WHERE id = $1 AND usuario_id = $2
FOR UPDATE;

-- name: ListarListasPorUsuario :many
SELECT * FROM listas
WHERE usuario_id = $1
ORDER BY created_at DESC, id DESC;

-- name: AtualizarLista :one
UPDATE listas
SET nome = $3,
    descricao = $4,
    meta = $5,
    updated_at = now()
WHERE id = $1 AND usuario_id = $2
RETURNING *;

-- name: ExcluirLista :execrows
DELETE FROM listas
WHERE id = $1 AND usuario_id = $2;

-- name: CriarItemLista :one
INSERT INTO lista_itens (
    lista_id, igdb_id, nome, console, igdb_capa_url, ano_lancamento, posicao, jogo_zerado_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: BuscarItemPorID :one
SELECT i.*
FROM lista_itens i
JOIN listas l ON l.id = i.lista_id
WHERE i.id = $1 AND l.usuario_id = $2;

-- name: ListarItensPorLista :many
SELECT i.*
FROM lista_itens i
WHERE i.lista_id = $1
ORDER BY i.posicao ASC, i.id ASC;

-- name: ListarTodosItensDoUsuario :many
SELECT i.*
FROM lista_itens i
JOIN listas l ON l.id = i.lista_id
WHERE l.usuario_id = $1
ORDER BY i.lista_id, i.posicao ASC, i.id ASC;

-- name: ObterMaxPosicaoItem :one
SELECT COALESCE(MAX(posicao), 0)::int AS max_posicao
FROM lista_itens
WHERE lista_id = $1;

-- name: ExcluirItem :execrows
DELETE FROM lista_itens
WHERE id = $1 AND lista_id = $2;

-- name: AtualizarPosicaoItem :exec
UPDATE lista_itens
SET posicao = $3
WHERE id = $1 AND lista_id = $2;

-- name: AssociarJogoZeradoItem :one
UPDATE lista_itens
SET jogo_zerado_id = $3
WHERE id = $1 AND lista_id = $2
RETURNING *;

-- name: DesassociarJogoZeradoItem :one
UPDATE lista_itens
SET jogo_zerado_id = NULL
WHERE id = $1 AND lista_id = $2
RETURNING *;

-- name: ListarJogosZeradosUsuarioParaMatching :many
SELECT id, usuario_id, igdb_id, nome, console, genero, finalizado_em, nota, igdb_capa_url
FROM jogos_zerados
WHERE usuario_id = $1 AND deleted_at IS NULL
ORDER BY finalizado_em ASC, id ASC;

-- name: BuscarJogoZeradoDoUsuario :one
SELECT id, usuario_id, igdb_id, nome, console, genero, finalizado_em, nota, igdb_capa_url
FROM jogos_zerados
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;
