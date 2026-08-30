param(
    [string]$CaminhoEnv = (Join-Path $PSScriptRoot '..\backend\.env')
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Import-Module (Join-Path $PSScriptRoot 'Kaptei.Ambiente.psm1') -Force

$caminhoResolvido = (Resolve-Path -LiteralPath $CaminhoEnv -ErrorAction Stop).Path
$arquivo = Get-Item -LiteralPath $caminhoResolvido -ErrorAction Stop
if ($arquivo.PSIsContainer -or ($arquivo.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
    throw 'backend\.env deve ser um arquivo regular, não um link ou diretório.'
}

$configuracao = Get-ConfiguracaoEnv $caminhoResolvido
$chaveExistente = if ($configuracao.ContainsKey('CONFIG_ENCRYPTION_KEY')) {
    $configuracao['CONFIG_ENCRYPTION_KEY'].Trim()
} else {
    ''
}

if (-not [string]::IsNullOrWhiteSpace($chaveExistente)) {
    try {
        $bytes = [Convert]::FromBase64String($chaveExistente)
    } catch {
        throw 'CONFIG_ENCRYPTION_KEY existe, mas não é Base64 válido. O valor não foi alterado.'
    }
    if ($bytes.Length -ne 32) {
        throw 'CONFIG_ENCRYPTION_KEY existe, mas não possui 32 bytes. O valor não foi alterado.'
    }
    Write-Host 'Chave de criptografia local já configurada e válida.' -ForegroundColor Green
    exit 0
}

if (-not (Test-AmbienteLocal $configuracao)) {
    throw 'A geração automática da chave é permitida somente quando ENV declara desenvolvimento local.'
}

$bytesNovos = [byte[]]::new(32)
$gerador = [Security.Cryptography.RandomNumberGenerator]::Create()
try {
    $gerador.GetBytes($bytesNovos)
} finally {
    $gerador.Dispose()
}
$novaChave = [Convert]::ToBase64String($bytesNovos)
$conteudo = [IO.File]::ReadAllText($caminhoResolvido)
$linhaNova = "CONFIG_ENCRYPTION_KEY=$novaChave"

if ([regex]::IsMatch($conteudo, '(?m)^\s*CONFIG_ENCRYPTION_KEY\s*=.*$')) {
    $conteudoAtualizado = [regex]::Replace(
        $conteudo,
        '(?m)^\s*CONFIG_ENCRYPTION_KEY\s*=.*$',
        [System.Text.RegularExpressions.MatchEvaluator]{ param($_) $linhaNova },
        1
    )
} else {
    $separador = if ($conteudo.Length -eq 0 -or $conteudo.EndsWith("`n")) { '' } else { [Environment]::NewLine }
    $conteudoAtualizado = $conteudo + $separador + $linhaNova + [Environment]::NewLine
}

[IO.File]::WriteAllText($caminhoResolvido, $conteudoAtualizado, [Text.UTF8Encoding]::new($false))
$validacao = Get-ConfiguracaoEnv $caminhoResolvido
if (-not $validacao.ContainsKey('CONFIG_ENCRYPTION_KEY') -or
    ([Convert]::FromBase64String($validacao['CONFIG_ENCRYPTION_KEY'])).Length -ne 32) {
    throw 'A chave foi gerada, mas a validação posterior do arquivo falhou.'
}

Write-Host 'Chave de criptografia local gerada e salva no arquivo ignorado pelo Git.' -ForegroundColor Green
