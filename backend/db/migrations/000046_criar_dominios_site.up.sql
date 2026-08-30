CREATE TABLE dominios_site (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    hostname varchar(253) NOT NULL,
    token_verificacao varchar(120) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'PENDENTE',
    verificado_em timestamptz,
    ultima_verificacao_em timestamptz,
    ultimo_erro varchar(500),
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT dominios_site_conta_unique UNIQUE (conta_id),
    CONSTRAINT dominios_site_status_ck CHECK (status IN ('PENDENTE','ATIVO','FALHOU'))
);

CREATE UNIQUE INDEX idx_dominios_site_hostname_unico ON dominios_site (LOWER(hostname));
CREATE INDEX idx_dominios_site_ativos ON dominios_site (LOWER(hostname)) WHERE status='ATIVO';

