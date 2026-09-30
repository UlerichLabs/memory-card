CREATE TABLE jogos_abandonados (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
    igdb_id INTEGER,
    igdb_capa_url TEXT,
    nome VARCHAR(200) NOT NULL,
    console VARCHAR(100) NOT NULL,
    tempo_jogado INTEGER NOT NULL DEFAULT 0 CHECK (tempo_jogado >= 0),
    motivo TEXT CHECK (motivo IS NULL OR char_length(motivo) <= 500),
    abandonado_em TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_jogos_abandonados_usuario_data
    ON jogos_abandonados (usuario_id, abandonado_em DESC, id DESC)
    WHERE deleted_at IS NULL;
