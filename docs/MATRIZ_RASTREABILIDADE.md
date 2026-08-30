# Matriz de rastreabilidade do Kaptei

## Como interpretar

Esta matriz é a fonte de verdade do escopo técnico em 07 de agosto de 2026. “Implementado” significa que há código e ligação entre as camadas. “Aguardando homologação” significa que a implementação existe, mas ainda depende de migrations, serviços externos ou testes manuais. “Não implementado” não deve ser anunciado, vendido ou tratado como concluído.

## Capacidades do produto

| Capacidade | Estado | Evidência principal | Condição para aceite |
|---|---|---|---|
| Cadastro, login, Google, recuperação e troca de senha | Implementado; aguardando homologação | `backend/internal/core/services/auth_service.go`, `frontend/src/pages/Auth` | Executar jornadas com SMTP e Google reais |
| Sessão HttpOnly, logout e revogação por dispositivo | Implementado; aguardando homologação | migrations 26 e 33, `jwt_middleware.go`, `sessao_postgres.go` | Validar cookies HTTPS, expiração e revogação |
| Isolamento multiempresa e autorização por papel | Implementado; aguardando testes de integração | migrations 21, 22, 29 e 32; repositórios por `conta_id`; `RoleRoute.tsx` | Tentar CRUD cruzado entre dois tenants |
| CRM de clientes e interações | Implementado | `cliente_service.go`, `interacao_service.go`, páginas em `frontend/src/pages/CRM` | Homologar CRUD, paginação e permissões |
| Caixa e distribuição de leads | Implementado | `lead_service.go`, migration 27, `LeadsInbox.tsx` | Homologar distribuição manual/round-robin e conversão |
| Entrada genérica de leads por webhook | Implementado | `POST /api/webhooks/leads`, token rotacionável em hash | Homologar idempotência do produtor e rotação do token |
| Cadastro, fotos e publicação de imóveis | Implementado | `imovel_service.go`, migrations 10, 28 e 32 | Homologar upload, ordenação e isolamento tenant |
| Agenda vinculada a cliente/imóvel | Implementado | migration 20, `agendamento_service.go`, `BotaoAgendar.tsx` | Homologar calendário e ações rápidas |
| Site público, catálogo, formulários e UTM | Implementado; aguardando conteúdo real | migrations 23, 30 e 34; `site_publico_service.go`; páginas em `frontend/src/pages/Publico` | Publicar tenant real, revisar textos e consentimento |
| Sitemap, robots, canonical, domínios e pré-renderização | Implementado; aguardando DNS/TLS real | migration 46, `dominio_prerender.go`, `seo_dominio.go`, rotas `/sitemap.xml` e `/robots.txt`; `useMetadadosPagina.ts` | Homologar domínio real, cache, previews sociais e robôs sem JavaScript |
| Dashboard operacional | Implementado | `dashboard_service.go`, Tremor em `IndicadorDashboard` e ECharts modular | Comparar métricas com consultas de controle |
| Equipe, convite e limite do plano | Implementado; aguardando homologação de e-mail | migration 31, `equipe_service.go`, `frontend/src/pages/Equipe` | Testar convite, expiração, cancelamento e inativação |
| Stripe Checkout, portal e webhook | Implementado; depende da conta Stripe | `stripe_gateway.go`, `billing_service.go`, migration 24 | Configurar IDs/segredos e homologar todos os eventos |
| Auditoria, request ID, logs, health e ready | Implementado | migration 25, middlewares em `backend/internal/adapters/middlewares`, rotas `/health` e `/ready` | Conectar coleta de logs, métricas e alertas |
| Paginação de clientes, leads e imóveis | Implementado | migration 37, tipos de paginação e repositórios PostgreSQL | Testar volume e planos de execução na base real |
| Proteção do token de leads e senha SMTP | Implementado | migration 35, `segredos.go`, `smtp_service.go` | Validar rotação da chave e recuperação documentada |
| Outbox transacional de e-mails | Implementado; aguardando migration e teste de integração | migration 38, `outbox_postgres.go`, `processador_outbox.go` | Simular SMTP indisponível, reinício e duas instâncias |
| Upload e armazenamento de fotos | Implementado; aguardando migration e homologação S3 | migration 39, adaptadores em `internal/adapters/storage`, `FotosImovel.tsx` | Testar formatos/limites, CDN, exclusão e duas instâncias |
| Meta Lead Ads | Implementado; aguardando migration e homologação no aplicativo Meta | migration 40, `integracao_meta_service.go`, `meta_graph.go`, `processador_integracao_meta.go`, `ConfiguracaoMetaLeads.tsx` | Configurar aplicativo/página reais, validar assinatura, reentrega, token, permissões e lead completo |
| WhatsApp Cloud API inbound | Implementado; aguardando migrations e homologação Meta | migrations 41 e 42, `integracao_whatsapp_service.go`, `processador_integracao_whatsapp.go`, `codec_integracao.go` | Configurar WABA/número reais, validar assinatura, reentrega, conversa, janela e captação única |
| WhatsApp Cloud API outbound e caixa de atendimento | Implementado; aguardando migrations e homologação Meta | migration 43, `whatsapp_graph.go`, codec/tratador da outbox e `frontend/src/pages/WhatsApp` | Validar janela, templates, consentimento, opt-out, reentrega e status fora de ordem |
| Métricas operacionais Prometheus | Implementado; aguardando scraping e alertas | migration 44, `MetricasAplicacao`, middleware HTTP, métricas das filas e painel administrativo | Ativar com token, validar cardinalidade, painéis e alertas de SLO |
| Direitos do titular e governança LGPD | Implementado; aguardando homologação jurídica/operacional | migration 45, `privacidade_service.go`, rotas públicas/administrativas e `CentralPrivacidade` | Validar identidade, decisão, exportação, correção, anonimização, exclusão e auditoria |
| Retenção, anonimização e bloqueio legal | Implementado; aguardando execução controlada em cópia | migration 47, `retencao_service.go` e `PainelRetencao` | Simular política, validar elegibilidade, legal hold, auditoria e irreversibilidade |
| Grupo OLX por VRSync e webhook de leads | Implementado; aguardando homologação do portal | migration 48, domínio canônico, `portais/vrsync`, webhook idempotente e `ConfiguracaoPortais` | Homologar carga completa, remoção, Basic auth, reentrega e isolamento tenant |

