ALTER TABLE imovel_fotos
    ADD COLUMN IF NOT EXISTS url_thumbnail TEXT NULL,
    ADD COLUMN IF NOT EXISTS chave_objeto VARCHAR(1024) NULL,
    ADD COLUMN IF NOT EXISTS chave_thumbnail VARCHAR(1024) NULL,
    ADD COLUMN IF NOT EXISTS provedor_storage VARCHAR(30) NULL,
    ADD COLUMN IF NOT EXISTS tipo_conteudo VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS tamanho_bytes BIGINT NULL,
    ADD COLUMN IF NOT EXISTS largura INTEGER NULL,
    ADD COLUMN IF NOT EXISTS altura INTEGER NULL,
    ADD COLUMN IF NOT EXISTS hash_sha256 CHAR(64) NULL;

ALTER TABLE imovel_fotos
    ADD CONSTRAINT chk_imovel_fotos_tamanho_positivo
        CHECK (tamanho_bytes IS NULL OR tamanho_bytes > 0),
    ADD CONSTRAINT chk_imovel_fotos_dimensoes_positivas
        CHECK ((largura IS NULL AND altura IS NULL) OR (largura > 0 AND altura > 0)),
    ADD CONSTRAINT chk_imovel_fotos_storage_coerente
        CHECK (
            (chave_objeto IS NULL AND provedor_storage IS NULL)
            OR (chave_objeto IS NOT NULL AND provedor_storage IS NOT NULL)
        );

CREATE INDEX IF NOT EXISTS idx_imovel_fotos_hash
    ON imovel_fotos (imovel_id, hash_sha256)
    WHERE hash_sha256 IS NOT NULL;
