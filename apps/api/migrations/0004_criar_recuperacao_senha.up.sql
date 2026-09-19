CREATE TABLE tokens_reset_senha (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
    token_hash CHAR(64) NOT NULL UNIQUE,
    expira_em TIMESTAMPTZ NOT NULL,
    usado_em TIMESTAMPTZ
);

CREATE INDEX tokens_reset_senha_token_hash_idx ON tokens_reset_senha (token_hash);
CREATE INDEX tokens_reset_senha_usuario_id_idx ON tokens_reset_senha (usuario_id);

CREATE TABLE limites_solicitacao_reset_senha (
    email VARCHAR(200) PRIMARY KEY,
    janela_iniciada_em TIMESTAMPTZ NOT NULL,
    quantidade SMALLINT NOT NULL CHECK (quantidade > 0)
);

CREATE TABLE tokens_refresh_ativos (
    jti TEXT PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
    expira_em TIMESTAMPTZ NOT NULL
);

CREATE INDEX tokens_refresh_ativos_usuario_id_idx ON tokens_refresh_ativos (usuario_id);
