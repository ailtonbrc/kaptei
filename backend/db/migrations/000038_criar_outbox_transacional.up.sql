CREATE TABLE IF NOT EXISTS eventos_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NULL REFERENCES contas_saas(id) ON DELETE SET NULL,
    tipo VARCHAR(80) NOT NULL,
    payload_protegido TEXT NOT NULL,
    chave_idempotencia VARCHAR(200) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDENTE'
        CHECK (status IN ('PENDENTE', 'PROCESSANDO', 'CONCLUIDO', 'FALHOU')),
    tentativas INTEGER NOT NULL DEFAULT 0 CHECK (tentativas >= 0),
    maximo_tentativas INTEGER NOT NULL CHECK (maximo_tentativas BETWEEN 1 AND 100),
    disponivel_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    bloqueado_ate TIMESTAMPTZ NULL,
    bloqueado_por VARCHAR(120) NULL,
    ultimo_erro VARCHAR(1000) NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processado_em TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_eventos_outbox_processamento
    ON eventos_outbox (disponivel_em, criado_em)
    WHERE status IN ('PENDENTE', 'PROCESSANDO');

CREATE INDEX IF NOT EXISTS idx_eventos_outbox_conta_criado
    ON eventos_outbox (conta_id, criado_em DESC)
    WHERE conta_id IS NOT NULL;
