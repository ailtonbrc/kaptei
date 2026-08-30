CREATE TABLE integracoes_portais (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    portal varchar(30) NOT NULL,
    ativa boolean NOT NULL DEFAULT false,
    token_feed_hash char(64),
    token_feed_prefixo varchar(12),
    nome_contato varchar(120) NOT NULL DEFAULT '',
    email_contato varchar(254) NOT NULL DEFAULT '',
    telefone_contato varchar(32) NOT NULL DEFAULT '',
    exibicao_endereco varchar(20) NOT NULL DEFAULT 'BAIRRO',
    atualizado_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integracoes_portais_conta_portal_unique UNIQUE (conta_id, portal),
    CONSTRAINT integracoes_portais_token_unique UNIQUE (token_feed_hash),
    CONSTRAINT integracoes_portais_portal_ck CHECK (portal IN ('GRUPO_OLX')),
    CONSTRAINT integracoes_portais_endereco_ck CHECK (exibicao_endereco IN ('BAIRRO','LOGRADOURO','COMPLETO')),
    CONSTRAINT integracoes_portais_token_ck CHECK (
        (token_feed_hash IS NULL AND token_feed_prefixo IS NULL)
        OR (token_feed_hash IS NOT NULL AND token_feed_prefixo IS NOT NULL)
    )
);

CREATE TABLE publicacoes_portais (
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    portal varchar(30) NOT NULL,
    imovel_id uuid NOT NULL REFERENCES imoveis(id) ON DELETE CASCADE,
    ativa boolean NOT NULL DEFAULT false,
    tipo_publicacao varchar(20) NOT NULL DEFAULT 'STANDARD',
    atualizado_por uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (conta_id, portal, imovel_id),
    CONSTRAINT publicacoes_portais_portal_ck CHECK (portal IN ('GRUPO_OLX')),
    CONSTRAINT publicacoes_portais_tipo_ck CHECK (tipo_publicacao IN ('STANDARD','PREMIUM','SUPER_PREMIUM'))
);

CREATE INDEX idx_publicacoes_portais_feed
    ON publicacoes_portais (conta_id, portal, ativa, imovel_id);
