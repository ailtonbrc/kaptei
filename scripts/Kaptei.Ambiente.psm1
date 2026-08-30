Set-StrictMode -Version Latest

function Get-ConfiguracaoEnv {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Caminho
    )

    $caminhoResolvido = (Resolve-Path -LiteralPath $Caminho -ErrorAction Stop).Path
    $configuracao = [System.Collections.Generic.Dictionary[string, string]]::new(
        [System.StringComparer]::OrdinalIgnoreCase
    )

    foreach ($linha in [System.IO.File]::ReadAllLines($caminhoResolvido)) {
        $texto = $linha.Trim()
        if ($texto.Length -eq 0 -or $texto.StartsWith('#') -or -not $texto.Contains('=')) {
            continue
        }

        $partes = $texto.Split('=', 2)
        $chave = $partes[0].Trim()
        $valor = $partes[1].Trim()
        if ($chave.StartsWith('export ', [System.StringComparison]::OrdinalIgnoreCase)) {
            $chave = $chave.Substring(7).Trim()
        }
        if ($valor.Length -ge 2 -and (
            ($valor.StartsWith('"') -and $valor.EndsWith('"')) -or
            ($valor.StartsWith("'") -and $valor.EndsWith("'"))
        )) {
            $valor = $valor.Substring(1, $valor.Length - 2)
        }
        if ($chave.Length -gt 0) {
            $configuracao[$chave] = $valor
        }
    }

    return $configuracao
}

function Test-AmbienteLocal {
    param(
        [Parameter(Mandatory = $true)]
        [System.Collections.Generic.IDictionary[string, string]]$Configuracao
    )

    if (-not $Configuracao.ContainsKey('ENV')) {
        return $false
    }
    return $Configuracao['ENV'].Trim().ToLowerInvariant() -in @(
        'development', 'desenvolvimento', 'local', 'dev'
    )
}

function Get-ConexaoPostgreSQL {
    param(
        [Parameter(Mandatory = $true)]
        [System.Collections.Generic.IDictionary[string, string]]$Configuracao
    )

    if ($Configuracao.ContainsKey('DATABASE_URL') -and
        -not [string]::IsNullOrWhiteSpace($Configuracao['DATABASE_URL'])) {
        $endereco = [Uri]$Configuracao['DATABASE_URL']
        if ($endereco.Scheme -notin @('postgres', 'postgresql') -or
            [string]::IsNullOrWhiteSpace($endereco.Host)) {
            throw 'DATABASE_URL deve ser uma URL PostgreSQL válida.'
        }

        $usuario = ''
        $senha = ''
        if (-not [string]::IsNullOrWhiteSpace($endereco.UserInfo)) {
            $credenciais = $endereco.UserInfo.Split(':', 2)
            $usuario = [Uri]::UnescapeDataString($credenciais[0])
            if ($credenciais.Count -gt 1) {
                $senha = [Uri]::UnescapeDataString($credenciais[1])
            }
        }

        $modoSSL = ''
        foreach ($par in $endereco.Query.TrimStart('?').Split('&', [System.StringSplitOptions]::RemoveEmptyEntries)) {
            $partes = $par.Split('=', 2)
            if ([Uri]::UnescapeDataString($partes[0]) -eq 'sslmode' -and $partes.Count -gt 1) {
                $modoSSL = [Uri]::UnescapeDataString($partes[1])
                break
            }
        }

        return [PSCustomObject]@{
            Host      = $endereco.Host
            Porta     = if ($endereco.IsDefaultPort) { 5432 } else { $endereco.Port }
            Usuario   = $usuario
            Senha     = $senha
            Banco     = [Uri]::UnescapeDataString($endereco.AbsolutePath.TrimStart('/'))
            ModoSSL   = $modoSSL
        }
    }

    $obrigatorias = @('DB_HOST', 'DB_USER', 'DB_PASSWORD', 'DB_DATABASE')
    $faltantes = @($obrigatorias | Where-Object {
        -not $Configuracao.ContainsKey($_) -or [string]::IsNullOrWhiteSpace($Configuracao[$_])
    })
    if ($faltantes.Count -gt 0) {
        throw "Configuração PostgreSQL incompleta. Faltam: $($faltantes -join ', ')."
    }

    $porta = 5432
    if ($Configuracao.ContainsKey('DB_PORT') -and -not [string]::IsNullOrWhiteSpace($Configuracao['DB_PORT'])) {
        if (-not [int]::TryParse($Configuracao['DB_PORT'], [ref]$porta) -or $porta -lt 1 -or $porta -gt 65535) {
            throw 'DB_PORT deve ser um inteiro entre 1 e 65535.'
        }
    }

    $modoSSL = if ($Configuracao.ContainsKey('DB_SSLMODE')) { $Configuracao['DB_SSLMODE'].Trim() } else { '' }
    if ([string]::IsNullOrWhiteSpace($modoSSL)) {
        $modoSSL = if (Test-AmbienteLocal $Configuracao) { 'disable' } else { 'require' }
    }

    return [PSCustomObject]@{
        Host      = $Configuracao['DB_HOST'].Trim()
        Porta     = $porta
        Usuario   = $Configuracao['DB_USER'].Trim()
        Senha     = $Configuracao['DB_PASSWORD']
        Banco     = $Configuracao['DB_DATABASE'].Trim()
        ModoSSL   = $modoSSL
    }
}

