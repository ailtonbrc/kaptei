CREATE TABLE IF NOT EXISTS agendamentos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    cliente_id UUID REFERENCES clientes(id) ON DELETE SET NULL,
    imovel_id UUID REFERENCES imoveis(id) ON DELETE SET NULL,
    titulo VARCHAR(255) NOT NULL,
    descricao TEXT,
    data_hora_inicio TIMESTAMPTZ NOT NULL,
    data_hora_fim TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) DEFAULT 'AGENDADO',
    tipo VARCHAR(50) DEFAULT 'VISITA',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT agendamentos_periodo_valido CHECK (data_hora_fim > data_hora_inicio)
);

CREATE INDEX idx_agendamentos_conta_id ON agendamentos(conta_id);
CREATE INDEX idx_agendamentos_data_inicio ON agendamentos(data_hora_inicio);
