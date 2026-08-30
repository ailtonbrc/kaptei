INSERT INTO planos (codigo, tipo, nome, subtitle, preco, cor, recomendado, features, missing) VALUES
('CORRETOR_BASICO', 'CORRETOR', 'BÁSICO', null, 97.00, '#3b82f6', false, 
 '["CRM + Pipeline Kanban", "Cadastro de imóveis (formulário)", "Morar.com", "Termo de visita digital SMS", "Proposta e Contrato PDF", "Mini-site do corretor", "Cartão digital profissional", "Simulador de financiamento", "AMC — Avaliação de Mercado", "Kai — Agente de IA", "Arte para redes sociais", "App mobile PWA offline"]'::jsonb, 
 '["Publicação portais externos", "Robô WhatsApp 24h", "Relatórios e métricas", "IA para anúncios", "Nota fiscal / RPA"]'::jsonb),

('CORRETOR_PRO', 'CORRETOR', 'PROFISSIONAL', null, 197.00, '#8b5cf6', true, 
 '["CRM + Pipeline Kanban", "Cadastro de imóveis (formulário)", "Morar.com", "Termo de visita digital SMS", "Proposta e Contrato PDF", "Mini-site do corretor", "Cartão digital profissional", "Simulador de financiamento", "AMC — Avaliação de Mercado", "Kai — Agente de IA", "Arte para redes sociais", "App mobile PWA offline", "Publicação portais externos", "Robô WhatsApp 24h", "Relatórios e métricas"]'::jsonb, 
 '["IA para anúncios", "Nota fiscal / RPA"]'::jsonb),

('CORRETOR_PREMIUM', 'CORRETOR', 'PREMIUM', null, 347.00, '#eab308', false, 
 '["CRM + Pipeline Kanban", "Cadastro de imóveis (formulário)", "Morar.com", "Termo de visita digital SMS", "Proposta e Contrato PDF", "Mini-site + domínio próprio", "Cartão digital profissional", "Simulador de financiamento", "AMC avançada", "Kai — Agente de IA", "Arte para redes sociais", "App mobile PWA offline", "Publicação portais externos", "Robô WhatsApp 24h", "Relatórios e métricas", "IA para anúncios", "Nota fiscal / RPA"]'::jsonb, 
 '[]'::jsonb),

('IMOBILIARIA_BASICA', 'IMOBILIARIA', 'BÁSICA', 'Até 5 corretores', 497.00, '#3b82f6', false, 
 '["Tudo do Corretor Premium", "Gestão de equipe (até 5 corret)", "Distribuição automática de leads", "Escala de plantão mensal", "Relatório consolidado da equipe", "Carteira compartilhada", "Metas e ranking por corretor", "Relatório de captação"]'::jsonb, 
 '["Módulo de Locação completo", "API pública B2B"]'::jsonb),

('IMOBILIARIA_PRO', 'IMOBILIARIA', 'PROFISSIONAL', 'Corretores ilimitados', 797.00, '#8b5cf6', true, 
 '["Tudo do Corretor Premium", "Gestão de equipe (ilimitada)", "Distribuição automática de leads", "Escala de plantão mensal", "Relatório consolidado da equipe", "Carteira compartilhada", "Metas e ranking por corretor", "Relatório de captação", "API pública B2B"]'::jsonb, 
 '["Módulo de Locação completo"]'::jsonb),

('IMOBILIARIA_COMPLETA', 'IMOBILIARIA', 'COMPLETA', 'Corretores ilimitados', 1197.00, '#eab308', false, 
 '["Tudo do Corretor Premium", "Gestão de equipe (ilimitada)", "Distribuição automática de leads", "Escala de plantão mensal", "Relatório consolidado da equipe", "Carteira compartilhada", "Metas e ranking por corretor", "Relatório de captação", "Módulo de Locação completo", "API pública B2B"]'::jsonb, 
 '[]'::jsonb)
ON CONFLICT (codigo) DO NOTHING;
