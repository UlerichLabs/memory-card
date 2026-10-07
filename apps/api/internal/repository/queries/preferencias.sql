-- name: BuscarPreferenciasPorUsuarioID :one
SELECT
    u.idioma,
    p.formato_data,
    p.fuso_horario,
    p.visual_biblioteca,
    p.itens_por_pagina,
    p.reduzir_animacoes,
    p.modo_tema,
    p.estilo_tema,
    p.updated_at
FROM usuarios u
LEFT JOIN preferencias_usuario p ON p.usuario_id = u.id
WHERE u.id = $1;

-- name: UpsertPreferenciasUsuario :one
INSERT INTO preferencias_usuario (
    usuario_id,
    formato_data,
    fuso_horario,
    visual_biblioteca,
    itens_por_pagina,
    reduzir_animacoes,
    modo_tema,
    estilo_tema,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, now()
)
ON CONFLICT (usuario_id) DO UPDATE SET
    formato_data = EXCLUDED.formato_data,
    fuso_horario = EXCLUDED.fuso_horario,
    visual_biblioteca = EXCLUDED.visual_biblioteca,
    itens_por_pagina = EXCLUDED.itens_por_pagina,
    reduzir_animacoes = EXCLUDED.reduzir_animacoes,
    modo_tema = EXCLUDED.modo_tema,
    estilo_tema = EXCLUDED.estilo_tema,
    updated_at = now()
RETURNING usuario_id, formato_data, fuso_horario, visual_biblioteca, itens_por_pagina, reduzir_animacoes, modo_tema, estilo_tema, updated_at;

-- name: AtualizarIdiomaUsuario :execrows
UPDATE usuarios
SET idioma = $2
WHERE id = $1;
