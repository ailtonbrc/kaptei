ALTER TABLE leads
    ADD COLUMN consentimento_lgpd_em TIMESTAMPTZ,
    ADD COLUMN consentimento_lgpd_versao VARCHAR(40);

UPDATE leads
SET consentimento_lgpd_em = criado_em,
    consentimento_lgpd_versao = 'legado-sem-versao'
WHERE consentimento_lgpd = TRUE
  AND consentimento_lgpd_em IS NULL;

ALTER TABLE leads
    ADD CONSTRAINT leads_consentimento_evidencia_ck
    CHECK (
        consentimento_lgpd = FALSE OR
        (consentimento_lgpd_em IS NOT NULL AND consentimento_lgpd_versao IS NOT NULL)
    );
