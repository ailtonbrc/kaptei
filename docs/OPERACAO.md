# Operação e implantação

## Pré-requisitos

- Go compatível com a versão declarada em `backend/go.mod`;
- Node.js e npm compatíveis com Vite 8;
- PostgreSQL com backup validado;
- HTTPS no frontend e na API em produção;
- CLI `migrate` ou o migrador Go do projeto.

## Banco de dados

Faça backup antes de qualquer migration. A partir da pasta `backend`, o operador deve configurar `DATABASE_URL` no ambiente e executar:

```powershell
go run ./cmd/migrador
```

As migrations 21 a 51 formam um conjunto incremental: isolamento tenant, site público, cobrança, auditoria, sessões revogáveis, equipe, idempotência, tokens protegidos, catálogo comercial coerente, índices das listagens, outbox transacional, armazenamento de fotos, integrações, privacidade, portais, métricas de conversão e validação final das regras de domínio. Nunca execute `down` automaticamente em produção.

Antes de aplicar em uma base existente:

- prefira `DATABASE_URL`; durante a transição, o carregador aceita o conjunto completo `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` e `DB_DATABASE`;
- credenciais legadas são escapadas em memória e nunca concatenadas; `DB_SSLMODE` assume `disable` somente com `ENV=development|desenvolvimento|local|dev` e `require` nos demais ambientes;
- não mantenha a compatibilidade legada como padrão de uma implantação nova; migre para `DATABASE_URL` no cofre do ambiente;
- procure e corrija e-mails duplicados quando comparados por `LOWER(email)`; a migration 32 cria unicidade sem diferenciar maiúsculas;
- valide registros sem tenant ou com relacionamentos cruzados antes das constraints 21, 22 e 29;
- confirme que o usuário do banco pode habilitar `pgcrypto`, exigido pela migration 35;
- execute primeiro em uma cópia recente, registre duração e mantenha backup restaurável.
- a migration 35 normaliza o legado conhecido `ACTIVE` para `ATIVO` antes de atualizar a linha do token; isso evita que a constraint `NOT VALID` criada na migration 32 rejeite a própria atualização;
- a migration 50 repete defensivamente essa normalização e valida as constraints da migration 32; qualquer outro legado inválido interrompe a atualização para tratamento explícito;
- a migration 51 atualiza apenas o catálogo comercial para refletir WhatsApp, Grupo OLX/VRSync, privacidade, domínio próprio e SEO já entregues, sem criar novos limites de plano.

### Iniciador local seguro

`executar_backend.bat` orquestra etapas com responsabilidades separadas:

1. `scripts/preparar_ambiente_local.ps1` valida `CONFIG_ENCRYPTION_KEY` e, exclusivamente em ambiente local declarado, gera uma chave de 32 bytes quando ela está ausente. O valor nunca é impresso e permanece em `backend/.env`, ignorado pelo Git.
2. `scripts/migrar_banco_seguro.ps1` identifica as versões atual e esperada. Quando há migrations pendentes, cria `backups/banco/*.dump`, valida o catálogo com `pg_restore --list`, registra SHA-256 e só então executa o migrador.
3. O `.bat` compila e inicia a API somente após ambiente, backup e schema passarem.

O preflight abaixo valida ferramentas, conexão e versões sem gerar chave, backup ou executar migrations:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\migrar_banco_seguro.ps1 -SomenteVerificar
```

Para gerar e validar um backup sem aplicar migrations:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\migrar_banco_seguro.ps1 -SomenteBackup
```

Cada novo dump recebe um nome único. Se `pg_dump` falhar ou `pg_restore --list` recusar o catálogo, o script remove somente o arquivo parcial que corresponda ao padrão e ao diretório controlado; nunca amplia o alvo da limpeza.

Os backups locais são ignorados pelo Git e a pasta `backups/banco` recebe DACL exclusiva para o usuário atual, SYSTEM e Administradores. O módulo persiste somente as regras de acesso, preserva o proprietário e não depende de `SeSecurityPrivilege` ou execução como Administrador. Eles continuam sendo dados sensíveis: criptografe o volume, aplique retenção e copie o backup de produção para mídia/cofre separado antes de uma atualização real. Validar o catálogo não substitui um ensaio periódico de restauração em banco descartável.

