DROP TABLE IF EXISTS convites_equipe;

ALTER TABLE planos
    DROP CONSTRAINT IF EXISTS planos_limite_corretores_positivo_ck,
    DROP COLUMN IF EXISTS limite_corretores;
