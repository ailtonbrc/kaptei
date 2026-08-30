DROP INDEX IF EXISTS idx_leads_captura_idempotente;
ALTER TABLE leads DROP COLUMN IF EXISTS chave_idempotencia;