## Segredos e primeiro administrador

`JWT_SECRET` e `CONFIG_ENCRYPTION_KEY` são obrigatórios. Gere a chave de criptografia uma vez e mantenha-a no cofre do ambiente:

```powershell
[Convert]::ToBase64String([Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
```

Não troque `CONFIG_ENCRYPTION_KEY` diretamente: a senha SMTP cifrada deixaria de ser legível. Uma rotação deve primeiro decifrar com a chave antiga e persistir novamente com a nova.

Depois das migrations, crie o primeiro administrador somente se ainda não existir nenhum `SUPER_ADMIN`:

```powershell
go run ./cmd/criar-superadmin -nome "Administrador Kaptei" -email "admin@seudominio.com.br"
```

O comando recusa duplicidade e imprime uma senha aleatória uma única vez. Altere-a no primeiro acesso.

## Integração de leads

Gere ou rotacione o token em Configurações. Integrações novas devem enviar `POST /api/webhooks/leads` com `X-Kaptei-Token`. O valor completo aparece apenas na geração; o banco guarda seu SHA-256. A rota legada com token na URL permanece apenas para transição e deixa de funcionar após a rotação.

## Outbox e e-mails

Recuperações de senha e convites gravam a notificação na mesma transação que cria o token ou convite. O payload, inclusive o link temporário, é cifrado com `CONFIG_ENCRYPTION_KEY`. O worker reserva lotes com bloqueio expirável e `SKIP LOCKED`, portanto várias instâncias podem processar a fila sem compartilhar memória.

A entrega é “pelo menos uma vez”: uma queda depois de o SMTP aceitar a mensagem e antes de a outbox ser concluída pode provocar reenvio. O Kaptei usa um `Message-ID` determinístico por evento para favorecer deduplicação no servidor/cliente de e-mail, mas o provedor SMTP precisa ser homologado para esse comportamento. Nunca descreva esse fluxo como “exatamente uma vez”.

Parâmetros de operação:

- `OUTBOX_POLL_INTERVAL`: intervalo de busca;
- `OUTBOX_BATCH_SIZE`: quantidade máxima reservada por ciclo;
- `OUTBOX_MAX_ATTEMPTS`: limite de tentativas armazenado no evento;
- `OUTBOX_LOCK_DURATION`: prazo para outra instância recuperar um trabalho interrompido;
- `OUTBOX_INITIAL_BACKOFF` e `OUTBOX_MAX_BACKOFF`: retentativa exponencial.

Monitore sem expor `payload_protegido`:

```sql
SELECT status, tipo, COUNT(*)
FROM eventos_outbox
GROUP BY status, tipo
ORDER BY status, tipo;

SELECT id, tipo, tentativas, maximo_tentativas, disponivel_em, ultimo_erro, criado_em
FROM eventos_outbox
WHERE status = 'FALHOU'
ORDER BY criado_em DESC
LIMIT 100;
```

Não altere manualmente payloads ou chaves de idempotência. Antes de rotacionar `CONFIG_ENCRYPTION_KEY`, drene todos os eventos `PENDENTE` e `PROCESSANDO`, pois eles foram cifrados com a chave vigente.

## Meta Lead Ads

O aplicativo Meta é uma configuração da instalação; cada imobiliária configura somente sua página e seu token no painel. Para habilitar o canal no servidor:

```env
META_APP_SECRET=SEGREDO_DO_APLICATIVO
META_WEBHOOK_VERIFY_TOKEN=TOKEN_ALEATORIO_COM_PELO_MENOS_32_CARACTERES
META_GRAPH_BASE_URL=https://graph.facebook.com
META_GRAPH_API_VERSION=v25.0
META_HTTP_TIMEOUT=15s
```

Confirme a versão suportada antes da implantação; ela é obrigatoriamente configurável porque versões do Graph API expiram. No aplicativo Meta, registre `GET/POST https://SUA_API/api/webhooks/meta/leads`, assine o campo `leadgen` das páginas e conceda as permissões exigidas pelo produto Lead Ads.

