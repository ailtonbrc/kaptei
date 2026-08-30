CREATE TABLE IF NOT EXISTS imoveis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id UUID NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    
    titulo VARCHAR(255) NOT NULL,
    tipo VARCHAR(50) NOT NULL,
    finalidade VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Ativo',
    
    valor_venda NUMERIC(15,2),
    valor_locacao NUMERIC(15,2),
    valor_condominio NUMERIC(15,2),
    valor_iptu NUMERIC(15,2),
    
    area_total NUMERIC(10,2),
    area_util NUMERIC(10,2),
    quartos INTEGER DEFAULT 0,
    suites INTEGER DEFAULT 0,
    banheiros INTEGER DEFAULT 0,
    vagas INTEGER DEFAULT 0,
    
    cep VARCHAR(20),
    logradouro VARCHAR(255),
    numero VARCHAR(50),
    complemento VARCHAR(150),
    bairro VARCHAR(150),
    cidade VARCHAR(150),
    estado VARCHAR(2),
    
    descricao TEXT,
    
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices para otimizar busca no dashboard
CREATE INDEX idx_imoveis_conta_id ON imoveis(conta_id);
CREATE INDEX idx_imoveis_tipo ON imoveis(tipo);
CREATE INDEX idx_imoveis_finalidade ON imoveis(finalidade);
CREATE INDEX idx_imoveis_status ON imoveis(status);

CREATE TABLE IF NOT EXISTS imovel_fotos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    imovel_id UUID NOT NULL REFERENCES imoveis(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    ordem INTEGER DEFAULT 0,
    is_capa BOOLEAN DEFAULT FALSE,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_imovel_fotos_imovel_id ON imovel_fotos(imovel_id);
