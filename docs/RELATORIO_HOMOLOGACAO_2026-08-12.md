# Relatório de homologação — 12/08/2026

## Resultado desta rodada

- `npm.cmd run typecheck`: aprovado;
- `npm.cmd run lint`: aprovado;
- `npm.cmd run build`: aprovado;
- `git diff --check`: aprovado;
- `gofmt -d backend`: aprovado, sem arquivos com sintaxe/formatação divergente;
- PostgreSQL: schema na migration 19, `dirty=false`; migrations 20–51 pendentes;
- nomenclatura da funcionalidade padronizada como “métricas de conversão do site” em código e documentação;
- PostgreSQL local: porta 5432 disponível;
- API e frontend locais: portas 8080 e 5173 indisponíveis durante a verificação;
- frontend compilado: prévia temporária aprovada na porta 4173, com shell React respondendo `200` em `/`, `/login`, `/cadastro`, `/politica-de-privacidade` e `/s/site-inexistente`;
- ativos compilados: três recursos JS/CSS encontrados no HTML e servidos com conteúdo;
- automação visual pelo navegador: bloqueada pelas ACLs do ambiente Windows antes da interação com a página.
- scripts PowerShell: sintaxe aprovada no Windows PowerShell 5.1 e codificação UTF-8 com BOM;
- módulo de ambiente: testes isolados aprovados com arquivo temporário e credenciais fictícias;
- preflight de migrations: aprovado, detectando schema 19 e destino 51 sem alterar o banco.
- `CONFIG_ENCRYPTION_KEY`: gerada localmente, validada com 32 bytes e armazenada somente no `backend/.env` ignorado pelo Git;
- backup pré-migration: dump custom criado para o schema 19, catálogo validado com `pg_restore --list` e SHA-256 conferido;
- backup e manifesto: ignorados pelo Git;
- ACL do backup: herança ampla removida; acesso limitado ao usuário atual, SYSTEM e Administradores.
- fluxo de backup: nomes únicos e limpeza estritamente limitada ao dump parcial quando a geração ou a validação do catálogo falha.

O backend não foi compilado nem testado pelo Codex, conforme a política do projeto.

## Correções encontradas durante a auditoria

1. O dashboard tentava adicionar `conversao_site` antes de criar o mapa da resposta. A composição foi reordenada e ganhou testes para gestor e corretor de equipe.
2. A inserção de eventos ignorava somente conflito da chave do evento, mas a deduplicação por sessão e etapa também possui índice único. A gravação agora usa `ON CONFLICT DO NOTHING` para cobrir ambas as garantias.
3. O identificador da sessão do navegador era compartilhado entre slugs visitados na mesma aba. Sessão e atribuição UTM agora são isoladas por site da imobiliária.
4. Os iniciadores locais não verificavam migrations e reorganizavam dependências a cada execução. Agora validam ferramentas e `.env`, interrompem em erro, aplicam migrations antes da API e não executam binário anterior quando a compilação falha.
5. A primeira proteção do diretório de backup reaplicava um descritor completo e, em uma pasta já endurecida, o Windows exigia `SeSecurityPrivilege`. O módulo agora persiste exclusivamente a DACL, preserva proprietário e auditoria e foi aprovado na pasta real sem elevação administrativa.
6. Na primeira aplicação real, a migration 35 tentou atualizar o token da conta ainda marcada como `ACTIVE`; a constraint `NOT VALID` da migration 32 rejeitou a nova versão da linha e o migrador deixou `35|dirty`. O SQL foi corrigido para normalizar o legado antes da atualização. Como a transação da migration 35 havia revertido integralmente, o marcador foi recuperado de forma condicionada para `34|false`, após validar backup, ausência de efeitos parciais da 35 e presença integral da 34. As migrations 35–51 foram então ensaiadas em transação única com `ROLLBACK` e aprovadas.

## Estado atual do ambiente

O arquivo `backend/.env` usa o conjunto legado `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` e `DB_DATABASE`. O carregador passou a aceitar esse conjunto de forma transitória, montando a URL em memória com escape seguro. Como `ENV` declara desenvolvimento local, `APP_PUBLIC_URL` e `STORAGE_PUBLIC_BASE_URL` usam padrões localhost quando ausentes.

As configurações obrigatórias do ambiente local estão agora estruturalmente válidas. `CONFIG_ENCRYPTION_KEY` foi gerada automaticamente sem exibição; o `JWT_SECRET` atende ao tamanho mínimo e nenhum valor foi exposto neste relatório.

Produção continua exigindo URLs explícitas, TLS e segredos fornecidos pelo cofre do ambiente. Segredos não devem ser copiados para scripts, commits ou relatórios.

## Pré-validação do banco atual

Consultas agregadas confirmaram ausência de e-mails duplicados, números negativos e relacionamentos cruzados entre tenants nas tabelas existentes. `pgcrypto` está disponível. Foi encontrado um único status de plano legado `ACTIVE`, sem identificar a conta; a migration 35 agora o normaliza antes da primeira atualização da linha e a migration 50 valida todas as constraints de domínio da migration 32.

A migration 51 corrige o catálogo da migration 36, que ainda classificava WhatsApp e portais externos como ausentes. Nenhuma migration foi aplicada nesta rodada e nenhum registro do banco foi alterado.

## Próxima execução autorizada pelo responsável

Com backup/restauração previamente ensaiados, executar em terminais separados:

```powershell
./executar_backend.bat
./executar_frontend.bat
```

O iniciador do backend validará a chave local, criará e validará um novo backup, aplicará as migrations 20–51 e somente iniciará a API se todas as etapas e a compilação forem concluídas. Em produção, mantenha também uma cópia externa e um ensaio de restauração. Confirmar então:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/ready
```

## Aceite runtime ainda pendente

- login, sessão, RBAC e isolamento entre duas contas;
- site público, catálogo, formulário e criação do lead;
- eventos de conversão e painel dos últimos 30 dias;
- WhatsApp inbound/outbound em ambiente Meta homologado;
- Stripe e Grupo OLX/VRSync com credenciais de homologação;
- domínio, sitemap, robots e pré-renderização;
- Central LGPD, bloqueios e retenção em banco descartável;
- teclado, leitor de tela, 320–1440 px, navegadores e aparelho real;
- testes e build do backend executados pelo responsável.

Este relatório não declara homologação completa enquanto esses itens não tiverem evidência runtime.
