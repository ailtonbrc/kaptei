DROP INDEX IF EXISTS ux_contas_saas_lead_token_hash;

ALTER TABLE contas_saas
    ALTER COLUMN lead_token_integracao SET DEFAULT gen_random_uuid(),
    DROP COLUMN IF EXISTS lead_token_prefixo,
    DROP COLUMN IF EXISTS lead_token_hash;
