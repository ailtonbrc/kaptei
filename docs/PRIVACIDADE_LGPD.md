# Privacidade e direitos do titular

## Objetivo

O módulo de privacidade registra e conduz solicitações de titulares de dados vinculadas ao site público de cada conta. O fluxo foi desenhado para impedir que uma requisição anônima provoque exportação, anonimização ou exclusão automática.

## Fluxo operacional

1. O titular abre a solicitação em `/s/{slug}/privacidade`.
2. O sistema devolve um protocolo e armazena nome, contato e detalhes criptografados.
3. Gestor, corretor solo ou superadministrador acessa a Central de Privacidade.
4. O operador confirma a identidade por canal previamente conhecido e registra método e evidência.
5. O controlador registra decisão, fundamento legal e observação.
6. Somente uma solicitação aprovada pode gerar exportação ou executar o tratamento correspondente.
7. Toda transição gera evento imutável na trilha da solicitação e as mutações autenticadas também passam pela auditoria HTTP geral.

## Direitos suportados

- Confirmação, acesso, portabilidade e informação sobre compartilhamento: geram arquivo JSON após aprovação.
- Correção e bloqueio: o operador realiza o ajuste no CRM e confirma a conclusão na Central.
- Revogação: remove consentimentos LGPD e de marketing do WhatsApp dentro da conta.
- Anonimização: elimina conversas associadas e remove identificadores dos registros de CRM, preservando apenas dados operacionais não identificáveis.
- Exclusão: remove conversas, agendamentos vinculados, leads e clientes localizados pelo contato verificado, sempre dentro do tenant.

## Controles de segurança

- Nome, contato, detalhes e evidência de identidade ficam criptografados em repouso.
- E-mail e telefone possuem apenas hash normalizado para pesquisa da própria solicitação.
- A listagem administrativa não revela PII; os dados são descriptografados somente ao abrir uma solicitação autorizada.
- A exportação usa `Cache-Control: no-store` e nunca é enviada automaticamente ao contato público.
- Operações destrutivas exigem identidade verificada, decisão aprovada e confirmação explícita na interface.
- Consultas, atualizações e exclusões incluem `conta_id`; corretores de equipe não acessam o módulo.

## Métricas de conversão do site público

O site registra eventos estritamente necessários para medir visitas, visualizações de imóveis, início de formulário, contato enviado e cliques nos canais de atendimento. Cada sessão recebe um UUID pseudônimo mantido em `sessionStorage`; não são armazenados IP, user-agent, nome, e-mail, telefone ou mensagem nessa tabela. Parâmetros UTM podem ser associados à sessão para atribuição da campanha.

Os eventos são isolados por `conta_id`, deduplicados por sessão e etapa e expiram em 13 meses. Um processo interno remove registros vencidos no intervalo configurado por `METRICAS_CONVERSAO_EXPURGO_INTERVAL`. Essas métricas não substituem o consentimento exigido para o envio do formulário nem autorizam campanhas de marketing.

A controladora deve documentar sua avaliação da base legal aplicável, assegurar transparência e respeitar oposição e demais direitos quando cabíveis. O identificador pseudônimo não deve ser combinado com dados do CRM para criar perfil individual.

## Retenção e exceções legais

A política de retenção é configurada por conta e nasce desativada. Ela somente pode ser ativada com fundamento legal documentado, prazos entre 30 e 3.650 dias e lote entre 1 e 1.000 registros por tipo.

A execução considera exclusivamente:

- leads com status `DESCARTADO` e prazo vencido;
- clientes com status `PERDIDO` e prazo vencido;
- clientes sem lead ativo ou recente e sem atendimento futuro agendado ou confirmado;
- registros sem bloqueio legal vigente.

Antes da execução, a Central apresenta a quantidade de candidatos e de bloqueios. O operador precisa digitar exatamente `ANONIMIZAR DADOS EXPIRADOS`. A política é novamente bloqueada e relida dentro da transação, evitando que uma execução concorrente use parâmetros antigos. Cada lote registra operador, quantidades, fundamento e data.

A anonimização remove identificadores do CRM e o conteúdo pessoal das conversas vinculadas. A exclusão das conversas é intencional porque número, nome, mensagens e evidências de consentimento formam uma unidade pessoal; a trilha agregada do lote permanece em `execucoes_retencao` sem conservar esses dados.

Bloqueios legais podem ser criados para um lead ou cliente, com motivo obrigatório e vigência opcional. Eles devem ser usados quando houver litígio, auditoria, investigação ou obrigação de conservação compatível com o artigo 16 da LGPD.

Rotas administrativas protegidas:

- `GET|PUT /api/v1/privacidade/retencao/politica`;
- `GET /api/v1/privacidade/retencao/relatorio`;
- `POST /api/v1/privacidade/retencao/execucao`;
- `GET|POST /api/v1/privacidade/retencao/bloqueios`;
- `DELETE /api/v1/privacidade/retencao/bloqueios/{id}`.

## Homologação do backend

As validações de backend devem ser executadas pelo responsável do ambiente:

```powershell
cd "D:\Desenvolvimento MSDev\kaptei\backend"
go test ./...
go build ./cmd/api
```

Validar em banco descartável as migrações `000045_criar_solicitacoes_titular` e `000047_criar_politicas_retencao`, além dos cenários de acesso, rejeição, revogação, anonimização, exclusão, bloqueio legal e lotes concorrentes. Conferir especialmente que registros de outra conta e registros bloqueados permanecem intocados.
