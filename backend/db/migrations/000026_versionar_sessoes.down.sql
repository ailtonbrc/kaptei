ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_versao_sessao_positiva;
ALTER TABLE usuarios DROP COLUMN IF EXISTS versao_sessao;
