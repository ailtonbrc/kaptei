-- O catálogo comunica capacidades existentes; limites e diferenciação por plano
-- só podem ser anunciados quando houver enforcement correspondente no backend.
UPDATE planos
SET features = '[
    "CRM de clientes e leads",
    "Cadastro e publicação de imóveis",
    "Agenda de visitas e tarefas",
    "Site público responsivo com domínio próprio e SEO",
    "Formulários de captação com consentimento LGPD",
    "Central de privacidade e retenção",
    "Webhook seguro para integração de leads",
    "WhatsApp Cloud API configurável",
    "Publicação configurável no Grupo OLX via VRSync",
    "Dashboard de captação e conversão",
    "Gestão segura da assinatura"
]'::jsonb,
missing = '[
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
    "Site público responsivo com domínio próprio e SEO",
    "Formulários de captação com consentimento LGPD",
    "Central de privacidade e retenção",
    "Equipe com convites e controle de acesso",
    "Distribuição manual ou automática de leads",
    "WhatsApp Cloud API com caixa de atendimento",
    "Publicação configurável no Grupo OLX via VRSync",
    "Dashboard consolidado da imobiliária",
    "Webhook seguro para integração de leads",
    "Gestão segura da assinatura"
]'::jsonb,
missing = '[
    "Propostas e contratos PDF",
    "Aplicativo móvel offline",
    "IA para anúncios e atendimento",
    "Metas e ranking por corretor",
    "Módulo completo de locação",
    "Nota fiscal / RPA"
]'::jsonb
WHERE tipo = 'IMOBILIARIA';
