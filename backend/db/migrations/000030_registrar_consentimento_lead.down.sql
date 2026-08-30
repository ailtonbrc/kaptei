ALTER TABLE leads
    DROP CONSTRAINT IF EXISTS leads_consentimento_evidencia_ck,
    DROP COLUMN IF EXISTS consentimento_lgpd_versao,
    DROP COLUMN IF EXISTS consentimento_lgpd_em;
