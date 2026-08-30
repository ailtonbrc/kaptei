CREATE TABLE IF NOT EXISTS contas_saas (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tipo_conta text NOT NULL DEFAULT 'CORRETOR_SOLO', -- 'CORRETOR_SOLO' ou 'IMOBILIARIA'
    nome_conta text, -- Nome da Imobiliária (se for o caso)
    status_plano text NOT NULL DEFAULT 'TRIAL',
    feature_flags jsonb DEFAULT '{}',
    criado_em timestamptz DEFAULT now(),
    atualizado_em timestamptz DEFAULT now()
);

-- Modificações na tabela usuarios
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS senha_hash text;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS google_id text UNIQUE;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS conta_id uuid REFERENCES contas_saas(id);

-- Removendo antigas amarras de tenant baseadas no id do usuario
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_id_imobiliaria_fkey;

ALTER TABLE usuarios DROP COLUMN IF EXISTS id_imobiliaria;
