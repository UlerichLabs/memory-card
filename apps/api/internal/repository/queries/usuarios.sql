-- name: CriarUsuario :one
INSERT INTO usuarios (
    nome,
    email,
    senha_hash
) VALUES (
    $1, $2, $3
)
RETURNING id, nome, email, username, avatar_url, bio, idioma, created_at;

-- name: ExisteUsuarioComEmail :one
SELECT EXISTS(
    SELECT 1 FROM usuarios WHERE email = $1
);

-- name: BuscarUsuarioPorEmail :one
SELECT id, nome, email, senha_hash, username, avatar_url, bio, idioma, created_at
FROM usuarios
WHERE email = $1;

-- name: BuscarUsuarioPorID :one
SELECT id, nome, email, idioma, created_at
FROM usuarios
WHERE id = $1;

-- name: BuscarCredenciaisUsuarioPorID :one
SELECT id, nome, email, senha_hash, username, avatar_url, bio, idioma, created_at
FROM usuarios
WHERE id = $1;

-- name: AtualizarSenhaUsuario :exec
UPDATE usuarios
SET senha_hash = $2
WHERE id = $1;

-- name: BuscarPerfilCompletoPorID :one
SELECT
    u.id,
    u.nome,
    u.email,
    u.username,
    u.avatar_url,
    u.avatar_jogo_id,
    aj.igdb_capa_url AS avatar_jogo_capa_url,
    u.bio,
    u.jogo_favorito_id,
    u.console_favorito,
    u.jogando_desde,
    j.id AS fav_id,
    j.nome AS fav_nome,
    j.igdb_capa_url AS fav_capa_url,
    COALESCE((
        SELECT EXTRACT(YEAR FROM MIN(finalizado_em))::int
        FROM jogos_zerados
        WHERE usuario_id = u.id AND deleted_at IS NULL
    ), 0)::int AS primeiro_ano_zerado
FROM usuarios u
LEFT JOIN jogos_zerados j ON j.id = u.jogo_favorito_id AND j.usuario_id = u.id AND j.deleted_at IS NULL
LEFT JOIN jogos_zerados aj ON aj.id = u.avatar_jogo_id AND aj.usuario_id = u.id AND aj.deleted_at IS NULL
WHERE u.id = $1;

-- name: ExisteJogoZeradoDoUsuario :one
SELECT EXISTS(
    SELECT 1 FROM jogos_zerados
    WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL
);

-- name: AtualizarPerfilUsuario :execrows
UPDATE usuarios
SET
    nome = $2,
    username = $3,
    bio = $4,
    jogo_favorito_id = $5,
    console_favorito = $6,
    jogando_desde = $7
WHERE id = $1;

-- name: BuscarAvatarUsuario :one
SELECT avatar_url, avatar_jogo_id
FROM usuarios
WHERE id = $1;

-- name: AtualizarAvatarUpload :execrows
UPDATE usuarios
SET avatar_url = $2,
    avatar_jogo_id = NULL
WHERE id = $1;

-- name: BuscarJogoZeradoParaCapa :one
SELECT id, igdb_capa_url
FROM jogos_zerados
WHERE id = $1 AND usuario_id = $2 AND deleted_at IS NULL;

-- name: AtualizarAvatarCapa :execrows
UPDATE usuarios
SET avatar_url = NULL,
    avatar_jogo_id = $2
WHERE id = $1;

-- name: RemoverAvatar :execrows
UPDATE usuarios
SET avatar_url = NULL,
    avatar_jogo_id = NULL
WHERE id = $1;

