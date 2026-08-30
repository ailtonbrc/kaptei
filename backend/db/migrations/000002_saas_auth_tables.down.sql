ALTER TABLE imoveis DROP CONSTRAINT IF EXISTS imoveis_conta_id_fkey;
ALTER TABLE clientes DROP CONSTRAINT IF EXISTS clientes_conta_id_fkey;
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_conta_id_fkey;

ALTER TABLE imoveis RENAME COLUMN conta_id TO id_imobiliaria;
ALTER TABLE clientes RENAME COLUMN conta_id TO id_imobiliaria;

ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS id_imobiliaria uuid REFERENCES usuarios(id);
ALTER TABLE usuarios DROP COLUMN IF EXISTS conta_id;
ALTER TABLE usuarios DROP COLUMN IF EXISTS google_id;
ALTER TABLE usuarios DROP COLUMN IF EXISTS senha_hash;

DROP TABLE IF EXISTS contas_saas;
