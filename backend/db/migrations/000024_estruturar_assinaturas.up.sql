ALTER TABLE planos ADD COLUMN gateway_price_id VARCHAR(120);

ALTER TABLE contas_saas
    ADD COLUMN billing_customer_id VARCHAR(120),
    ADD COLUMN billing_subscription_id VARCHAR(120),
    ADD COLUMN billing_status VARCHAR(40),
    ADD COLUMN billing_periodo_fim TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_contas_billing_customer_unico ON contas_saas (billing_customer_id) WHERE billing_customer_id IS NOT NULL;
CREATE UNIQUE INDEX idx_contas_billing_subscription_unico ON contas_saas (billing_subscription_id) WHERE billing_subscription_id IS NOT NULL;

CREATE TABLE billing_eventos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provedor VARCHAR(30) NOT NULL,
    evento_id VARCHAR(180) NOT NULL,
    tipo VARCHAR(120) NOT NULL,
    conta_id UUID REFERENCES contas_saas(id) ON DELETE SET NULL,
    recebido_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processado_em TIMESTAMPTZ,
    erro TEXT,
    CONSTRAINT billing_eventos_provedor_evento_unique UNIQUE (provedor, evento_id)
);
CREATE INDEX idx_billing_eventos_pendentes ON billing_eventos (recebido_em) WHERE processado_em IS NULL;
