CREATE TABLE IF NOT EXISTS planos (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo text UNIQUE NOT NULL,
    tipo text NOT NULL, -- 'CORRETOR' ou 'IMOBILIARIA'
    nome text NOT NULL,
    subtitle text,
    preco numeric NOT NULL DEFAULT 0,
    cor text,
    recomendado boolean DEFAULT false,
    features jsonb DEFAULT '[]',
    missing jsonb DEFAULT '[]',
    ativo boolean DEFAULT true,
    criado_em timestamptz DEFAULT now(),
    atualizado_em timestamptz DEFAULT now()
);
