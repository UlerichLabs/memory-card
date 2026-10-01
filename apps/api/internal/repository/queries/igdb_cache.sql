-- name: BuscarSnapshotIGDB :one
SELECT payload FROM igdb_cache WHERE chave = $1;

-- name: SalvarSnapshotIGDB :exec
INSERT INTO igdb_cache (chave, payload, atualizado_em)
VALUES ($1, $2, now())
ON CONFLICT (chave) DO UPDATE SET payload = EXCLUDED.payload, atualizado_em = now();