O POST só é aceito com `X-Hub-Signature-256` válido. A resposta HTTP ocorre depois da persistência na fila, antes da consulta ao Graph API. O worker usa token de página cifrado, `appsecret_proof`, bloqueio expirável, retentativa exponencial e uma chave UUID determinística derivada da `leadgen_id`. Assim, uma reentrega ou queda entre criar o lead e concluir o evento não duplica o lead nem avança duas vezes a roleta.

Monitore a fila sem registrar tokens ou respostas completas do provedor:

```sql
SELECT status, provedor, tipo, COUNT(*)
FROM eventos_integracao
GROUP BY status, provedor, tipo
ORDER BY status, provedor, tipo;

SELECT id, conta_id, identificador_externo, tentativas, maximo_tentativas,
       disponivel_em, ultimo_erro, criado_em
FROM eventos_integracao
WHERE status = 'FALHOU'
ORDER BY criado_em DESC
LIMIT 100;
```

A garantia é pelo menos uma vez com efeito idempotente no Kaptei. Um evento sem e-mail e telefone esgota as tentativas e permanece `FALHOU` para análise; ele não é transformado em contato incompleto silenciosamente.

## WhatsApp Cloud API

O WhatsApp reutiliza o aplicativo Meta configurado no ambiente, mas cada tenant possui WABA ID, Phone Number ID e token de usuário de sistema próprios. O callback é `GET/POST https://SUA_API/api/webhooks/meta/whatsapp`. A WABA precisa estar inscrita no aplicativo para que os eventos do campo `messages` sejam enviados.

O webhook valida `X-Hub-Signature-256`, resolve o tenant exclusivamente pelo `phone_number_id` e persiste o evento antes de responder. O conteúdo fica cifrado com `CONFIG_ENCRYPTION_KEY`. O worker deduplica a mensagem pela `wamid` e o lead por uma chave determinística do número do contato; novas mensagens continuam na mesma conversa.

Uma mensagem inbound abre ou renova `janela_atendimento_ate` por 24 horas. `consentimento_marketing` começa falso: iniciar uma conversa permite atendimento dentro da janela, mas não equivale a opt-in para campanhas ou templates promocionais. O envio outbound deve respeitar essa distinção quando for habilitado.

```sql
SELECT status, COUNT(*)
FROM eventos_integracao
WHERE provedor = 'WHATSAPP'
GROUP BY status;

SELECT conta_id, COUNT(*) AS conversas,
       COUNT(*) FILTER (WHERE janela_atendimento_ate > NOW()) AS janelas_abertas
FROM conversas_whatsapp
GROUP BY conta_id;
```

## Armazenamento de fotos

O endpoint autenticado `POST /api/v1/negocios/imoveis/{id}/fotos/upload` recebe `multipart/form-data` nos campos `arquivo` e `is_capa`. Apenas JPEG, PNG e WebP decodificáveis são aceitos. O conteúdo é recodificado como JPEG, sem preservar metadados, e gera uma versão principal e uma miniatura. A URL HTTPS legada continua disponível apenas para imagens que já estejam num CDN confiável.

Desenvolvimento com disco local:

```env
STORAGE_PROVIDER=local
STORAGE_LOCAL_DIR=./data/objetos
STORAGE_PUBLIC_BASE_URL=http://localhost:8080/arquivos
```

Produção com AWS S3, Cloudflare R2 ou serviço compatível:

```env
STORAGE_PROVIDER=s3
STORAGE_PUBLIC_BASE_URL=https://cdn.seudominio.com.br
STORAGE_S3_REGION=REGIAO_DO_PROVEDOR
STORAGE_S3_BUCKET=BUCKET_PRIVADO_OU_ORIGEM_DO_CDN
STORAGE_S3_ENDPOINT=
STORAGE_S3_ACCESS_KEY=
STORAGE_S3_SECRET_KEY=
STORAGE_S3_PATH_STYLE=false
```

