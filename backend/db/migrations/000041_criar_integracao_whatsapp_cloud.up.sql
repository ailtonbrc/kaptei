CREATE TABLE integracoes_whatsapp_cloud (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conta_id uuid NOT NULL REFERENCES contas_saas(id) ON DELETE CASCADE,
    waba_id varchar(64) NOT NULL,
    numero_telefone_id varchar(64) NOT NULL,
    numero_exibicao varchar(32),
    token_acesso_protegido text NOT NULL,
    ativa boolean NOT NULL DEFAULT true,
    criado_em timestamptz NOT NULL DEFAULT now(),
    atualizado_em timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integracoes_whatsapp_conta_unique UNIQUE (conta_id),
    CONSTRAINT integracoes_whatsapp_numero_unique UNIQUE (numero_telefone_id),
    CONSTRAINT integracoes_whatsapp_waba_formato CHECK (waba_id ~ '^[0-9]{1,64}$'),
    CONSTRAINT integracoes_whatsapp_numero_formato CHECK (numero_telefone_id ~ '^[0-9]{1,64}$')
);
