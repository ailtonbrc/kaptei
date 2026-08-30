CREATE INDEX IF NOT EXISTS idx_clientes_conta_atualizado
    ON clientes (conta_id, atualizado_em DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_clientes_conta_corretor_atualizado
    ON clientes (conta_id, corretor_id, atualizado_em DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_clientes_conta_status
    ON clientes (conta_id, status_funil);

CREATE INDEX IF NOT EXISTS idx_leads_conta_criado
    ON leads (conta_id, criado_em DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_leads_conta_usuario_criado
    ON leads (conta_id, usuario_id, criado_em DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_leads_conta_status
    ON leads (conta_id, status);

CREATE INDEX IF NOT EXISTS idx_imoveis_conta_criado
    ON imoveis (conta_id, criado_em DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_imoveis_conta_catalogo
    ON imoveis (conta_id, tipo, finalidade, status);