function Get-FerramentaPostgreSQL {
    param(
        [Parameter(Mandatory = $true)]
        [ValidateSet('psql.exe', 'pg_dump.exe', 'pg_restore.exe')]
        [string]$Nome
    )

    $comando = Get-Command $Nome -ErrorAction SilentlyContinue
    if ($null -ne $comando) {
        return $comando.Source
    }

    $instalacoes = Get-ChildItem -LiteralPath 'C:\Program Files\PostgreSQL' -Directory -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -match '^\d+$' } |
        Sort-Object { [int]$_.Name } -Descending
    foreach ($instalacao in $instalacoes) {
        $candidato = Join-Path $instalacao.FullName "bin\$Nome"
        if (Test-Path -LiteralPath $candidato -PathType Leaf) {
            return $candidato
        }
    }
    throw "$Nome não foi encontrado. Instale as ferramentas cliente do PostgreSQL."
}

function Set-DiretorioPrivado {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Caminho
    )

    $item = Get-Item -LiteralPath (Resolve-Path -LiteralPath $Caminho -ErrorAction Stop).Path -ErrorAction Stop
    if (-not $item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
        throw 'O diretório privado deve ser uma pasta regular, não um link ou arquivo.'
    }

    $usuarioAtual = [Security.Principal.WindowsIdentity]::GetCurrent().User
    $sistema = [Security.Principal.SecurityIdentifier]::new('S-1-5-18')
    $administradores = [Security.Principal.SecurityIdentifier]::new('S-1-5-32-544')
    $heranca = [Security.AccessControl.InheritanceFlags]::ContainerInherit -bor
        [Security.AccessControl.InheritanceFlags]::ObjectInherit
    $proprietarioOriginal = (Get-Acl -LiteralPath $item.FullName -ErrorAction Stop).Owner
    $acl = [Security.AccessControl.DirectorySecurity]::new()
    $acl.SetAccessRuleProtection($true, $false)
    foreach ($identidade in @($usuarioAtual, $sistema, $administradores)) {
        $regra = [Security.AccessControl.FileSystemAccessRule]::new(
            $identidade,
            [Security.AccessControl.FileSystemRights]::FullControl,
            $heranca,
            [Security.AccessControl.PropagationFlags]::None,
            [Security.AccessControl.AccessControlType]::Allow
        )
        [void]$acl.AddAccessRule($regra)
    }

    # Set-Acl tenta reaplicar proprietário e auditoria de ACLs antigas. Persistir
    # somente Access evita exigir SeSecurityPrivilege sem reduzir a proteção.
    $metodoPersistir = [Security.AccessControl.NativeObjectSecurity].GetMethod(
        'Persist',
        [Reflection.BindingFlags]'Instance,NonPublic',
        $null,
        [type[]]@([string], [Security.AccessControl.AccessControlSections]),
        $null
    )
    if ($null -eq $metodoPersistir) {
        throw 'A API do Windows para persistir somente a DACL não está disponível.'
    }
    [void]$metodoPersistir.Invoke(
        $acl,
        @($item.FullName, [Security.AccessControl.AccessControlSections]::Access)
    )

    $permitidos = @($usuarioAtual.Value, $sistema.Value, $administradores.Value)
    $aclAplicada = Get-Acl -LiteralPath $item.FullName
    $regrasNaoPermitidas = @($aclAplicada.Access | Where-Object {
        $sid = $_.IdentityReference.Translate([Security.Principal.SecurityIdentifier]).Value
        $sid -notin $permitidos -or
            $_.AccessControlType -ne [Security.AccessControl.AccessControlType]::Allow -or
            ($_.FileSystemRights -band [Security.AccessControl.FileSystemRights]::FullControl) -ne
                [Security.AccessControl.FileSystemRights]::FullControl
    })
    $sidsAplicados = @($aclAplicada.Access | ForEach-Object {
        $_.IdentityReference.Translate([Security.Principal.SecurityIdentifier]).Value
    } | Sort-Object -Unique)
    $regrasAusentes = @($permitidos | Where-Object { $_ -notin $sidsAplicados })
    if (-not $aclAplicada.AreAccessRulesProtected -or
        $aclAplicada.Owner -ne $proprietarioOriginal -or
        $regrasNaoPermitidas.Count -gt 0 -or $regrasAusentes.Count -gt 0) {
        throw 'Não foi possível restringir integralmente as permissões do diretório privado.'
    }
}

function Remove-BackupParcialSeguro {
    param(
        [Parameter(Mandatory = $true)][string]$Caminho,
        [Parameter(Mandatory = $true)][string]$DiretorioPermitido
    )

    if (-not (Test-Path -LiteralPath $Caminho)) {
        return
    }

    $diretorioResolvido = (Resolve-Path -LiteralPath $DiretorioPermitido -ErrorAction Stop).Path
    $item = Get-Item -LiteralPath $Caminho -Force -ErrorAction Stop
    $nomeValido = $item.Name -match '^kaptei-v\d+-\d{8}-\d{9}-[0-9a-f]{8}\.dump$'
    $diretorioValido = $item.Directory.FullName.Equals(
        $diretorioResolvido,
        [StringComparison]::OrdinalIgnoreCase
    )
    if (-not $nomeValido -or -not $diretorioValido -or $item.PSIsContainer -or
        ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
        throw 'O arquivo parcial não pertence ao destino de backup controlado.'
    }

    Remove-Item -LiteralPath $item.FullName -Force -ErrorAction Stop
    if (Test-Path -LiteralPath $item.FullName) {
        throw 'Não foi possível remover o arquivo parcial de backup.'
    }
}

Export-ModuleMember -Function Get-ConfiguracaoEnv, Test-AmbienteLocal, Get-ConexaoPostgreSQL, Get-FerramentaPostgreSQL, Set-DiretorioPrivado, Remove-BackupParcialSeguro
