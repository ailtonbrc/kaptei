CREATE TABLE IF NOT EXISTS clientes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    telefone VARCHAR(50),
    status_funil VARCHAR(50) DEFAULT 'NOVO',
    origem VARCHAR(100),
    interesse_tipo VARCHAR(50),
    notas TEXT,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clientes_conta_id ON clientes(conta_id);
CREATE INDEX idx_clientes_status_funil ON clientes(status_funil);
