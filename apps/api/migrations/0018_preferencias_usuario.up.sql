CREATE TABLE preferencias_usuario (
    usuario_id          INTEGER PRIMARY KEY REFERENCES usuarios(id) ON DELETE CASCADE,
    formato_data        VARCHAR(10) NOT NULL DEFAULT 'dmy' CHECK (formato_data IN ('dmy', 'mdy', 'ymd')),
    fuso_horario        VARCHAR(100) NOT NULL DEFAULT 'America/Sao_Paulo',
    visual_biblioteca   VARCHAR(10) NOT NULL DEFAULT 'grade' CHECK (visual_biblioteca IN ('grade', 'lista')),
    itens_por_pagina    INTEGER NOT NULL DEFAULT 24 CHECK (itens_por_pagina IN (24, 50, 100)),
    reduzir_animacoes   BOOLEAN NOT NULL DEFAULT false,
    modo_tema           VARCHAR(20) NOT NULL DEFAULT 'escuro' CHECK (modo_tema IN ('claro', 'escuro', 'automatico')),
    estilo_tema         VARCHAR(20) NOT NULL DEFAULT 'padrao' CHECK (estilo_tema IN ('padrao', 'playstation', 'nintendo', 'xbox', 'steam')),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
