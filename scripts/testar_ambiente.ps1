$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Import-Module (Join-Path $PSScriptRoot 'Kaptei.Ambiente.psm1') -Force

$temporario = Join-Path ([IO.Path]::GetTempPath()) ("kaptei-env-teste-{0}.env" -f [guid]::NewGuid())
$diretorioBackupTeste = Join-Path ([IO.Path]::GetTempPath()) ("kaptei-backup-teste-{0}" -f [guid]::NewGuid())
$backupValido = Join-Path $diretorioBackupTeste 'kaptei-v19-20260812-221727123-a1b2c3d4.dump'
$backupInvalido = Join-Path $diretorioBackupTeste 'arquivo-nao-controlado.dump'
try {
    $conteudo = @"
# comentário
ENV=development
DB_HOST=localhost
DB_PORT=5433
DB_USER="usuario@local"
DB_PASSWORD='s:e/n?h#a'
DB_DATABASE=kaptei_teste
"@
    [IO.File]::WriteAllText($temporario, $conteudo, [Text.UTF8Encoding]::new($false))
    $configuracao = Get-ConfiguracaoEnv $temporario
    if (-not (Test-AmbienteLocal $configuracao)) {
        throw 'ENV de desenvolvimento não foi reconhecido.'
    }
    $conexao = Get-ConexaoPostgreSQL $configuracao
    if ($conexao.Host -ne 'localhost' -or $conexao.Porta -ne 5433 -or
        $conexao.Usuario -ne 'usuario@local' -or $conexao.Senha -ne 's:e/n?h#a' -or
        $conexao.Banco -ne 'kaptei_teste' -or $conexao.ModoSSL -ne 'disable') {
        throw 'A configuração PostgreSQL legada não foi interpretada corretamente.'
    }

    $url = [System.Collections.Generic.Dictionary[string, string]]::new([StringComparer]::OrdinalIgnoreCase)
    $url['DATABASE_URL'] = 'postgres://usuario%40local:s%3Ae%2Fn%3Fh%23a@banco:5432/kaptei?sslmode=require'
    $conexaoURL = Get-ConexaoPostgreSQL $url
    if ($conexaoURL.Usuario -ne 'usuario@local' -or $conexaoURL.Senha -ne 's:e/n?h#a' -or
        $conexaoURL.Host -ne 'banco' -or $conexaoURL.Banco -ne 'kaptei' -or
        $conexaoURL.ModoSSL -ne 'require') {
        throw 'DATABASE_URL não foi interpretada corretamente.'
    }

    [void](New-Item -ItemType Directory -Path $diretorioBackupTeste -ErrorAction Stop)
    $proprietarioOriginal = (Get-Acl -LiteralPath $diretorioBackupTeste).Owner
    Set-DiretorioPrivado $diretorioBackupTeste
    $aclPrivada = Get-Acl -LiteralPath $diretorioBackupTeste
    $permitidos = @(
        [Security.Principal.WindowsIdentity]::GetCurrent().User.Value,
        'S-1-5-18',
        'S-1-5-32-544'
    )
    $sids = @($aclPrivada.Access | ForEach-Object {
        $_.IdentityReference.Translate([Security.Principal.SecurityIdentifier]).Value
    } | Sort-Object -Unique)
    if (-not $aclPrivada.AreAccessRulesProtected -or $aclPrivada.Owner -ne $proprietarioOriginal -or
        $sids.Count -ne 3 -or @($permitidos | Where-Object { $_ -notin $sids }).Count -gt 0) {
        throw 'O diretório privado não preservou proprietário e DACL esperados.'
    }

    [IO.File]::WriteAllBytes($backupValido, [byte[]](1, 2, 3))
    Remove-BackupParcialSeguro $backupValido $diretorioBackupTeste
    if (Test-Path -LiteralPath $backupValido) {
        throw 'O backup parcial controlado não foi removido.'
    }

    [IO.File]::WriteAllBytes($backupInvalido, [byte[]](1, 2, 3))
    $nomeInvalidoRecusado = $false
    try {
        Remove-BackupParcialSeguro $backupInvalido $diretorioBackupTeste
    } catch {
        $nomeInvalidoRecusado = $true
    }
    if (-not $nomeInvalidoRecusado -or -not (Test-Path -LiteralPath $backupInvalido)) {
        throw 'A limpeza de backup não respeitou o limite de nome e diretório.'
    }

    Write-Host 'Testes do módulo de ambiente aprovados.' -ForegroundColor Green
} finally {
    if (Test-Path -LiteralPath $temporario) {
        Remove-Item -LiteralPath $temporario -ErrorAction Stop
    }
    if (Test-Path -LiteralPath $backupInvalido -PathType Leaf) {
        Remove-Item -LiteralPath $backupInvalido -Force -ErrorAction Stop
    }
    if (Test-Path -LiteralPath $diretorioBackupTeste -PathType Container) {
        Remove-Item -LiteralPath $diretorioBackupTeste -Force -ErrorAction Stop
    }
}
