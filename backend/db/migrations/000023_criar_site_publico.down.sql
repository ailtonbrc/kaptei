ALTER TABLE leads
    DROP COLUMN IF EXISTS consentimento_lgpd,
    DROP COLUMN IF EXISTS utm_campaign,
    DROP COLUMN IF EXISTS utm_medium,
    DROP COLUMN IF EXISTS utm_source,
    DROP COLUMN IF EXISTS pagina_origem;

DROP INDEX IF EXISTS idx_imoveis_catalogo_publico;
DROP INDEX IF EXISTS idx_imoveis_slug_publico_unico;

ALTER TABLE imoveis
    DROP COLUMN IF EXISTS descricao_seo,
    DROP COLUMN IF EXISTS titulo_seo,
    DROP COLUMN IF EXISTS destaque,
    DROP COLUMN IF EXISTS publicado,
    DROP COLUMN IF EXISTS slug_publico;

DROP INDEX IF EXISTS idx_contas_slug_publico_unico;

ALTER TABLE contas_saas
    DROP COLUMN IF EXISTS site_config,
    DROP COLUMN IF EXISTS site_publicado,
    DROP COLUMN IF EXISTS slug_publico;
