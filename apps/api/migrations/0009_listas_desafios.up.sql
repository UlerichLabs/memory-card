CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TYPE lista_tipo AS ENUM ('fila', 'desafio');
CREATE TYPE lista_regra_tipo AS ENUM ('franquia', 'plataforma', 'genero', 'manual');

CREATE TABLE listas (
    id BIGSERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    tipo lista_tipo NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao VARCHAR(200) NULL,
    regra_tipo lista_regra_tipo NULL,
    regra_valor VARCHAR(150) NULL,
    regra_igdb_id INTEGER NULL,
    meta INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_listas_tipo_regras CHECK (
        (tipo = 'fila' AND regra_tipo IS NULL AND regra_valor IS NULL AND regra_igdb_id IS NULL AND meta IS NULL)
        OR
        (tipo = 'desafio' AND regra_tipo IS NOT NULL)
    ),
    CONSTRAINT chk_listas_meta CHECK (
        meta IS NULL OR (meta >= 1 AND meta <= 10000)
    )
);

CREATE INDEX idx_listas_usuario_created ON listas (usuario_id, created_at DESC);

CREATE TABLE lista_itens (
    id BIGSERIAL PRIMARY KEY,
    lista_id BIGINT NOT NULL REFERENCES listas(id) ON DELETE CASCADE,
    igdb_id INTEGER NULL,
    nome VARCHAR(200) NOT NULL,
    console VARCHAR(100) NULL,
    igdb_capa_url TEXT NULL,
    ano_lancamento INTEGER NULL,
    posicao INTEGER NOT NULL,
    jogo_zerado_id INTEGER NULL REFERENCES jogos_zerados(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_lista_itens_lista_igdb ON lista_itens (lista_id, igdb_id) WHERE igdb_id IS NOT NULL;
CREATE INDEX idx_lista_itens_lista_posicao ON lista_itens (lista_id, posicao);