Quando a infraestrutura fornecer credenciais por identidade da instância, deixe as chaves vazias. Para R2 ou outro endpoint compatível, configure endpoint, credenciais e `STORAGE_S3_PATH_STYLE` conforme o provedor. O bucket deve aceitar leitura pela CDN/URL pública configurada, mas não escrita pública.

O armazenamento local não é adequado para múltiplas instâncias nem publicação efêmera. Em produção, use S3/CDN com versionamento, criptografia em repouso, política de retenção e CORS restrito. Antes de trocar de provedor, drene eventos `OBJETO_EXCLUIR`; um worker configurado para um provedor não exclui objetos pertencentes a outro.

Os limites de bytes, pixels, dimensões, qualidade e quantidade de processamentos simultâneos são configurados por `IMAGE_*`. A orientação EXIF é aplicada antes de os metadados serem removidos. A decodificação e recodificação reduzem o risco de arquivos disfarçados, mas não substituem uma esteira antivírus caso o produto passe a aceitar PDFs ou outros documentos.

## SEO do catálogo

A API gera `/sitemap.xml` e `/robots.txt` com os sites e imóveis publicados. Se frontend e API forem processos separados, o proxy reverso deve encaminhar exatamente essas duas rotas para a API. Metadados de navegação são atualizados pelo React; pré-renderização ou SSR continua necessário para redes sociais e robôs que não executam JavaScript.

## Métricas de conversão do site

O endpoint público `POST /api/public/sites/{slug}/eventos-conversao` recebe eventos first-party do próprio Kaptei. O payload contém somente chaves UUID, tipo do evento, imóvel opcional e UTM; não contém dados de contato, IP, user-agent ou mensagem. A limitação de leitura pública também protege essa rota.

A migration 49 cria `eventos_conversao_site`, isolada por tenant, com deduplicação por sessão e etapa e validade de 13 meses. O worker remove eventos vencidos conforme:

```env
METRICAS_CONVERSAO_EXPURGO_INTERVAL=24h
```

O painel apresenta sessões únicas dos últimos 30 dias. Esses números são indicadores operacionais e não devem ser usados para cobrança, identificação individual ou decisão automatizada. Monitore volume e abuso sem adicionar labels ou logs com identificadores de sessão.

```sql
SELECT tipo, COUNT(DISTINCT sessao_id) FROM eventos_conversao_site WHERE conta_id = $1 AND criado_em >= NOW() - INTERVAL '30 days' GROUP BY tipo;
```

## Portais imobiliários — Grupo OLX

A migration 48 cria a configuração e a seleção de publicações por tenant. O servidor precisa de `GRUPO_OLX_WEBHOOK_SECRET`, guardado no cofre do ambiente, para autenticar o webhook HTTP Basic. O painel gera uma credencial criptográfica por imobiliária e exibe as URLs completas do feed e do webhook apenas no momento da rotação; o banco persiste somente o hash.

O endpoint VRSync publica uma carga completa. Se um único imóvel selecionado estiver inválido, a geração inteira falha para evitar desativação acidental ou estado parcial no portal. Não contorne o diagnóstico removendo validações. A ativação, os campos obrigatórios, a rotação e o roteiro de homologação estão em `docs/PORTAIS_IMOBILIARIOS.md`.

## Stripe

1. Crie produtos e preços recorrentes no Stripe.
2. Como `SUPER_ADMIN`, informe cada `price_id` na configuração de cobrança do Kaptei.
3. Habilite o Customer Portal no Stripe.
4. Registre `POST https://SUA_API/api/webhooks/stripe` para, no mínimo: `checkout.session.completed`, `invoice.paid`, `invoice.payment_failed`, `customer.subscription.updated` e `customer.subscription.deleted`.
5. Grave o segredo de assinatura em `STRIPE_WEBHOOK_SECRET` e a chave da API em `STRIPE_SECRET_KEY`.
6. Confirme em homologação que reenvios do mesmo evento não duplicam alterações.

## Checklist de produção

