DROP INDEX IF EXISTS idx_leads_cliente_unico;
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_cliente_mesma_conta_fk;
ALTER TABLE leads DROP COLUMN IF EXISTS cliente_id;
