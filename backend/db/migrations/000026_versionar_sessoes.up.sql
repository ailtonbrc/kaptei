ALTER TABLE usuarios
    ADD COLUMN versao_sessao INTEGER NOT NULL DEFAULT 1;

ALTER TABLE usuarios
    ADD CONSTRAINT usuarios_versao_sessao_positiva CHECK (versao_sessao > 0);
