INSERT INTO configuracoes_sistema (chave, valor, descricao, atualizado_em)
VALUES (
    'OBSERVABILIDADE_CONFIG',
    '{"ativa":false,"token":""}'::jsonb,
    'Acesso protegido ao endpoint Prometheus',
    NOW()
)
ON CONFLICT (chave) DO NOTHING;
