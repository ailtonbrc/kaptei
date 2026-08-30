ALTER TABLE contas_saas
    ADD COLUMN slug_publico VARCHAR(80),
    ADD COLUMN site_publicado BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN site_config JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX idx_contas_slug_publico_unico
    ON contas_saas (LOWER(slug_publico))
    WHERE slug_publico IS NOT NULL;

ALTER TABLE imoveis
    ADD COLUMN slug_publico VARCHAR(180),
    ADD COLUMN publicado BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN destaque BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN titulo_seo VARCHAR(180),
    ADD COLUMN descricao_seo VARCHAR(320);

CREATE UNIQUE INDEX idx_imoveis_slug_publico_unico
    ON imoveis (conta_id, LOWER(slug_publico))
    WHERE slug_publico IS NOT NULL;

CREATE INDEX idx_imoveis_catalogo_publico
    ON imoveis (conta_id, destaque DESC, atualizado_em DESC)
    WHERE publicado = TRUE AND status = 'Ativo';

ALTER TABLE leads
    ADD COLUMN pagina_origem TEXT,
    ADD COLUMN utm_source VARCHAR(120),
    ADD COLUMN utm_medium VARCHAR(120),
    ADD COLUMN utm_campaign VARCHAR(180),
    ADD COLUMN consentimento_lgpd BOOLEAN NOT NULL DEFAULT FALSE;
