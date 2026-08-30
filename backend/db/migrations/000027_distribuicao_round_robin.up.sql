CREATE TABLE lead_distribuicao_estado (
    conta_id UUID PRIMARY KEY REFERENCES contas_saas(id) ON DELETE CASCADE,
    ultimo_usuario_id UUID,
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lead_distribuicao_usuario_mesma_conta_fk
        FOREIGN KEY (ultimo_usuario_id, conta_id) REFERENCES usuarios(id, conta_id)
);
