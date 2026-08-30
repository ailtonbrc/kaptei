ALTER TABLE leads ADD COLUMN chave_idempotencia UUID;

CREATE UNIQUE INDEX idx_leads_captura_idempotente
    ON leads (conta_id, chave_idempotencia)
    WHERE chave_idempotencia IS NOT NULL;
