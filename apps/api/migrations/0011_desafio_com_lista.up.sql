DELETE FROM lista_itens WHERE ignorado;
DROP INDEX IF EXISTS idx_lista_itens_lista_ignorado;
ALTER TABLE lista_itens DROP COLUMN ignorado;
UPDATE listas SET meta = NULL;
