-- PostgreSQL não permite transformar uma constraint validada em NOT VALID.
-- A reversão precisa recriá-las, preservando a proteção para novas gravações.
-- A normalização ACTIVE -> ATIVO não é revertida para não reintroduzir dívida.
ALTER TABLE usuarios
    DROP CONSTRAINT usuarios_conta_obrigatoria_ck,
    DROP CONSTRAINT usuarios_papel_valido_ck,
    DROP CONSTRAINT usuarios_status_valido_ck,
    ADD CONSTRAINT usuarios_conta_obrigatoria_ck CHECK (conta_id IS NOT NULL) NOT VALID,
    ADD CONSTRAINT usuarios_papel_valido_ck CHECK (UPPER(papel) IN ('SUPER_ADMIN','GESTOR','CORRETOR_EQUIPE','CORRETOR_SOLO','VISITANTE')) NOT VALID,
    ADD CONSTRAINT usuarios_status_valido_ck CHECK (UPPER(status) IN ('ATIVO','INATIVO')) NOT VALID;

ALTER TABLE contas_saas
    DROP CONSTRAINT contas_tipo_valido_ck,
    DROP CONSTRAINT contas_status_plano_valido_ck,
    DROP CONSTRAINT contas_lead_estrategia_valida_ck,
    ADD CONSTRAINT contas_tipo_valido_ck CHECK (tipo_conta IN ('CORRETOR_SOLO','IMOBILIARIA')) NOT VALID,
    ADD CONSTRAINT contas_status_plano_valido_ck CHECK (status_plano IN ('TRIAL','AGUARDANDO_PAGAMENTO','ATIVO','INADIMPLENTE','CANCELADO','GRATUITO')) NOT VALID,
    ADD CONSTRAINT contas_lead_estrategia_valida_ck CHECK (lead_estrategia IN ('CAIXA_ENTRADA','ROLETA')) NOT VALID;

ALTER TABLE imoveis
    DROP CONSTRAINT imoveis_tipo_valido_ck,
    DROP CONSTRAINT imoveis_finalidade_valida_ck,
    DROP CONSTRAINT imoveis_status_valido_ck,
    DROP CONSTRAINT imoveis_valores_positivos_ck,
    ADD CONSTRAINT imoveis_tipo_valido_ck CHECK (tipo IN ('Casa','Apartamento','Terreno','Comercial','Galpão','Rural')) NOT VALID,
    ADD CONSTRAINT imoveis_finalidade_valida_ck CHECK (finalidade IN ('Venda','Locação','Venda e Locação')) NOT VALID,
    ADD CONSTRAINT imoveis_status_valido_ck CHECK (status IN ('Ativo','Inativo','Vendido','Alugado')) NOT VALID,
    ADD CONSTRAINT imoveis_valores_positivos_ck CHECK (
        COALESCE(valor_venda,0) >= 0 AND COALESCE(valor_locacao,0) >= 0 AND
        COALESCE(valor_condominio,0) >= 0 AND COALESCE(valor_iptu,0) >= 0 AND
        COALESCE(area_total,0) >= 0 AND COALESCE(area_util,0) >= 0 AND
        quartos >= 0 AND suites >= 0 AND banheiros >= 0 AND vagas >= 0
    ) NOT VALID;

ALTER TABLE clientes
    DROP CONSTRAINT clientes_status_funil_valido_ck,
    DROP CONSTRAINT clientes_temperatura_valida_ck,
    ADD CONSTRAINT clientes_status_funil_valido_ck CHECK (status_funil IN ('NOVO','ATENDIMENTO','VISITA','PROPOSTA','FECHADO','PERDIDO')) NOT VALID,
    ADD CONSTRAINT clientes_temperatura_valida_ck CHECK (temperatura IS NULL OR temperatura IN ('FRIO','MORNO','QUENTE')) NOT VALID;

ALTER TABLE leads
    DROP CONSTRAINT leads_status_valido_ck,
    ADD CONSTRAINT leads_status_valido_ck CHECK (status IN ('NOVO','EM_ATENDIMENTO','QUALIFICADO','DESCARTADO')) NOT VALID;

ALTER TABLE agendamentos
    DROP CONSTRAINT agendamentos_status_valido_ck,
    DROP CONSTRAINT agendamentos_tipo_valido_ck,
    ADD CONSTRAINT agendamentos_status_valido_ck CHECK (status IN ('AGENDADO','CONFIRMADO','CONCLUIDO','CANCELADO','NAO_COMPARECEU')) NOT VALID,
    ADD CONSTRAINT agendamentos_tipo_valido_ck CHECK (tipo IN ('VISITA','LIGACAO','REUNIAO_ONLINE','TAREFA')) NOT VALID;
