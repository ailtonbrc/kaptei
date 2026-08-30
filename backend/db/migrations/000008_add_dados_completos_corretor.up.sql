ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS rg text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS nacionalidade text DEFAULT 'Brasileira';
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS estado_civil text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS cep text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS logradouro text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS numero text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS complemento text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS bairro text;
