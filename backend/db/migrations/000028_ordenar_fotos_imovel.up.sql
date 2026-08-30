WITH capas_duplicadas AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY imovel_id ORDER BY ordem, criado_em, id) AS posicao
    FROM imovel_fotos
    WHERE is_capa = TRUE
)
UPDATE imovel_fotos f
SET is_capa = FALSE
FROM capas_duplicadas d
WHERE f.id = d.id AND d.posicao > 1;

UPDATE imovel_fotos f
SET is_capa = TRUE
WHERE f.id IN (
    SELECT DISTINCT ON (imovel_id) id
    FROM imovel_fotos origem
    WHERE NOT EXISTS (
        SELECT 1 FROM imovel_fotos capa
        WHERE capa.imovel_id = origem.imovel_id AND capa.is_capa = TRUE
    )
    ORDER BY imovel_id, ordem, criado_em, id
);

CREATE UNIQUE INDEX idx_imovel_fotos_capa_unica
    ON imovel_fotos (imovel_id)
    WHERE is_capa = TRUE;

CREATE INDEX idx_imovel_fotos_ordem
    ON imovel_fotos (imovel_id, ordem, criado_em);
