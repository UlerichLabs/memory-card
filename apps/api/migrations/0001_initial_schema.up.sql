CREATE TYPE dificuldade AS ENUM ('C', 'B', 'A', 'AA', 'AAA');
CREATE TYPE tipo_desafio AS ENUM ('QUANTIDADE', 'LISTA_COMPLETA', 'FRANQUIA', 'EMPRESA', 'GENERO', 'MANUAL');
CREATE TYPE tipo_favorito AS ENUM ('JOGO', 'GENERO', 'TIPO');

CREATE TABLE usuarios (
    id              SERIAL PRIMARY KEY,
    nome            VARCHAR(100) NOT NULL,
    email           VARCHAR(200) NOT NULL UNIQUE,
    senha_hash      VARCHAR(255) NOT NULL,
    username        VARCHAR(50) NOT NULL UNIQUE,
    avatar_url      TEXT,
    bio             TEXT,
    redes_sociais   JSONB DEFAULT '{}',
    created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE jogos_zerados (
    id                  SERIAL PRIMARY KEY,
    usuario_id          INTEGER NOT NULL REFERENCES usuarios(id),
    igdb_id             INTEGER,
    nome                VARCHAR(200) NOT NULL,
    console             VARCHAR(50) NOT NULL,
    genero              VARCHAR(50),
    tipo                VARCHAR(50),
    iniciado_em         TIMESTAMP,
    finalizado_em       TIMESTAMP NOT NULL,
    tempo_jogado        INTEGER NOT NULL,          -- segundos
    nota                INTEGER NOT NULL CHECK (nota BETWEEN 1 AND 11),
    dificuldade         dificuldade NOT NULL,
    condicao_zeramento  TEXT,
    destaque            BOOLEAN DEFAULT false,
    igdb_capa_url       TEXT,
    igdb_descricao      TEXT,
    deleted_at          TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE campanhas (
    id          SERIAL PRIMARY KEY,
    usuario_id  INTEGER NOT NULL REFERENCES usuarios(id),
    nome        VARCHAR(100) NOT NULL,
    descricao   TEXT,
    icone       VARCHAR(10),
    concluida   BOOLEAN DEFAULT false,
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE desafios (
    id            SERIAL PRIMARY KEY,
    usuario_id    INTEGER NOT NULL REFERENCES usuarios(id),
    campanha_id   INTEGER REFERENCES campanhas(id),
    nome          VARCHAR(100) NOT NULL,
    tipo          tipo_desafio NOT NULL,
    filtro_valor  VARCHAR(100),
    meta          INTEGER,
    concluido     BOOLEAN DEFAULT false,
    data_conclusao TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE desafio_itens (
    id            SERIAL PRIMARY KEY,
    desafio_id    INTEGER NOT NULL REFERENCES desafios(id),
    igdb_id       INTEGER,
    nome          VARCHAR(200) NOT NULL,
    concluido     BOOLEAN DEFAULT false,
    zeramento_id  INTEGER REFERENCES jogos_zerados(id),
    na_colecao    BOOLEAN DEFAULT false
);

CREATE TABLE conquistas (
    id              SERIAL PRIMARY KEY,
    usuario_id      INTEGER NOT NULL REFERENCES usuarios(id),
    desafio_id      INTEGER REFERENCES desafios(id),
    campanha_id     INTEGER REFERENCES campanhas(id),
    nome            VARCHAR(100) NOT NULL,
    descricao       TEXT,
    badge           VARCHAR(10),
    data_conquista  TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE favoritos (
    id          SERIAL PRIMARY KEY,
    usuario_id  INTEGER NOT NULL REFERENCES usuarios(id),
    jogo_id     INTEGER REFERENCES jogos_zerados(id),
    tipo        tipo_favorito NOT NULL,
    posicao     INTEGER NOT NULL CHECK (posicao BETWEEN 1 AND 10),
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_jogos_usuario ON jogos_zerados(usuario_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_jogos_console ON jogos_zerados(usuario_id, console) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_destaque_por_ano ON jogos_zerados(usuario_id, EXTRACT(YEAR FROM finalizado_em)) WHERE destaque = true AND deleted_at IS NULL;
