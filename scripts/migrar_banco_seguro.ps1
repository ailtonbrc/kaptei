param(
    [string]$DiretorioRaiz = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path,
    [switch]$SomenteVerificar,
    [switch]$SomenteBackup
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Import-Module (Join-Path $PSScriptRoot 'Kaptei.Ambiente.psm1') -Force

$raiz = (Resolve-Path -LiteralPath $DiretorioRaiz -ErrorAction Stop).Path
$backend = (Resolve-Path -LiteralPath (Join-Path $raiz 'backend') -ErrorAction Stop).Path
$caminhoEnv = (Resolve-Path -LiteralPath (Join-Path $backend '.env') -ErrorAction Stop).Path
$migrations = (Resolve-Path -LiteralPath (Join-Path $backend 'db\migrations') -ErrorAction Stop).Path
$configuracao = Get-ConfiguracaoEnv $caminhoEnv
$conexao = Get-ConexaoPostgreSQL $configuracao

$psql = Get-FerramentaPostgreSQL 'psql.exe'
$pgDump = Get-FerramentaPostgreSQL 'pg_dump.exe'
$pgRestore = Get-FerramentaPostgreSQL 'pg_restore.exe'
$go = (Get-Command go -ErrorAction Stop).Source

$versoes = Get-ChildItem -LiteralPath $migrations -Filter '*.up.sql' -File | ForEach-Object {
    [int]$_.BaseName.Split('_')[0]
}
$versaoEsperada = ($versoes | Measure-Object -Maximum).Maximum
if ($null -eq $versaoEsperada) {
    throw 'Nenhuma migration foi encontrada.'
}

$env:PGPASSWORD = $conexao.Senha
if (-not [string]::IsNullOrWhiteSpace($conexao.ModoSSL)) {
    $env:PGSSLMODE = $conexao.ModoSSL
}
$consultaEstadoSchema = 'SELECT version::text || ''|'' || dirty::text FROM schema_migrations LIMIT 1;'

function Invoke-ConsultaEscalar {
    param([Parameter(Mandatory = $true)][string]$Sql)
    $resultado = & $psql -X --no-password --host $conexao.Host --port $conexao.Porta `
        --username $conexao.Usuario --dbname $conexao.Banco --tuples-only --no-align --command $Sql
    if ($LASTEXITCODE -ne 0) {
        throw 'Falha ao consultar o estado do PostgreSQL.'
    }
    return (($resultado | Select-Object -First 1) -as [string]).Trim()
}

try {
    $tabelaMigrations = Invoke-ConsultaEscalar "SELECT COALESCE(to_regclass('public.schema_migrations')::text, '');"
    if ([string]::IsNullOrWhiteSpace($tabelaMigrations)) {
        $versaoAtual = 0
        $sujo = $false
    } else {
        $estado = Invoke-ConsultaEscalar $consultaEstadoSchema
        $partes = $estado.Split('|', 2)
        $versaoAtual = [int]$partes[0]
        $sujo = [bool]::Parse($partes[1])
    }

    if ($sujo) {
        throw "O schema está marcado como dirty na versão $versaoAtual. Corrija-o antes de continuar."
    }
    if ($versaoAtual -gt $versaoEsperada) {
        throw "O banco está na versão $versaoAtual, superior ao código local ($versaoEsperada)."
    }
    if ($versaoAtual -eq $versaoEsperada) {
        Write-Host "Schema já atualizado na versão $versaoEsperada." -ForegroundColor Green
        exit 0
    }
    if ($SomenteVerificar) {
        Write-Host ("Pré-validação concluída: schema {0}, destino {1}, backup e migrations pendentes." -f $versaoAtual, $versaoEsperada)
        exit 0
    }

    $diretorioBackup = Join-Path $raiz 'backups\banco'
    $diretorioBackupCompleto = [IO.Path]::GetFullPath($diretorioBackup)
    if (-not $diretorioBackupCompleto.StartsWith($raiz + '\', [StringComparison]::OrdinalIgnoreCase)) {
        throw 'O diretório de backup saiu da raiz do Kaptei.'
    }
    if (-not (Test-Path -LiteralPath $diretorioBackupCompleto)) {
        New-Item -ItemType Directory -Path $diretorioBackupCompleto | Out-Null
    }
    $pastaBackup = Get-Item -LiteralPath $diretorioBackupCompleto
    if ($pastaBackup.Attributes -band [IO.FileAttributes]::ReparsePoint) {
        throw 'O diretório de backup não pode ser um link ou junction.'
    }
    Set-DiretorioPrivado $diretorioBackupCompleto

    $instante = Get-Date -Format 'yyyyMMdd-HHmmssfff'
    $sufixoUnico = [Guid]::NewGuid().ToString('N').Substring(0, 8)
    $arquivoBackup = Join-Path $diretorioBackupCompleto "kaptei-v$versaoAtual-$instante-$sufixoUnico.dump"
    if (Test-Path -LiteralPath $arquivoBackup) {
        throw 'O destino único do backup já existe; a operação foi interrompida.'
    }

    Write-Host "Criando backup anterior às migrations $($versaoAtual + 1)–$versaoEsperada..."
    & $pgDump --host $conexao.Host --port $conexao.Porta --username $conexao.Usuario `
        --dbname $conexao.Banco --format custom --no-owner --no-privileges --file $arquivoBackup
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $arquivoBackup) -or
        (Get-Item -LiteralPath $arquivoBackup).Length -eq 0) {
        Remove-BackupParcialSeguro $arquivoBackup $diretorioBackupCompleto
        throw 'O backup PostgreSQL falhou ou resultou em arquivo vazio.'
    }

    & $pgRestore --list $arquivoBackup *> $null
    if ($LASTEXITCODE -ne 0) {
        Remove-BackupParcialSeguro $arquivoBackup $diretorioBackupCompleto
        throw 'O backup inválido foi removido porque o catálogo não pôde ser validado pelo pg_restore.'
    }
    $hash = (Get-FileHash -LiteralPath $arquivoBackup -Algorithm SHA256).Hash
    $manifesto = [ordered]@{
        arquivo       = [IO.Path]::GetFileName($arquivoBackup)
        criado_em     = (Get-Date).ToUniversalTime().ToString('o')
        schema_origem = $versaoAtual
        schema_destino = $versaoEsperada
        sha256        = $hash
    } | ConvertTo-Json
    [IO.File]::WriteAllText("$arquivoBackup.json", $manifesto, [Text.UTF8Encoding]::new($false))
    Write-Host "Backup validado: $arquivoBackup" -ForegroundColor Green
    if ($SomenteBackup) {
        Write-Host 'Modo somente backup concluído; nenhuma migration foi executada.' -ForegroundColor Green
        exit 0
    }

    Push-Location $backend
    try {
        & $go run ./cmd/migrador
        if ($LASTEXITCODE -ne 0) {
            throw 'O migrador retornou erro. A API não será iniciada.'
        }
    } finally {
        Pop-Location
    }

    $estadoFinal = Invoke-ConsultaEscalar $consultaEstadoSchema
    if ($estadoFinal -ne ([string]$versaoEsperada + '|false')) {
        throw ('Estado inesperado após migrations: {0}.' -f $estadoFinal)
    }
    $consultaConstraints = @"
SELECT COUNT(*)
FROM pg_constraint
WHERE NOT convalidated
  AND conname IN ('usuarios_conta_obrigatoria_ck','usuarios_papel_valido_ck','usuarios_status_valido_ck','contas_tipo_valido_ck','contas_status_plano_valido_ck','contas_lead_estrategia_valida_ck','imoveis_tipo_valido_ck','imoveis_finalidade_valida_ck','imoveis_status_valido_ck','imoveis_valores_positivos_ck','clientes_status_funil_valido_ck','clientes_temperatura_valida_ck','leads_status_valido_ck','agendamentos_status_valido_ck','agendamentos_tipo_valido_ck');
"@
    $constraintsPendentes = [int](Invoke-ConsultaEscalar $consultaConstraints)
    if ($constraintsPendentes -ne 0) {
        throw ('Ainda existem {0} constraints de domínio não validadas.' -f $constraintsPendentes)
    }
    Write-Host ('Migrations concluídas e verificadas na versão {0}.' -f $versaoEsperada) -ForegroundColor Green
} finally {
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
    Remove-Item Env:PGSSLMODE -ErrorAction SilentlyContinue
}
