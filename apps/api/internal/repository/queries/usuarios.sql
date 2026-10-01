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
