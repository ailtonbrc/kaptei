-- Tabela genérica para configurações do sistema (Chave/Valor)
CREATE TABLE IF NOT EXISTS configuracoes_sistema (
    chave text PRIMARY KEY,
    valor jsonb NOT NULL,
    descricao text,
    atualizado_em timestamptz DEFAULT now()
);

-- Tabela para gerenciar os tokens de recuperação de senha
CREATE TABLE IF NOT EXISTS recuperacao_senha_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    token text NOT NULL UNIQUE,
    expira_em timestamptz NOT NULL,
    usado boolean DEFAULT false,
    criado_em timestamptz DEFAULT now()
);

-- Inserindo a configuração padrão de SMTP vazia (para ser preenchida pelo usuário depois no banco)
INSERT INTO configuracoes_sistema (chave, valor, descricao) 
VALUES (
    'SMTP_CONFIG', 
    '{"host": "", "port": 587, "user": "", "password": "", "from_email": "", "from_name": "Sistema Kaptei"}',
    'Configurações do servidor SMTP para envio de e-mails do sistema'
) ON CONFLICT (chave) DO NOTHING;
