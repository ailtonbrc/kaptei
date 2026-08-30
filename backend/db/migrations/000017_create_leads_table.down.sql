DROP TABLE IF EXISTS leads CASCADE;

ALTER TABLE contas_saas DROP CONSTRAINT IF EXISTS contas_saas_lead_token_unique;
ALTER TABLE contas_saas DROP COLUMN IF EXISTS lead_estrategia;
ALTER TABLE contas_saas DROP COLUMN IF EXISTS lead_token_integracao;
