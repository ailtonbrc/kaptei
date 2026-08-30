INSERT INTO planos (codigo, tipo, nome, subtitle, preco, cor, recomendado, features, missing) VALUES
('CORRETOR_TRIAL', 'CORRETOR', 'TRIAL 14 DIAS', 'Acesso gratuito (Premium)', 0.00, '#10b981', false, 
 '["CRM + Pipeline Kanban", "Cadastro de imóveis (formulário)", "Morar.com", "Termo de visita digital SMS", "Proposta e Contrato PDF", "Mini-site + domínio próprio", "Cartão digital profissional", "Simulador de financiamento", "AMC avançada", "Kai — Agente de IA", "Arte para redes sociais", "App mobile PWA offline", "Publicação portais externos", "Robô WhatsApp 24h", "Relatórios e métricas", "IA para anúncios", "Nota fiscal / RPA"]'::jsonb, 
 '[]'::jsonb),

('IMOBILIARIA_TRIAL', 'IMOBILIARIA', 'TRIAL 14 DIAS', 'Acesso gratuito (Completa)', 0.00, '#10b981', false, 
 '["Tudo do Corretor Premium", "Gestão de equipe (ilimitada)", "Distribuição automática de leads", "Escala de plantão mensal", "Relatório consolidado da equipe", "Carteira compartilhada", "Metas e ranking por corretor", "Relatório de captação", "Módulo de Locação completo", "API pública B2B"]'::jsonb, 
 '[]'::jsonb)
ON CONFLICT (codigo) DO UPDATE 
SET features = EXCLUDED.features, 
    missing = EXCLUDED.missing,
    subtitle = EXCLUDED.subtitle,
    recomendado = EXCLUDED.recomendado;
