# Checklist de homologação

## Segurança e tenancy

- autenticar por senha e Google sem JWT acessível ao JavaScript;
- confirmar `HttpOnly`, `Secure` e `SameSite=Lax` no cookie de produção;
- validar logout e expiração da sessão;
- confirmar revogação imediata da sessão e invalidação das demais sessões após troca de senha;
- tentar acessar, alterar e excluir IDs pertencentes a outra conta e esperar bloqueio;
- testar limitação de login, recuperação, webhook de lead e formulário público;
- confirmar que recuperação não revela se o e-mail existe e que o token só funciona uma vez.
- convidar, aceitar, cancelar e inativar um corretor; validar o limite do plano e impedir auto-inativação do gestor;
- interromper o SMTP, solicitar recuperação/convite, confirmar evento `PENDENTE`, restaurar SMTP e validar entrega com retentativa sem duplicidade;
- iniciar duas instâncias da API e confirmar que um mesmo evento da outbox é entregue apenas uma vez;

## Captação e CRM

- publicar site e imóveis, pesquisar, paginar e abrir detalhes em celular e desktop;
- enviar JPEG, PNG e WebP, confirmar recodificação, miniatura, limites de bytes/pixels e bloqueio de conteúdo falso;
- enviar foto vertical com orientação EXIF e confirmar que principal e miniatura permanecem na orientação correta;
- excluir foto e imóvel, confirmar eventos `OBJETO_EXCLUIR` concluídos e ausência dos objetos no storage;
- validar URL legada somente com HTTPS e confirmar que o bucket/CDN não permite escrita pública;
- validar `/sitemap.xml`, `/robots.txt`, URLs canônicas e metadados usando o domínio real;
- enviar lead com consentimento, origem e UTM; verificar distribuição e auditoria;
- abrir site e imóvel, iniciar e concluir formulário e clicar em WhatsApp/telefone; conferir as etapas no painel de conversão;
- confirmar deduplicação da mesma etapa por sessão e separação completa entre duas contas;
- inspecionar o payload de `/eventos-conversao` e confirmar ausência de IP, user-agent, nome, e-mail, telefone e mensagem;
- simular evento vencido em banco descartável e validar o expurgo sem remover eventos ainda válidos;
- atribuir, qualificar, descartar e converter lead sem duplicar cliente;
- criar, editar e excluir agendamentos dentro do tenant.
- criar mais de 100 clientes, leads e imóveis e validar paginação, busca, filtros e carregamento incremental sem misturar tenants;
- rotacionar o token de lead, confirmar que o anterior falha e que o novo funciona somente no cabeçalho;
- configurar uma página Meta real, verificar o callback e receber um Lead Ads de teste com nome, e-mail e telefone;
- adulterar o corpo mantendo a assinatura Meta original e confirmar rejeição; testar também assinatura ausente;
- reenviar a mesma `leadgen_id`, interromper o worker entre a criação do lead e a conclusão do evento e confirmar um único lead e um único avanço da roleta;
- revogar o token da página e confirmar retentativas/estado `FALHOU`; cadastrar novo token sem a API devolver o segredo anterior;
- configurar WABA e Phone Number ID reais, inscrever a WABA e receber texto de um número de teste;
- reenviar a mesma `wamid` e enviar várias mensagens do mesmo contato; confirmar uma conversa, uma mensagem por `wamid` e um único lead;
- confirmar que mensagem e payload persistidos estão cifrados e que logs não apresentam texto, telefone ou token;
- validar abertura/renovação da janela de 24 horas e `consentimento_marketing=false` por padrão;

## Cobrança

- iniciar Checkout apenas como gestor ou corretor solo;
- validar assinatura do webhook e rejeitar payload adulterado;
- reenviar o mesmo evento e confirmar idempotência;
- testar pagamento aprovado, falha, inadimplência, cancelamento e retorno ao portal;
- confirmar que a aplicação nunca recebe número de cartão ou CVV.

## Operação


## LGPD, retenção e domínios

- abrir uma solicitação pública para cada direito suportado e confirmar protocolo, limitação de taxa e ausência de enumeração de dados;
- verificar identidade, registrar decisão fundamentada e executar exportação, correção, anonimização ou exclusão conforme o caso;
- confirmar que exclusão/anonimização não ocorre antes da verificação e aprovação do controlador;
- criar bloqueio legal e confirmar que a política de retenção não processa o recurso protegido;
- simular a retenção, conferir contadores e só então executar um lote controlado com auditoria;
- configurar domínio próprio, publicar o desafio DNS, validar isolamento entre tenants e ativar somente após DNS e TLS corretos;
- verificar canonical, Open Graph, sitemap, robots, pré-renderização e redirecionamento no domínio final.

