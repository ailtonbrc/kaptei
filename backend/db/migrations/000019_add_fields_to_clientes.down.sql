DROP TABLE IF EXISTS interacoes;

ALTER TABLE clientes 
    DROP COLUMN cpf,
    DROP COLUMN data_nascimento,
    DROP COLUMN estado_civil,
    DROP COLUMN financeiro,
    DROP COLUMN origem_utm,
    DROP COLUMN corretor_id,
    DROP COLUMN temperatura,
    DROP COLUMN proxima_acao;
