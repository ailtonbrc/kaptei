DROP TABLE IF EXISTS billing_eventos;
DROP INDEX IF EXISTS idx_contas_billing_subscription_unico;
DROP INDEX IF EXISTS idx_contas_billing_customer_unico;
ALTER TABLE contas_saas DROP COLUMN IF EXISTS billing_periodo_fim, DROP COLUMN IF EXISTS billing_status, DROP COLUMN IF EXISTS billing_subscription_id, DROP COLUMN IF EXISTS billing_customer_id;
ALTER TABLE planos DROP COLUMN IF EXISTS gateway_price_id;
