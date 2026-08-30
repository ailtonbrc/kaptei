ALTER TABLE planos
    ADD COLUMN limite_corretores INTEGER;

UPDATE planos SET limite_corretores = 5 WHERE codigo = 'IMOBILIARIA_BASICA';

ALTER TABLE planos
    ADD CONSTRAINT planos_limite_corretores_positivo_ck
    CHECK (limite_corretores IS NULL OR limite_corretores > 0);

CREATE TABLE convites_equipe (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    email VARCHAR(254) NOT NULL,
    papel VARCHAR(30) NOT NULL DEFAULT 'CORRETOR_EQUIPE',
    token_hash CHAR(64) NOT NULL UNIQUE,
    convidado_por UUID NOT NULL,
    expira_em TIMESTAMPTZ NOT NULL,
    usado_em TIMESTAMPTZ,
    revogado_em TIMESTAMPTZ,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT convites_equipe_papel_ck CHECK (papel = 'CORRETOR_EQUIPE'),
    CONSTRAINT convites_equipe_convidador_mesma_conta_fk
        FOREIGN KEY (convidado_por, conta_id) REFERENCES usuarios(id, conta_id)
);

CREATE UNIQUE INDEX idx_convites_equipe_email_pendente
    ON convites_equipe (conta_id, LOWER(email))
    WHERE usado_em IS NULL AND revogado_em IS NULL;

CREATE INDEX idx_convites_equipe_pendentes
    ON convites_equipe (conta_id, criado_em DESC)
    WHERE usado_em IS NULL AND revogado_em IS NULL;
