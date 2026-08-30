# SEO, pré-renderização e domínios personalizados

## Responsabilidades

O frontend React mantém a experiência interativa. O pacote Go `cmd/static-server` serve os artefatos do Vite e pré-renderiza as páginas públicas para mecanismos de busca e compartilhamento social. A API continua responsável por dados, publicação do site e validação de domínio.

O servidor estático gera no primeiro HTML:

- título, descrição e URL canônica;
- Open Graph;
- JSON-LD de imobiliária e anúncio imobiliário;
- conteúdo semântico e links internos;
- `robots.txt` e `sitemap.xml` por domínio.

As rotas públicas da plataforma permanecem em `/s/{slug}` e `/s/{slug}/imoveis/{imovel}`. Em domínio personalizado, as rotas públicas são `/`, `/imoveis/{imovel}` e `/privacidade`.

## Publicação

Gerar os artefatos do frontend:

```powershell
cd "D:\Desenvolvimento MSDev\kaptei\frontend"
npm.cmd run build
```

O responsável pelo ambiente pode compilar o servidor estático com todo o pacote:

```powershell
cd "D:\Desenvolvimento MSDev\kaptei\backend"
go build -o servidor_web.exe ./cmd/static-server
```

Exemplo de inicialização:

```powershell
.\servidor_web.exe `
  -port 8013 `
  -dir "D:\caminho\frontend\dist" `
  -api "http://127.0.0.1:8080/api" `
  -public-url "https://app.seudominio.com.br"
```

`-api` é usado pelo pré-renderizador para consultar dados públicos. `-public-url` define a origem canônica da plataforma; quando vazio, o host validado da requisição é usado. Em domínio personalizado ativo, a URL canônica sempre utiliza o domínio solicitado.

## Topologia de rede e TLS

O `static-server` escuta HTTP interno. Em produção, um reverse proxy ou serviço de borda deve:

1. terminar TLS com certificado válido para cada domínio;
2. preservar o cabeçalho `Host` original;
3. encaminhar páginas e ativos ao `static-server`;
4. encaminhar `/api` ao backend sem remover indevidamente o prefixo esperado;
5. limitar tamanho, tempo e frequência de requisições na borda;
6. redirecionar HTTP para HTTPS.

O domínio do cliente precisa apontar por `A`, `AAAA` ou `CNAME` para essa borda. Esse apontamento é independente da prova de propriedade por TXT.

## Ativação segura do domínio

1. O gestor informa o hostname na área de configurações.
2. O Kaptei gera um token aleatório e instrui a criar `_kaptei-verificacao.<hostname>`.
3. O valor TXT esperado é `kaptei-verificacao=<token>`.
4. O gestor solicita a verificação.
5. Apenas uma resposta DNS que corresponda ao token atualmente persistido ativa o domínio.
6. O site precisa estar publicado para a resolução pública retornar a conta.

A comparação do token ocorre junto à atualização condicional no banco. Assim, uma resposta DNS atrasada não consegue ativar um hostname ou token substituído durante a consulta.

## Checklist de homologação

- confirmar certificado, redirecionamento HTTPS e preservação do `Host`;
- abrir `/`, `/imoveis/{slug}` e `/privacidade` no domínio personalizado;
- abrir `/s/{slug}` na origem da plataforma;
- validar canônica, Open Graph e JSON-LD no HTML recebido, sem depender da execução de JavaScript;
- confirmar que `sitemap.xml` contém apenas URLs publicadas daquele domínio;
- confirmar que `robots.txt` referencia o sitemap do mesmo host;
- testar hostname inválido, domínio não verificado, token alterado durante a verificação e site despublicado;
- conferir que um domínio de outra conta nunca resolve para o tenant atual.
