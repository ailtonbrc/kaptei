CREATE TABLE integracoes_meta_leads (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    pagina_id varchar(64) NOT NULL,
    token_pagina_protegido text NOT NULL,
    ativa boolean NOT NULL DEFAULT true,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integracoes_meta_leads_conta_unique UNIQUE (conta_id),
    CONSTRAINT integracoes_meta_leads_pagina_unique UNIQUE (pagina_id),
    CONSTRAINT integracoes_meta_leads_pagina_formato CHECK (pagina_id ~ '^[0-9]{1,64}$')
);

CREATE TABLE eventos_integracao (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    provedor varchar(40) NOT NULL,
    tipo varchar(80) NOT NULL,
    identificador_externo varchar(255) NOT NULL,
    pagina_id varchar(64) NOT NULL,
    formulario_id varchar(64),
    anuncio_id varchar(64),
    ocorrido_em timestamptz,
    status varchar(20) NOT NULL DEFAULT 'PENDENTE',
    tentativas integer NOT NULL DEFAULT 0,
    maximo_tentativas integer NOT NULL DEFAULT 8,
    disponivel_em timestamptz NOT NULL DEFAULT now(),
    bloqueado_ate timestamptz,
    bloqueado_por varchar(255),
    ultimo_erro varchar(1000),
    criado_em timestamptz NOT NULL DEFAULT now(),
    processado_em timestamptz,
    CONSTRAINT eventos_integracao_idempotencia UNIQUE (provedor, identificador_externo),
    CONSTRAINT eventos_integracao_status_check CHECK (status IN ('PENDENTE', 'PROCESSANDO', 'CONCLUIDO', 'FALHOU')),
    CONSTRAINT eventos_integracao_tentativas_check CHECK (tentativas >= 0 AND maximo_tentativas > 0)
);

CREATE INDEX idx_eventos_integracao_processamento
    ON eventos_integracao (disponivel_em, criado_em)
    WHERE status IN ('PENDENTE', 'PROCESSANDO');

CREATE INDEX idx_eventos_integracao_conta
    ON eventos_integracao (conta_id, criado_em DESC);
