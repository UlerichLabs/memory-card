ALTER TABLE lista_itens ADD COLUMN ignorado BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_lista_itens_lista_ignorado ON lista_itens (lista_id) WHERE ignorado;
