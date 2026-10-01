CREATE TABLE jogos_em_andamento (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
    igdb_id INTEGER,
    igdb_capa_url TEXT,
    nome VARCHAR(200) NOT NULL,
    iniciado_em TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_jogos_em_andamento_usuario_iniciado
ON jogos_em_andamento (usuario_id, iniciado_em DESC, id DESC)
WHERE deleted_at IS NULL;
