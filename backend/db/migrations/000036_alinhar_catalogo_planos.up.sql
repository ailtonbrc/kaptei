-- O catálogo comercial deve anunciar apenas capacidades efetivamente entregues.
-- Enquanto não houver regras de entitlement por plano, os recursos ativos são
-- iguais dentro de cada tipo de conta. A diferenciação futura deve ser criada
-- junto com a respectiva regra de autorização no backend.

UPDATE planos
SET features = '[
    "CRM de clientes e leads",
    "Cadastro e publicação de imóveis",
    "Agenda de visitas e tarefas",
    "Mini-site público responsivo",
    "Formulários de captação com consentimento LGPD",
    "Webhook seguro para integração de leads",
    "Dashboard de captação e conversão",
    "Gestão segura da assinatura"
]'::jsonb,
missing = '[
    "Publicação automática em portais externos",
    "Automação de WhatsApp",
    "Propostas e contratos PDF",
    "Aplicativo móvel offline",
    "IA para anúncios e atendimento",
    "Nota fiscal / RPA"
]'::jsonb
WHERE tipo = 'CORRETOR';

UPDATE planos
SET features = '[
    "CRM e carteira compartilhada",
    "Cadastro e publicação de imóveis",
    "Agenda de visitas e tarefas",
    "Mini-site público responsivo",
    "Formulários de captação com consentimento LGPD",
    "Equipe com convites e controle de acesso",
    "Distribuição manual ou automática de leads",
    "Dashboard consolidado da imobiliária",
    "Webhook seguro para integração de leads",
    "Gestão segura da assinatura"
]'::jsonb,
missing = '[
    "Publicação automática em portais externos",
    "Automação de WhatsApp",
    "Propostas e contratos PDF",
    "Aplicativo móvel offline",
    "Metas e ranking por corretor",
    "Módulo completo de locação",
    "Nota fiscal / RPA"
]'::jsonb
WHERE tipo = 'IMOBILIARIA';
