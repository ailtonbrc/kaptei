DROP TABLE IF EXISTS mensagens_whatsapp;
DROP TABLE IF EXISTS conversas_whatsapp;
ALTER TABLE eventos_integracao DROP COLUMN IF EXISTS payload_protegido;
