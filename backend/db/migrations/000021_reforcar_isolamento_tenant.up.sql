-- Chaves compostas permitem que relacionamentos validem o tenant do registro relacionado.
ALTER TABLE usuarios ADD CONSTRAINT usuarios_id_conta_unique UNIQUE (id, conta_id);
ALTER TABLE clientes ADD CONSTRAINT clientes_id_conta_unique UNIQUE (id, conta_id);
ALTER TABLE imoveis ADD CONSTRAINT imoveis_id_conta_unique UNIQUE (id, conta_id);

ALTER TABLE imoveis
    ADD CONSTRAINT imoveis_usuario_mesma_conta_fk
    FOREIGN KEY (usuario_id, conta_id) REFERENCES usuarios (id, conta_id) NOT VALID;

ALTER TABLE clientes
    ADD CONSTRAINT clientes_corretor_mesma_conta_fk
    FOREIGN KEY (corretor_id, conta_id) REFERENCES usuarios (id, conta_id) NOT VALID;

ALTER TABLE leads
    ADD CONSTRAINT leads_usuario_mesma_conta_fk
    FOREIGN KEY (usuario_id, conta_id) REFERENCES usuarios (id, conta_id) NOT VALID,
    ADD CONSTRAINT leads_imovel_mesma_conta_fk
    FOREIGN KEY (imovel_id, conta_id) REFERENCES imoveis (id, conta_id) NOT VALID;

ALTER TABLE interacoes
    ADD CONSTRAINT interacoes_cliente_mesma_conta_fk
    FOREIGN KEY (cliente_id, conta_id) REFERENCES clientes (id, conta_id) NOT VALID,
    ADD CONSTRAINT interacoes_corretor_mesma_conta_fk
    FOREIGN KEY (corretor_id, conta_id) REFERENCES usuarios (id, conta_id) NOT VALID;

ALTER TABLE agendamentos
    ADD CONSTRAINT agendamentos_usuario_mesma_conta_fk
    FOREIGN KEY (usuario_id, conta_id) REFERENCES usuarios (id, conta_id) NOT VALID,
    ADD CONSTRAINT agendamentos_cliente_mesma_conta_fk
    FOREIGN KEY (cliente_id, conta_id) REFERENCES clientes (id, conta_id) NOT VALID,
    ADD CONSTRAINT agendamentos_imovel_mesma_conta_fk
    FOREIGN KEY (imovel_id, conta_id) REFERENCES imoveis (id, conta_id) NOT VALID;

CREATE INDEX idx_clientes_conta_corretor ON clientes (conta_id, corretor_id);
CREATE INDEX idx_leads_conta_status_criado ON leads (conta_id, status, criado_em DESC);
CREATE INDEX idx_agendamentos_conta_periodo ON agendamentos (conta_id, data_hora_inicio, data_hora_fim);
