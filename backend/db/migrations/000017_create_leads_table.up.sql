ALTER TABLE contas_saas ADD COLUMN IF NOT EXISTS lead_estrategia text DEFAULT 'CAIXA_ENTRADA';
ALTER TABLE contas_saas ADD COLUMN IF NOT EXISTS lead_token_integracao uuid DEFAULT gen_random_uuid();

-- Tentativa de garantir unicidade (embora UUID randômico já seja)
ALTER TABLE contas_saas ADD CONSTRAINT contas_saas_lead_token_unique UNIQUE (lead_token_integracao);

CREATE TABLE IF NOT EXISTS leads (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    usuario_id uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    imovel_id uuid REFERENCES imoveis(id) ON DELETE SET NULL,
    
    nome varchar(255) NOT NULL,
    email varchar(255),
    telefone varchar(50),
    origem varchar(100),
    mensagem text,
    
    status varchar(50) DEFAULT 'NOVO',
    motivo_descarte text,
    
    criado_em timestamptz DEFAULT now(),
    atualizado_em timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_leads_conta_id ON leads(conta_id);
CREATE INDEX IF NOT EXISTS idx_leads_usuario_id ON leads(usuario_id);
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
