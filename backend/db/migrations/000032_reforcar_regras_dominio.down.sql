DROP INDEX IF EXISTS idx_usuarios_email_normalizado_unico;

ALTER TABLE usuarios
    DROP CONSTRAINT IF EXISTS usuarios_conta_obrigatoria_ck,
    DROP CONSTRAINT IF EXISTS usuarios_papel_valido_ck,
    DROP CONSTRAINT IF EXISTS usuarios_status_valido_ck;

ALTER TABLE contas_saas
    DROP CONSTRAINT IF EXISTS contas_tipo_valido_ck,
    DROP CONSTRAINT IF EXISTS contas_status_plano_valido_ck,
    DROP CONSTRAINT IF EXISTS contas_lead_estrategia_valida_ck;

ALTER TABLE imoveis
    DROP CONSTRAINT IF EXISTS imoveis_tipo_valido_ck,
    DROP CONSTRAINT IF EXISTS imoveis_finalidade_valida_ck,
    DROP CONSTRAINT IF EXISTS imoveis_status_valido_ck,
    DROP CONSTRAINT IF EXISTS imoveis_valores_positivos_ck;

ALTER TABLE clientes
    DROP CONSTRAINT IF EXISTS clientes_status_funil_valido_ck,
    DROP CONSTRAINT IF EXISTS clientes_temperatura_valida_ck;

ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_status_valido_ck;

ALTER TABLE agendamentos
    DROP CONSTRAINT IF EXISTS agendamentos_status_valido_ck,
    DROP CONSTRAINT IF EXISTS agendamentos_tipo_valido_ck;
