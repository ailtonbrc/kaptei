-- Normaliza somente equivalências legadas conhecidas. Valores desconhecidos
-- devem interromper a validação para exigir uma decisão explícita do operador.
UPDATE contas_saas
SET status_plano = 'ATIVO'
WHERE UPPER(status_plano) = 'ACTIVE';

ALTER TABLE usuarios
    VALIDATE CONSTRAINT usuarios_conta_obrigatoria_ck;
ALTER TABLE usuarios
    VALIDATE CONSTRAINT usuarios_papel_valido_ck;
ALTER TABLE usuarios
    VALIDATE CONSTRAINT usuarios_status_valido_ck;

ALTER TABLE contas_saas
    VALIDATE CONSTRAINT contas_tipo_valido_ck;
ALTER TABLE contas_saas
    VALIDATE CONSTRAINT contas_status_plano_valido_ck;
ALTER TABLE contas_saas
    VALIDATE CONSTRAINT contas_lead_estrategia_valida_ck;

ALTER TABLE imoveis
    VALIDATE CONSTRAINT imoveis_tipo_valido_ck;
ALTER TABLE imoveis
    VALIDATE CONSTRAINT imoveis_finalidade_valida_ck;
ALTER TABLE imoveis
    VALIDATE CONSTRAINT imoveis_status_valido_ck;
ALTER TABLE imoveis
    VALIDATE CONSTRAINT imoveis_valores_positivos_ck;

ALTER TABLE clientes
    VALIDATE CONSTRAINT clientes_status_funil_valido_ck;
ALTER TABLE clientes
    VALIDATE CONSTRAINT clientes_temperatura_valida_ck;

ALTER TABLE leads
    VALIDATE CONSTRAINT leads_status_valido_ck;

ALTER TABLE agendamentos
    VALIDATE CONSTRAINT agendamentos_status_valido_ck;
ALTER TABLE agendamentos
    VALIDATE CONSTRAINT agendamentos_tipo_valido_ck;
