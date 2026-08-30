CREATE TABLE politicas_retencao (
    conta_id uuid PRIMARY KEY REFERENCES contas_saas(id) ON DELETE CASCADE,
    ativa boolean NOT NULL DEFAULT false,
    dias_leads_descartados integer NOT NULL DEFAULT 730,
    dias_clientes_perdidos integer NOT NULL DEFAULT 1825,
    tamanho_lote integer NOT NULL DEFAULT 200,
    fundamento_legal varchar(2000) NOT NULL DEFAULT '',
    ultima_execucao_em timestamptz,
    atualizado_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT politicas_retencao_dias_ck CHECK (dias_leads_descartados BETWEEN 30 AND 3650 AND dias_clientes_perdidos BETWEEN 30 AND 3650),
    CONSTRAINT politicas_retencao_lote_ck CHECK (tamanho_lote BETWEEN 1 AND 1000)
);

CREATE TABLE bloqueios_retencao (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    tipo_recurso varchar(12) NOT NULL,
    recurso_id uuid NOT NULL,
    motivo varchar(1000) NOT NULL,
    valido_ate timestamptz,
    criado_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT bloqueios_retencao_tipo_ck CHECK (tipo_recurso IN ('LEAD','CLIENTE')),
    CONSTRAINT bloqueios_retencao_recurso_unique UNIQUE (conta_id,tipo_recurso,recurso_id)
);

CREATE INDEX idx_bloqueios_retencao_vigentes ON bloqueios_retencao (conta_id,tipo_recurso,recurso_id,valido_ate);

CREATE TABLE execucoes_retencao (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    usuario_id uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    leads_anonimizados integer NOT NULL DEFAULT 0,
    clientes_anonimizados integer NOT NULL DEFAULT 0,
    fundamento_legal varchar(2000) NOT NULL,
    executado_em timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_execucoes_retencao_conta_data ON execucoes_retencao (conta_id,executado_em DESC);

