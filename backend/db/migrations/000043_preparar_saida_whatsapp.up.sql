ALTER TABLE mensagens_whatsapp ALTER COLUMN identificador_externo DROP NOT NULL;
ALTER TABLE mensagens_whatsapp ADD COLUMN evento_outbox_id uuid REFERENCES eventos_outbox(id) ON DELETE SET NULL;
ALTER TABLE mensagens_whatsapp ADD COLUMN enviada_em timestamptz;
ALTER TABLE mensagens_whatsapp ADD COLUMN entregue_em timestamptz;
ALTER TABLE mensagens_whatsapp ADD COLUMN lida_em timestamptz;
ALTER TABLE mensagens_whatsapp ADD COLUMN falhou_em timestamptz;
ALTER TABLE mensagens_whatsapp ADD COLUMN erro_codigo varchar(40);
ALTER TABLE mensagens_whatsapp ADD COLUMN erro_detalhe varchar(500);
ALTER TABLE conversas_whatsapp ADD COLUMN consentimento_marketing_origem varchar(80);
ALTER TABLE conversas_whatsapp ADD COLUMN consentimento_marketing_evidencia varchar(500);

CREATE UNIQUE INDEX idx_mensagens_whatsapp_evento_outbox
    ON mensagens_whatsapp (evento_outbox_id)
    WHERE evento_outbox_id IS NOT NULL;

CREATE TABLE status_mensagens_whatsapp (
    identificador_externo varchar(255) PRIMARY KEY,
    status varchar(20) NOT NULL,
    ordem smallint NOT NULL,
    ocorrido_em timestamptz NOT NULL,
    erro_codigo varchar(40),
    erro_detalhe varchar(500),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT status_mensagens_whatsapp_status_check CHECK (status IN ('ENVIADA', 'ENTREGUE', 'LIDA', 'FALHOU'))
);
