CREATE TABLE sessoes_usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL,
    conta_id UUID NOT NULL,
    expira_em TIMESTAMPTZ NOT NULL,
    revogada_em TIMESTAMPTZ,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sessoes_usuario_mesma_conta_fk
        FOREIGN KEY (usuario_id, conta_id) REFERENCES usuarios(id, conta_id) ON DELETE CASCADE
);

CREATE INDEX idx_sessoes_usuario_ativas
    ON sessoes_usuario (usuario_id, expira_em DESC)
    WHERE revogada_em IS NULL;
