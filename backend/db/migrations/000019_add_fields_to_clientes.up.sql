ALTER TABLE clientes 
    ADD COLUMN cpf VARCHAR(20),
    ADD COLUMN data_nascimento VARCHAR(10),
    ADD COLUMN estado_civil VARCHAR(50),
    ADD COLUMN financeiro JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN origem_utm JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN corretor_id UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    ADD COLUMN temperatura VARCHAR(20),
    ADD COLUMN proxima_acao TIMESTAMP WITH TIME ZONE;

CREATE TABLE IF NOT EXISTS interacoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    cliente_id UUID NOT NULL REFERENCES clientes(id) ON DELETE CASCADE,
    corretor_id UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    tipo VARCHAR(50) NOT NULL,
    descricao TEXT,
    data_hora TIMESTAMP WITH TIME ZONE NOT NULL,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
