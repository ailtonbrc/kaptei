INSERT INTO configuracoes_sistema (chave, valor, descricao)
VALUES ('GOOGLE_CLIENT_ID', '""', 'ID do Cliente para autenticacao via Google')
ON CONFLICT (chave) DO NOTHING;
