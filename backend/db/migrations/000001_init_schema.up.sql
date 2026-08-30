-- Ativação da extensão para geração de UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- USUÁRIOS
CREATE TABLE IF NOT EXISTS usuarios (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 nome_completo text NOT NULL,
 email text NOT NULL UNIQUE,
 telefone text,
 cpf text UNIQUE,
 papel text NOT NULL DEFAULT 'visitante',
 status text NOT NULL DEFAULT 'ativo',
 creci text,
 creci_estado text DEFAULT 'MS',
 creci_verificado boolean DEFAULT false,
 creci_verificado_em timestamptz,
 biografia text,
 url_avatar text,
 cidade text,
 estado text DEFAULT 'MS',
 plano text DEFAULT 'trial',
 plano_expira_em timestamptz,
 id_imobiliaria uuid REFERENCES usuarios(id),
 pontuacao_saude integer DEFAULT 0,
 tempo_medio_resposta_minutos integer,
 instagram text,
 facebook text,
 url_site text,
 numero_whatsapp text,
 data_nascimento date,
 criado_em timestamptz DEFAULT now(),
 atualizado_em timestamptz DEFAULT now()
);

-- ASSINATURAS
CREATE TABLE IF NOT EXISTS assinaturas (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 id_usuario uuid REFERENCES usuarios(id) ON DELETE CASCADE,
 plano text NOT NULL,
 status text DEFAULT 'trial',
 fim_trial_em timestamptz DEFAULT now() + interval '14 days',
 inicio_periodo_atual timestamptz,
 fim_periodo_atual timestamptz,
 valor_negociado numeric,
 id_assinatura_gateway text,
 id_cliente_gateway text,
 funcionalidades_extras jsonb DEFAULT '{}',
 cancelar_no_fim_periodo boolean DEFAULT false,
 criado_em timestamptz DEFAULT now()
);


