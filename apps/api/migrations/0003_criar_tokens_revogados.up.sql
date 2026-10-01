CREATE TABLE tokens_revogados (
    jti TEXT PRIMARY KEY,
    expira_em TIMESTAMPTZ NOT NULL
);

CREATE INDEX tokens_revogados_expira_em_idx ON tokens_revogados (expira_em);
