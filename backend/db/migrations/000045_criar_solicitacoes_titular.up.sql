CREATE TABLE solicitacoes_titular (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    protocolo varchar(40) NOT NULL UNIQUE,
    tipo varchar(40) NOT NULL,
    nome_protegido text NOT NULL,
    email_hash char(64),
    telefone_hash char(64),
    contato_protegido text NOT NULL,
    detalhes_protegidos text,
    status varchar(24) NOT NULL DEFAULT 'RECEBIDA',
    prazo_resposta_em timestamptz NOT NULL,
    identidade_verificada_em timestamptz,
    identidade_verificada_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    metodo_verificacao varchar(80),
    evidencia_verificacao_protegida text,
    decisao varchar(16),
    fundamento_legal text,
    observacao_decisao text,
    decidida_em timestamptz,
    decidida_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    executada_em timestamptz,
    executada_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT solicitacoes_titular_tipo_ck CHECK (tipo IN (
        'CONFIRMACAO','ACESSO','CORRECAO','ANONIMIZACAO','BLOQUEIO',
        'EXCLUSAO','PORTABILIDADE','REVOGACAO','INFORMACAO_COMPARTILHAMENTO'
    )),
    CONSTRAINT solicitacoes_titular_status_ck CHECK (status IN (
        'RECEBIDA','VALIDANDO_IDENTIDADE','EM_ANALISE','APROVADA','REJEITADA','CONCLUIDA'
    )),
    CONSTRAINT solicitacoes_titular_decisao_ck CHECK (decisao IS NULL OR decisao IN ('APROVADA','REJEITADA')),
    CONSTRAINT solicitacoes_titular_contato_ck CHECK (email_hash IS NOT NULL OR telefone_hash IS NOT NULL)
);

CREATE INDEX idx_solicitacoes_titular_conta_criacao
    ON solicitacoes_titular (conta_id, criado_em DESC);
CREATE INDEX idx_solicitacoes_titular_conta_status_prazo
    ON solicitacoes_titular (conta_id, status, prazo_resposta_em);
CREATE INDEX idx_solicitacoes_titular_email
    ON solicitacoes_titular (conta_id, email_hash) WHERE email_hash IS NOT NULL;
CREATE INDEX idx_solicitacoes_titular_telefone
    ON solicitacoes_titular (conta_id, telefone_hash) WHERE telefone_hash IS NOT NULL;

CREATE TABLE eventos_solicitacao_titular (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    solicitacao_id uuid NOT NULL REFERENCES solicitacoes_titular(id) ON DELETE CASCADE,
    usuario_id uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    tipo varchar(40) NOT NULL,
    descricao varchar(500) NOT NULL,
    criado_em timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_eventos_solicitacao_titular_solicitacao
    ON eventos_solicitacao_titular (solicitacao_id, criado_em);