- `COOKIE_SECURE=true`, URLs HTTPS e CORS contendo somente origens conhecidas;
- `JWT_SECRET` aleatório, exclusivo e com no mínimo 32 caracteres;
- `CONFIG_ENCRYPTION_KEY` com 32 bytes aleatórios em Base64, armazenada fora do banco e incluída no plano de recuperação;
- `TRUST_PROXY_HEADERS=true` somente atrás de proxy controlado que sobrescreva cabeçalhos encaminhados;
- usuário PostgreSQL com privilégio mínimo e TLS habilitado;
- segredos fornecidos pelo ambiente do serviço, nunca pelo repositório;
- retenção e acesso restrito para logs e `auditoria_eventos`;
- backup, restauração ensaiada e monitoramento de `/health` e `/ready`;
- CSP, HSTS e limites adicionais configurados no proxy reverso;
- política de privacidade revisada com os dados reais da controladora.

## Validações antes da publicação

Frontend:

```powershell
cd frontend
npm.cmd run typecheck
npm.cmd run lint
npm.cmd run build
```

Backend — executar manualmente conforme a política do projeto:

```powershell
cd backend
go test ./...
go build ./cmd/api
```

O script `publicar_completo.ps1` é legado e contém parâmetros específicos do ambiente antigo. Não deve ser usado em produção até ser parametrizado, receber segredos pelo ambiente e ganhar estratégia de rollback/health check.

## Exceção temporária de dependência

Em 06 de agosto de 2026, `react-router-dom` 7.18.2 é a versão estável mais recente publicada no npm. A auditoria aponta `GHSA-qwww-vcr4-c8h2`, corrigida apenas na linha 8.3.0 ainda não publicada para `react-router-dom`. O advisory afeta exclusivamente as APIs instáveis de React Server Components; o Kaptei é uma SPA Vite e não usa RSC. Não reduzir para 7.11.0: essa versão reintroduz advisories aplicáveis a redirects e roteamento cliente. Remova esta exceção assim que uma versão estável corrigida estiver disponível e repita `npm audit --omit=dev`.

## WhatsApp outbound e caixa de atendimento

As migrations 43 e 44 completam, respectivamente, a saída do WhatsApp e a configuração protegida de observabilidade. Mensagens livres são aceitas somente enquanto `janela_atendimento_ate` estiver aberta. Templates exigem consentimento registrado com origem e evidência; a revogação bloqueia novos templates iniciados pela empresa.

O envio usa a outbox transacional e a API oficial do WhatsApp. A garantia externa é **pelo menos uma vez**: uma queda depois de a Meta aceitar a mensagem e antes de o `wamid` ser persistido pode causar reenvio. Não descreva o fluxo como exatamente uma vez. Status recebidos antes da confirmação local ou fora de ordem ficam em `status_mensagens_whatsapp` e convergem sem regressão de `LIDA` para `ENTREGUE`.

```sql
SELECT direcao, status, COUNT(*)
FROM mensagens_whatsapp
GROUP BY direcao, status;

SELECT status, COUNT(*), MIN(criado_em) AS mais_antigo
FROM eventos_outbox
GROUP BY status;
```

## Métricas Prometheus

O superadministrador configura `OBSERVABILIDADE_CONFIG` pela interface. O token precisa ter de 32 a 512 caracteres, é cifrado com a chave mestra e nunca retorna pela API. O endpoint `GET /metrics` responde `404` quando desativado e exige `Authorization: Bearer TOKEN` quando ativo. Alterações podem levar até 30 segundos por causa do cache de segurança.

Métricas principais:

- `kaptei_http_requisicoes_total` e `kaptei_http_duracao_segundos`;
- `kaptei_filas_itens_total` e `kaptei_filas_duracao_processamento_segundos`;
- `kaptei_filas_backlog` e `kaptei_filas_evento_mais_antigo_segundos`;
- `kaptei_banco_conexoes`, além das métricas oficiais do runtime Go e do processo.

Restrinja `/metrics` também no proxy/rede, rotacione o token periodicamente e não use labels com tenant, telefone, e-mail, IDs ou payloads. Alertas mínimos: backlog `FALHOU` maior que zero, idade de `PENDENTE` acima do SLO, taxa HTTP 5xx e saturação do pool.
