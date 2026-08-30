-- Adiciona campos de Gateway de Pagamento na conta
ALTER TABLE contas_saas ADD COLUMN gateway_customer_id VARCHAR(100);
ALTER TABLE contas_saas ADD COLUMN gateway_subscription_id VARCHAR(100);

-- Adiciona a configuração de Trial Days Padrão
INSERT INTO configuracoes_sistema (chave, valor, descricao) 
VALUES ('TRIAL_DIAS_PADRAO', '{"dias": 14}', 'Quantidade de dias do período de Trial gratuito.')
ON CONFLICT (chave) DO NOTHING;