## Capacidades explicitamente não entregues

| Capacidade | Situação atual | Próximo passo arquitetural |
|---|---|---|
| Automação avançada de atendimento por WhatsApp | Caixa manual segura implementada; jornadas automáticas não | Modelar regras versionadas, limites por tenant, opt-out automático, aprovação humana e métricas antes de automatizar |
| Antivírus para documentos | Fotos são decodificadas e recodificadas; documentos não são aceitos | Adotar scanner assíncrono antes de permitir PDF ou outros arquivos |
| Propostas e contratos PDF/assinatura | Não implementado | Modelar documentos versionados, templates, assinatura e trilha probatória |
| Aplicativo móvel/PWA offline | Não implementado | Definir casos de uso móveis e sincronização offline antes da interface |
| IA para anúncio ou atendimento | Não implementado | Definir governança, consentimento, custo, avaliação e fallback humano |
| Metas/ranking, locação, nota fiscal/RPA | Não implementado | Tratar como módulos independentes, após estabilizar aquisição e CRM |
| PicPay | Existe apenas um adaptador que retorna “não homologado” | Remover da oferta ou implementar/homologar segundo contrato atual do provedor |
| Outbox para novos canais push | E-mail e WhatsApp usam outbox; VRSync é feed pull e não precisa dela | Exigir novo tratador idempotente por provedor futuro, sem acoplar casos de uso |
| Traces distribuídos e alertas provisionados | Métricas Prometheus existem; traces e regras na infraestrutura não | Definir SLIs/SLOs, exportador de traces e alertas como código na plataforma escolhida |

