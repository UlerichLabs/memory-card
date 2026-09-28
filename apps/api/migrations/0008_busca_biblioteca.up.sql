CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE INDEX idx_jogos_usuario_finalizado ON jogos_zerados (usuario_id, finalizado_em DESC, id DESC) WHERE deleted_at IS NULL;
