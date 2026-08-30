# Integração com portais imobiliários

## Escopo entregue

O Kaptei possui um adaptador isolado para o Grupo OLX baseado no formato VRSync. A integração publica o estoque selecionado da imobiliária e recebe leads de ZAP, Viva Real e canais relacionados sem acoplar o domínio do CRM ao contrato externo.

Componentes principais:

- configuração e seleção por tenant nas tabelas `integracoes_portais` e `publicacoes_portais` (migration 48);
- modelo canônico no domínio, gerador VRSync no adaptador `portais/vrsync` e persistência PostgreSQL separada;
- feed público completo em `GET /api/public/portais/grupo-olx/{token}/vrsync.xml`;
- webhook autenticado em `POST /api/webhooks/portais/grupo-olx/{token}/leads`;
- painel administrativo em Configurações > Portais imobiliários.

## Segurança e isolamento

Defina no servidor uma credencial compartilhada com o Grupo OLX:

```env
GRUPO_OLX_WEBHOOK_SECRET=SEGREDO_ALEATORIO_COM_PELO_MENOS_16_CARACTERES
```

O webhook exige autenticação HTTP Basic. O nome do usuário informado pelo provedor não é usado como autorização; a senha é comparada em tempo constante com o segredo do ambiente. O token de URL identifica exclusivamente o tenant, é gerado com entropia criptográfica e somente seu SHA-256 é persistido. O valor completo e as URLs aparecem uma única vez após a rotação.

A rotação invalida imediatamente o feed e o webhook anteriores. Não registre essas URLs completas em tickets, logs, analytics ou ferramentas públicas. O proxy deve preservar o cabeçalho `Authorization`, aplicar HTTPS e nunca armazenar o corpo dos leads em logs de acesso.

## Regra de carga completa

O feed VRSync é uma fotografia completa do estoque selecionado. No processamento em lote do Grupo OLX, a ausência de um anúncio antes publicado pode desativá-lo. Por isso, o Kaptei opera em modo *fail closed*: se qualquer imóvel selecionado estiver inválido, nenhum XML parcial é entregue e a rota responde indisponível.

Antes de ativar, o diagnóstico precisa validar todos os anúncios selecionados. Entre os requisitos estão:

- site público publicado e origem pública válida;
- imóvel ativo, slug, finalidade, tipo e preço coerentes;
- título entre 10 e 100 caracteres e descrição entre 50 e 3.000 caracteres, sem HTML;
- CEP, estado, cidade, bairro e área compatível com o tipo do imóvel;
- ao menos uma foto pública HTTPS, JPEG, com tamanho conhecido de até 7 MB;
- tipo de publicação permitido pelo contrato comercial do portal.

O Kaptei não presume formato de foto pela extensão. Fotos legadas sem metadados devem ser reprocessadas pelo fluxo oficial de upload antes da publicação.

## Recebimento de leads

O endpoint aceita o JSON oficial e tolera campos adicionais para manter compatibilidade futura, mas limita o corpo a 64 KB e exige exatamente um objeto JSON. A autenticação é validada antes de consultar o tenant.

- `originLeadId` gera uma chave UUID determinística e impede duplicação por reentrega;
- leads comuns exigem `clientListingId` UUID pertencente ao estoque selecionado do tenant;
- leads `MCMV_OLX` podem não informar imóvel;
- o contato entra no motor existente de captação e distribuição;
- o recebimento não presume consentimento de marketing.

Respostas esperadas: `202` para aceite idempotente, `400` para payload inválido, `401` para credencial/token inválido e `503` quando a dependência local estiver indisponível. O provedor deve considerar apenas respostas `2xx` como sucesso.

## Ativação e homologação

1. Aplique a migration 48 em uma cópia recente e valide o rollback apenas em ambiente descartável.
2. Configure `APP_PUBLIC_URL` e `GRUPO_OLX_WEBHOOK_SECRET` no cofre do ambiente.
3. Em Configurações, preencha o contato, selecione os imóveis e corrija todos os erros do diagnóstico.
4. Rotacione a credencial de URL e copie feed e webhook para um cofre seguro.
5. Cadastre o feed e a autenticação Basic com o time do Grupo OLX.
6. Valide o XML completo, atualização e remoção controlada de um anúncio em ambiente de homologação.
7. Envie o mesmo `originLeadId` duas vezes e confirme a existência de apenas um lead.
8. Teste lead normal vinculado, lead MCMV sem imóvel, segredo incorreto, token antigo e campos adicionais.
9. Ative a integração somente após a aprovação comercial/técnica do portal.

Não há sincronização incremental nem envio ativo para uma API externa nesta versão. O portal consulta o feed completo; o webhook cobre apenas entrada de leads.

## Validação do backend pelo operador

Por política do projeto, os comandos abaixo devem ser executados pelo responsável em ambiente local/homologação, não automaticamente pelo agente:

```powershell
cd backend
go test ./internal/adapters/portais/vrsync ./internal/core/services
go test ./...
```

