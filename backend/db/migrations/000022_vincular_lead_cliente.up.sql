ALTER TABLE leads ADD COLUMN cliente_id UUID;

ALTER TABLE leads
    ADD CONSTRAINT leads_cliente_mesma_conta_fk
    FOREIGN KEY (cliente_id, conta_id) REFERENCES clientes (id, conta_id) NOT VALID;

CREATE UNIQUE INDEX idx_leads_cliente_unico
    ON leads (cliente_id)
    WHERE cliente_id IS NOT NULL;
