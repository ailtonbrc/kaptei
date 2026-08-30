CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Constraints NOT VALID da migration 32 ja validam qualquer nova versao da
-- linha. Normalize o valor legado conhecido antes de atualizar o token; a
-- migration 50 continua responsavel por validar integralmente o dominio.
UPDATE contas_saas
SET status_plano = 'ATIVO'
WHERE status_plano = 'ACTIVE';

ALTER TABLE contas_saas
    ADD COLUMN IF NOT EXISTS lead_token_hash CHAR(64),
    ADD COLUMN IF NOT EXISTS lead_token_prefixo VARCHAR(8);

UPDATE contas_saas
SET lead_token_hash = encode(digest(lead_token_integracao::text, 'sha256'), 'hex'),
    lead_token_prefixo = LEFT(lead_token_integracao::text, 8)
WHERE lead_token_integracao IS NOT NULL
  AND lead_token_hash IS NULL;

ALTER TABLE contas_saas
    ALTER COLUMN lead_token_integracao DROP DEFAULT;

CREATE UNIQUE INDEX IF NOT EXISTS ux_contas_saas_lead_token_hash
    ON contas_saas (lead_token_hash)
    WHERE lead_token_hash IS NOT NULL;
