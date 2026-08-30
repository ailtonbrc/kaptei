DELETE FROM configuracoes_sistema WHERE chave = 'TRIAL_DIAS_PADRAO';

ALTER TABLE contas_saas DROP COLUMN gateway_customer_id;
ALTER TABLE contas_saas DROP COLUMN gateway_subscription_id;