## Grupo OLX e VRSync

- configurar `GRUPO_OLX_WEBHOOK_SECRET` no cofre e confirmar que o proxy preserva `Authorization` sem registrar payloads;
- selecionar inventário válido e confirmar que o XML contém todos os anúncios esperados;
- introduzir um anúncio inválido e confirmar indisponibilidade total do feed, sem carga parcial;
- homologar inclusão, atualização e remoção controlada de um anúncio com o Grupo OLX;
- rotacionar o token e confirmar que feed e webhook antigos deixam de funcionar;
- receber lead comum vinculado ao `clientListingId` do tenant e lead MCMV sem imóvel;
- reenviar o mesmo `originLeadId` e confirmar um único lead e uma única distribuição;
- testar segredo Basic incorreto, token de outro tenant, imóvel não selecionado, payload acima de 64 KB e campos adicionais válidos.

## UX, acessibilidade e desempenho

- validar toda a navegação por teclado, incluindo link “Pular para o conteúdo”, drawer móvel, menu da conta e modais;
- testar fechamento do drawer por Escape, retorno do foco ao gatilho e ausência de elementos focáveis quando fechado;
- validar anúncios de carregamento/erro e resumos textuais dos gráficos com NVDA e VoiceOver;
- verificar contraste, foco visível, zoom de 200% e larguras de 320, 375, 768, 1024 e 1440 px;
- medir LCP, INP e CLS no site público e nas rotas internas em rede móvel simulada e aparelho real;
- confirmar que o build respeita os orçamentos definidos em `docs/QUALIDADE_FRONTEND.md`;
- validar que o Dashboard exibe indicadores antes de carregar o motor ECharts e não causa deslocamento de layout;
- executar jornada completa em Chrome/Edge, Firefox e Safari atuais.
- aplicar migrations sobre cópia recente da base;
- antes das migrations 20–51, confirmar ausência de e-mails duplicados e relacionamentos entre tenants; tratar valores de domínio desconhecidos sem desabilitar constraints;
- após a migration 50, verificar `convalidated=true` para todas as constraints `*_valido_ck`, `*_valida_ck`, `*_positivos_ck` e `usuarios_conta_obrigatoria_ck`;
- após a migration 51, confirmar que o catálogo não apresenta WhatsApp e Grupo OLX/VRSync simultaneamente como recursos entregues e ausentes;
- validar logs JSON e correlação por `X-Request-ID`;
- provocar erro controlado e confirmar resposta sem stack trace ao usuário;
- salvar SMTP, confirmar que a API mascara a senha e que o banco persiste valor com prefixo `enc:v1:`;
- verificar `/health` com banco indisponível e `/ready` retornando indisponibilidade;
- testar backup/restauração e retorno da versão anterior;
- executar typecheck/lint/build frontend e testes/build backend pelo responsável autorizado.
- monitorar eventos `FALHOU` sem consultar ou registrar `payload_protegido`.
- em produção, validar cache imutável, CORS, criptografia, versionamento e política de ciclo de vida do bucket.

## WhatsApp outbound e observabilidade

- enviar texto dentro da janela de 24 horas e confirmar `PENDENTE → ENVIADA → ENTREGUE → LIDA` na caixa de atendimento;
- tentar enviar texto fora da janela e confirmar bloqueio; registrar consentimento com origem/evidência e enviar template aprovado;
- revogar o consentimento e confirmar bloqueio de novos templates iniciados pela empresa;
- entregar status antes da confirmação local e em ordem diferente; confirmar convergência sem regressão de `LIDA` para `ENTREGUE`;
- interromper a API após aceitação pelo Graph e antes da confirmação local; verificar a possibilidade documentada de reenvio;
- autenticar como corretor de equipe e confirmar acesso somente às conversas vinculadas aos seus leads;
- confirmar que `/metrics` retorna `404` desativado, `401` sem token e métricas sem dados pessoais com token válido;
- provocar retentativa e falha definitiva nas filas, validar contadores, backlog e idade do evento mais antigo;
- validar alertas para HTTP 5xx, backlog `FALHOU`, atraso de `PENDENTE` e saturação do pool PostgreSQL.
