-- name: RevogarToken :exec
INSERT INTO tokens_revogados (jti, expira_em)
VALUES ($1, $2)
ON CONFLICT (jti) DO NOTHING;

-- name: TokenEstaRevogado :one
SELECT EXISTS(SELECT 1 FROM tokens_revogados WHERE jti = $1);

-- name: LimparTokensRevogadosExpirados :exec
DELETE FROM tokens_revogados WHERE expira_em <= CURRENT_TIMESTAMP;
