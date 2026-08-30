CREATE TABLE auditoria_eventos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID REFERENCES contas_saas(id) ON DELETE SET NULL,
    usuario_id UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    request_id VARCHAR(80),
    metodo VARCHAR(10) NOT NULL,
    rota TEXT NOT NULL,
    status_http INTEGER NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_auditoria_conta_criado ON auditoria_eventos (conta_id, criado_em DESC);
CREATE INDEX idx_auditoria_usuario_criado ON auditoria_eventos (usuario_id, criado_em DESC);
