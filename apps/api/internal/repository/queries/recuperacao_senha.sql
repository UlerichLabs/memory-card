-- name: RegistrarSolicitacaoResetSenha :one
INSERT INTO limites_solicitacao_reset_senha (email, janela_iniciada_em, quantidade)
VALUES ($1, CURRENT_TIMESTAMP, 1)
ON CONFLICT (email) DO UPDATE
SET
    janela_iniciada_em = CASE
        WHEN limites_solicitacao_reset_senha.janela_iniciada_em <= CURRENT_TIMESTAMP - INTERVAL '1 hour' THEN CURRENT_TIMESTAMP
        ELSE limites_solicitacao_reset_senha.janela_iniciada_em
    END,
    quantidade = CASE
        WHEN limites_solicitacao_reset_senha.janela_iniciada_em <= CURRENT_TIMESTAMP - INTERVAL '1 hour' THEN 1
        ELSE limites_solicitacao_reset_senha.quantidade + 1
    END
WHERE
    limites_solicitacao_reset_senha.janela_iniciada_em <= CURRENT_TIMESTAMP - INTERVAL '1 hour'
    OR limites_solicitacao_reset_senha.quantidade < 3
RETURNING quantidade;

-- name: CriarTokenResetSenha :exec
INSERT INTO tokens_reset_senha (usuario_id, token_hash, expira_em)
VALUES ($1, $2, $3);

-- name: BuscarTokenResetSenha :one
SELECT id, usuario_id, token_hash, expira_em, usado_em
FROM tokens_reset_senha
WHERE token_hash = $1;

-- name: ConsumirTokenResetEAtualizarSenha :one
WITH token AS (
    UPDATE tokens_reset_senha
    SET usado_em = CURRENT_TIMESTAMP
    WHERE tokens_reset_senha.id = $1
        AND tokens_reset_senha.usado_em IS NULL
        AND tokens_reset_senha.expira_em > CURRENT_TIMESTAMP
    RETURNING usuario_id
), usuario AS (
    UPDATE usuarios
    SET senha_hash = $2
    WHERE usuarios.id = (SELECT usuario_id FROM token)
    RETURNING usuarios.id
)
SELECT usuario.id FROM usuario;

-- name: RegistrarRefreshTokenAtivo :exec
INSERT INTO tokens_refresh_ativos (jti, usuario_id, expira_em)
VALUES ($1, $2, $3)
ON CONFLICT (jti) DO NOTHING;

-- name: ListarRefreshTokensAtivosPorUsuario :many
SELECT jti, usuario_id, expira_em
FROM tokens_refresh_ativos
WHERE usuario_id = $1 AND expira_em > CURRENT_TIMESTAMP;

-- name: RemoverRefreshTokensAtivosPorUsuario :exec
DELETE FROM tokens_refresh_ativos WHERE usuario_id = $1;
