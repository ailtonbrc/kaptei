DROP INDEX IF EXISTS idx_imovel_fotos_hash;

ALTER TABLE imovel_fotos
    DROP CONSTRAINT IF EXISTS chk_imovel_fotos_storage_coerente,
    DROP CONSTRAINT IF EXISTS chk_imovel_fotos_dimensoes_positivas,
    DROP CONSTRAINT IF EXISTS chk_imovel_fotos_tamanho_positivo,
    DROP COLUMN IF EXISTS hash_sha256,
    DROP COLUMN IF EXISTS altura,
    DROP COLUMN IF EXISTS largura,
    DROP COLUMN IF EXISTS tamanho_bytes,
    DROP COLUMN IF EXISTS tipo_conteudo,
    DROP COLUMN IF EXISTS provedor_storage,
    DROP COLUMN IF EXISTS chave_thumbnail,
    DROP COLUMN IF EXISTS chave_objeto,
    DROP COLUMN IF EXISTS url_thumbnail;