## Fases e portas de qualidade

### Fase 0 — Fundação segura

Código entregue: configuração tipada, migrations, sessões, tenant, RBAC, CSRF, rate limit, auditoria e segredos. A fase só pode ser aceita após executar migrations numa cópia recente e os testes do backend.

### Fase 1 — Motor de captação

Código entregue: site, catálogo, formulário com consentimento/UTM, webhook genérico, distribuição e CRM. A fase só pode ser aceita após uma jornada completa “visita → lead → atribuição → qualificação → cliente”.

### Fase 2 — Operação imobiliária

Código entregue: imóveis, fotos, agenda, ações rápidas, equipe, dashboard e paginação. A fase só pode ser aceita após teste com dois tenants, múltiplos corretores e volume representativo.

### Fase 3 — Monetização

Código entregue para Stripe, mas o aceite depende do ambiente de homologação do provedor e dos eventos descritos em `OPERACAO.md`. PicPay não está entregue.

### Fase 4 — Escala e integrações

Em andamento. Outbox, imagens local/S3, Meta Lead Ads, WhatsApp inbound/outbound, observabilidade, direitos LGPD, retenção, SEO/domínios e Grupo OLX VRSync foram entregues em código. Permanecem as homologações com provedores e infraestrutura reais, além das portas finais de qualidade.

### Fase 5 — Refinamento final

Refinamento estrutural entregue em código: navegação acessível, ECharts sob demanda, orçamento automático de bundle, dependências compatíveis e documentação de qualidade. Permanecem testes de usabilidade/leitor de tela, responsividade em aparelhos reais, planos de consulta PostgreSQL, métricas de conversão do site e ajustes visuais orientados por evidência.

## Evidências de validação desta revisão

- Frontend: `typecheck`, `lint` e `build` concluídos com sucesso em 07 de agosto de 2026; orçamento de chunks aprovado.
- Bundle: entrada da rota Dashboard em aproximadamente 31 KB e ECharts isolado em aproximadamente 559 KB antes de gzip; indicadores aparecem antes do motor de gráficos.
- Dependências: patches/minors compatíveis atualizados, árvore npm consistente e auditoria de produção com zero vulnerabilidades; React 19, Tailwind 4, TypeScript 7 e date-fns 4 permanecem migrações maiores deliberadamente separadas.
- Backend: arquivos formatados e testes unitários adicionados para segredos, CSRF, resposta de erro, paginação, assinatura/payload Meta e idempotência determinística, mas não executados por restrição explícita do projeto.
- Não há evidência nesta revisão de migrations aplicadas, integração Stripe real, envio SMTP real, login Google real, planos de execução PostgreSQL ou teste entre tenants.

## Ordem de continuidade

1. Aplicar as migrations 20 a 48 sobre cópia recente e executar os testes/builds do backend pelo responsável autorizado.
2. Homologar as jornadas completas das Fases 0 a 2 com dados descartáveis de dois tenants e múltiplos papéis.
3. Homologar SMTP, storage/CDN, Stripe, Meta Lead Ads, WhatsApp e Grupo OLX em ambientes de teste reais.
4. Homologar DNS/TLS, SEO, LGPD, retenção, métricas e recuperação operacional.
5. Executar a matriz de UX, acessibilidade, navegadores, aparelhos e desempenho descrita em `QUALIDADE_FRONTEND.md`.
6. Somente depois promover o release e iniciar módulos opcionais ainda não entregues.

## Estado da Fase 4

A Fase 4 está entregue em código para os canais priorizados, mas seu aceite depende das migrations, homologação Meta/Grupo OLX, DNS e TLS reais, validação dos fluxos LGPD e operação monitorada das filas.
