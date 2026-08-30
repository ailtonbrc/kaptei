# Kaptei

Plataforma SaaS multiempresa para captação imobiliária, CRM, agenda, catálogo público e assinaturas.

## Decisão de interface

O produto preserva a identidade visual atual e utiliza React, TypeScript, Tailwind e componentes Shadcn/Radix. Tremor atende aos indicadores executivos e ECharts modular às visualizações analíticas complexas. O projeto não usa Ant Design porque a diretriz local do repositório o proíbe e a migração integral acrescentaria peso e risco sem benefício proporcional.

## Estrutura

- `backend/cmd`: pontos de entrada da API, migrador e servidor estático.
- `backend/internal/core`: domínio, portas e regras de negócio.
- `backend/internal/adapters`: HTTP, PostgreSQL, e-mail e gateways externos.
- `backend/internal/plataforma`: configuração, banco e composição das dependências.
- `frontend/src/pages/Publico`: site de captação e catálogo imobiliário.
- `frontend/src/pages`: painel SaaS, CRM, agenda, imóveis, cobrança e configurações.
- `docs`: arquitetura, segurança e procedimentos operacionais.

O estado verificável de cada capacidade, suas evidências e os itens ainda não entregues estão em `docs/MATRIZ_RASTREABILIDADE.md`.

## Começar com segurança

1. Copie `backend/.env.example` para `backend/.env` e substitua todos os placeholders.
2. Copie `frontend/.env.example` para `frontend/.env`.
3. Aplique as migrations com o comando documentado em `docs/OPERACAO.md`.
4. Inicie API e frontend apenas depois de validar banco, CORS, cookies e URLs públicas.

O primeiro `SUPER_ADMIN` deve ser criado pelo comando seguro descrito em `docs/OPERACAO.md`; não existe credencial administrativa padrão.

Não coloque `.env`, chaves, credenciais ou binários no Git.
