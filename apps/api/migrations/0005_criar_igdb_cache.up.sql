CREATE TABLE igdb_cache (
    chave TEXT PRIMARY KEY,
    payload JSONB NOT NULL,
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);
