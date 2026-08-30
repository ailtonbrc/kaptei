# Memória de Projeto - Kaptei CRM (registro histórico)

> Revisão de 06 de agosto de 2026: este arquivo registra o estado percebido em julho e não é uma fonte de aceite. A afirmação antiga sobre webhooks próprios de WhatsApp/Meta Ads não corresponde ao código atual: existe um webhook genérico e protegido para entrada de leads, mas os adaptadores assinados de cada provedor ainda não foram implementados. O botão reutilizável de agendamento em clientes e imóveis foi entregue. Para o estado verificável, consulte `docs/MATRIZ_RASTREABILIDADE.md`.

**Data da última modificação:** 06 de Julho de 2026 (Madrugada)
**Status Geral:** O backend e frontend estão estáveis e rodando localmente. A API já conta com Banco de Dados em PostgreSQL, Roteamento de Leads, Webhooks de WhatsApp/MetaAds, Assinaturas (PicPay/Stripe) e, agora, Dashboard Premium e Calendário.

---

## O Que Foi Feito Hoje (Última Sessão)

1. **Correção do Deploy (IIS CORS 401):** 
   - Diagnosticado e corrigido o bloqueio de CORS causado pelo IIS no Servidor VPS (`192.168.100.120`). Configurado o IIS para não sobrescrever os erros `401 Unauthorized` da API Go (`existingResponse="PassThrough"`).

2. **Saneamento de Acesso (Dev):**
   - Credenciais preenchidas automaticamente foram removidas do frontend. Contas administrativas devem ser provisionadas por um comando seguro e auditável.

3. **Validação do Kanban (Pipeline de Vendas):** 
   - Conferido e validado que as colunas utilizam um fundo neutro, sendo as cores aplicadas exclusivamente ao cabeçalho (Status) e aos cartões (Oportunidades), conforme a regra estipulada pelo usuário.

4. **Fase 3 Premium - Dashboard Analítico Avançado:**
   - [Backend] Criada rota `/api/v1/negocios/dashboard/premium`.
   - [Backend] Feita a extração e agrupamento de dados do banco para Funil de Conversão e Origem de Leads.
   - [Frontend] Tela `Dashboard.tsx` reescrita com **Apache ECharts** e **Tremor** para um visual totalmente Enterprise.

5. **Fase 3 Premium - Módulo de Calendário (Agendamentos):**
   - [Backend] Criada a *migration* para a tabela `agendamentos`.
   - [Backend] Criado todo o Core Domain, Repositories, Services e Handlers para CRUD de Agendamentos.
   - [Frontend] Instaladas as bibliotecas `react-big-calendar` e `date-fns`.
   - [Frontend] Criada a tela interativa de calendário (`Agendamentos.tsx`) e adicionada a navegação lateral.

---

## Próximos Passos (Para Amanhã / Próxima Sessão)

Quando voltarmos a trabalhar neste projeto, devemos seguir com os pontos abaixo:

1. **Testes do Backend:** Reiniciar o arquivo `executar_backend.bat` para garantir que a *migration* do calendário `agendamentos` seja criada no banco PostgreSQL.
2. **Homologar a UI do Dashboard e Calendário:** Navegar pelas novas telas no frontend (`/app` e `/app/agendamentos`) para validar a renderização dos ECharts e do react-big-calendar.
3. **Botão de Ação Rápida:** Adicionar um botão "Agendar Visita" diretamente no Perfil do Lead ou do Imóvel, que abra um Modal para interagir rapidamente com a API do Calendário.
4. **Proteção e Versionamento de Código:** Fazer o commit (`git add . && git commit`) de tudo o que foi implementado hoje para o repositório `ailtonbrc/kaptei` no GitHub.

---

*Fim da anotação. Excelente trabalho hoje! Descanse bem.*
