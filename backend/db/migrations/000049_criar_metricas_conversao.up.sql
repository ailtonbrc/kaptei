CREATE TABLE eventos_conversao_site (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    chave_evento UUID NOT NULL,
    sessao_id UUID NOT NULL,
    tipo VARCHAR(40) NOT NULL,
    imovel_id UUID,
    utm_source VARCHAR(120),
    utm_medium VARCHAR(120),
    utm_campaign VARCHAR(180),
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expira_em TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '13 months'),
    CONSTRAINT eventos_conversao_tipo_ck CHECK (tipo IN (
        'SITE_VISUALIZADO',
        'IMOVEL_VISUALIZADO',
        'FORMULARIO_INICIADO',
        'LEAD_ENVIADO',
        'WHATSAPP_CLICADO',
        'TELEFONE_CLICADO'
    )),
    CONSTRAINT eventos_conversao_imovel_conta_fk
        FOREIGN KEY (imovel_id, conta_id) REFERENCES imoveis(id, conta_id),
    CONSTRAINT eventos_conversao_expiracao_ck CHECK (expira_em > criado_em),
    CONSTRAINT eventos_conversao_chave_unica UNIQUE (conta_id, chave_evento)
);

CREATE INDEX idx_eventos_conversao_conta_criado
    ON eventos_conversao_site (conta_id, criado_em DESC);

CREATE INDEX idx_eventos_conversao_conta_tipo_criado
    ON eventos_conversao_site (conta_id, tipo, criado_em DESC);

CREATE INDEX idx_eventos_conversao_expiracao
    ON eventos_conversao_site (expira_em);
CREATE UNIQUE INDEX idx_eventos_conversao_etapa_sessao_unica
    ON eventos_conversao_site (
        conta_id, sessao_id, tipo, COALESCE(imovel_id, '00000000-0000-0000-0000-000000000000'::uuid)
    );
