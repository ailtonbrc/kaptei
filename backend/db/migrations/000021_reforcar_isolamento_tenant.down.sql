DROP INDEX IF EXISTS idx_agendamentos_conta_periodo;
DROP INDEX IF EXISTS idx_leads_conta_status_criado;
DROP INDEX IF EXISTS idx_clientes_conta_corretor;

ALTER TABLE agendamentos
    DROP CONSTRAINT IF EXISTS agendamentos_imovel_mesma_conta_fk,
    DROP CONSTRAINT IF EXISTS agendamentos_cliente_mesma_conta_fk,
    DROP CONSTRAINT IF EXISTS agendamentos_usuario_mesma_conta_fk;

ALTER TABLE interacoes
    DROP CONSTRAINT IF EXISTS interacoes_corretor_mesma_conta_fk,
    DROP CONSTRAINT IF EXISTS interacoes_cliente_mesma_conta_fk;

ALTER TABLE leads
    DROP CONSTRAINT IF EXISTS leads_imovel_mesma_conta_fk,
    DROP CONSTRAINT IF EXISTS leads_usuario_mesma_conta_fk;

ALTER TABLE clientes DROP CONSTRAINT IF EXISTS clientes_corretor_mesma_conta_fk;
ALTER TABLE imoveis DROP CONSTRAINT IF EXISTS imoveis_usuario_mesma_conta_fk;

ALTER TABLE imoveis DROP CONSTRAINT IF EXISTS imoveis_id_conta_unique;
ALTER TABLE clientes DROP CONSTRAINT IF EXISTS clientes_id_conta_unique;
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_id_conta_unique;
