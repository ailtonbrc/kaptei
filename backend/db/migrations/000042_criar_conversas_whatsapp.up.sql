ALTER TABLE eventos_integracao ADD COLUMN payload_protegido text;

CREATE TABLE conversas_whatsapp (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    lead_id uuid REFERENCES leads(id) ON DELETE SET NULL,
    numero_contato varchar(32) NOT NULL,
    nome_contato varchar(120),
    consentimento_marketing boolean NOT NULL DEFAULT false,
    consentimento_marketing_em timestamptz,
    janela_atendimento_ate timestamptz,
    ultima_mensagem_em timestamptz NOT NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT conversas_whatsapp_contato_unique UNIQUE (conta_id, numero_contato),
    CONSTRAINT conversas_whatsapp_numero_formato CHECK (numero_contato ~ '^[0-9]{8,20}$')
);

CREATE TABLE mensagens_whatsapp (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    conversa_id uuid NOT NULL REFERENCES conversas_whatsapp(id) ON DELETE CASCADE,
    identificador_externo varchar(255) NOT NULL,
    direcao varchar(12) NOT NULL,
    tipo varchar(32) NOT NULL,
    conteudo_protegido text NOT NULL,
    status varchar(20) NOT NULL,
    ocorrida_em timestamptz NOT NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT mensagens_whatsapp_externo_unique UNIQUE (identificador_externo),
    CONSTRAINT mensagens_whatsapp_direcao_check CHECK (direcao IN ('ENTRADA', 'SAIDA')),
    CONSTRAINT mensagens_whatsapp_status_check CHECK (status IN ('RECEBIDA', 'PENDENTE', 'ENVIADA', 'ENTREGUE', 'LIDA', 'FALHOU'))
);

CREATE INDEX idx_conversas_whatsapp_conta_ultima
    ON conversas_whatsapp (conta_id, ultima_mensagem_em DESC);

CREATE INDEX idx_mensagens_whatsapp_conversa_data
    ON mensagens_whatsapp (conversa_id, ocorrida_em DESC);
