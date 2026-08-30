# ==============================================================================
# Script de Deploy Completo e Automatizado - Kaptei SaaS
# Executado a partir do ambiente local para publicação e criação de serviços Windows
# ==============================================================================

# Configurações do Deploy
$ServerIP = "192.168.100.120"
$DestPath = "\\$ServerIP\msd\MSDev-kaptei"
$ApiPort = "8012"
$FrontendPort = "8013"
$ProdDatabaseName = "msdev_kaptei_prod"

# Cores do terminal
$Yellow = "Yellow"
$Green = "Green"
$Cyan = "Cyan"
$Red = "Red"

Write-Host "======================================================================" -ForegroundColor $Cyan
Write-Host "INICIANDO DEPLOY COMPLETO - KAPTEI SAAS" -ForegroundColor $Cyan
Write-Host "======================================================================" -ForegroundColor $Cyan
Write-Host "Destino: $DestPath" -ForegroundColor $Yellow
Write-Host "Porta API: $ApiPort | Porta Frontend: $FrontendPort" -ForegroundColor $Yellow
Write-Host "----------------------------------------------------------------------" -ForegroundColor $Cyan

# 1. Validações Iniciais
if (-not (Test-Path $DestPath)) {
    Write-Host "Erro: O caminho de destino '$DestPath' não está acessível." -ForegroundColor $Red
    Write-Host "Verifique se a rede está ativa e se as permissões de acesso ao compartilhamento estão corretas." -ForegroundColor $Red
    exit 1
}

# 2. Configurações de Ambiente (Parâmetros no Frontend)
Write-Host "[1/6]# Configura as variáveis de ambiente do Frontend para o Deploy" -ForegroundColor $Yellow
$ViteEnvPath = ".\frontend\.env.production"
$ViteEnvContent = "VITE_API_URL=https://api-kaptei.msdevsolutions.com.br/api"
# Usa [System.IO.File]::WriteAllText para evitar o BOM (Byte Order Mark) que quebra o Vite
[System.IO.File]::WriteAllText((Resolve-Path .\frontend).Path + "\.env.production", $ViteEnvContent)
Write-Host "Arquivo .env.production gerado com sucesso para a API: https://api-kaptei.msdevsolutions.com.br/api" -ForegroundColor $Green

# 3. Compilação do Backend
Write-Host "[2/6] Compilando o Backend em Go (api.exe)..." -ForegroundColor $Yellow
Push-Location .\backend
# Compila o executável principal do backend
go build -o kaptei.exe cmd/api/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro na compilação do Backend em Go!" -ForegroundColor $Red
    Pop-Location
    exit 1
}
Write-Host "Backend compilado com sucesso!" -ForegroundColor $Green

# Compila o servidor estático para o frontend
Write-Host "Compilando o Servidor Estático do Frontend em Go (servidor_web.exe)..." -ForegroundColor $Yellow
go build -o servidor_web.exe ./cmd/static-server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro na compilação do Servidor Estático!" -ForegroundColor $Red
    Pop-Location
    exit 1
}
Write-Host "Servidor Estático do Frontend compilado!" -ForegroundColor $Green
Pop-Location

# 4. Geração da Build do Frontend
Write-Host "[3/6] Gerando a Build do Frontend React..." -ForegroundColor $Yellow
Push-Location .\frontend
if (-not (Test-Path ".\node_modules")) {
    Write-Host "node_modules ausente. Instalando dependências..." -ForegroundColor $Yellow
    npm install
}
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro ao gerar build do Frontend React!" -ForegroundColor $Red
    Pop-Location
    exit 1
}
Write-Host "Build do Frontend React gerada com sucesso!" -ForegroundColor $Green
Pop-Location

# 5. Obtendo NSSM do servidor remoto (caso não exista na pasta local)
$NssmPath = ".\nssm.exe"
$ServerNssmPath = "\\$ServerIP\msd\Tools\nssm\win64\nssm.exe"

if (-not (Test-Path $NssmPath)) {
    Write-Host "[4/6] Copiando NSSM do servidor ($ServerNssmPath)..." -ForegroundColor $Yellow
    if (Test-Path $ServerNssmPath) {
        Copy-Item $ServerNssmPath -Destination $NssmPath -Force
        Write-Host "NSSM copiado com sucesso para a pasta local!" -ForegroundColor $Green
    } else {
        Write-Host "Aviso: NSSM não encontrado em '$ServerNssmPath'." -ForegroundColor $Red
        Write-Host "Tentando baixar NSSM da internet..." -ForegroundColor $Yellow
        try {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile "nssm.zip"
            Expand-Archive -Path "nssm.zip" -DestinationPath "nssm_temp" -Force
            Copy-Item "nssm_temp\nssm-2.24\win64\nssm.exe" -Destination $NssmPath
            Remove-Item "nssm.zip" -Force
            Remove-Item "nssm_temp" -Recurse -Force
            Write-Host "NSSM obtido com sucesso da internet." -ForegroundColor $Green
        } catch {
            Write-Host "Erro: Não foi possível obter o NSSM localmente nem da internet." -ForegroundColor $Red
        }
    }
}

# 6. Copiando arquivos e implantando no Servidor (Deploy)
Write-Host "[5/6] Transferindo arquivos para o servidor remoto ($DestPath)..." -ForegroundColor $Yellow

Write-Host "Parando serviços no servidor remoto (\\$ServerIP) para liberar arquivos..." -ForegroundColor $Cyan
sc.exe \\$ServerIP stop MSDev_Kaptei_Backend *>$null
sc.exe \\$ServerIP stop MSDev_Kaptei_Frontend *>$null
# Tempo para garantir que os processos soltaram os arquivos .exe
Start-Sleep -Seconds 3

# Define pastas de destino
$BackendDest = Join-Path $DestPath "backend"
$FrontendDest = Join-Path $DestPath "frontend"

# Cria pastas se não existirem
New-Item -ItemType Directory -Force -Path $BackendDest | Out-Null
New-Item -ItemType Directory -Force -Path $FrontendDest | Out-Null

# Copiar arquivos do Backend
Write-Host "Copiando binários do Backend..." -ForegroundColor $Cyan
Copy-Item ".\backend\kaptei.exe" -Destination $BackendDest -Force

Write-Host "Copiando arquivos de Migration (db)..." -ForegroundColor $Cyan
Copy-Item ".\backend\db" -Destination $BackendDest -Recurse -Force
Write-Host "Atualizando .env no servidor com as configurações de Produção..." -ForegroundColor $Cyan
Copy-Item ".\backend\.env" -Destination $BackendDest -Force
$envContent = Get-Content (Join-Path $BackendDest ".env")
$envContent = $envContent -replace "PORT=8080", "PORT=$ApiPort"
$envContent = $envContent -replace "ENV=development", "ENV=production"
$envContent = $envContent -replace "ENVIRONMENT=development", "ENVIRONMENT=production"
# Atualizar as variáveis de banco de dados
$envContent = $envContent -replace "DB_DATABASE=.*", "DB_DATABASE=$ProdDatabaseName"
# Usar WriteAllLines para escrever UTF-8 sem BOM
[System.IO.File]::WriteAllLines((Join-Path $BackendDest ".env"), $envContent)

# Copiar arquivos do Frontend
Write-Host "Copiando build do Frontend..." -ForegroundColor $Cyan
Copy-Item ".\backend\servidor_web.exe" -Destination $FrontendDest -Force
# Remove dist antiga se houver para evitar lixo
Remove-Item (Join-Path $FrontendDest "dist") -Recurse -Force -ErrorAction SilentlyContinue
Copy-Item ".\frontend\dist" -Destination $FrontendDest -Recurse -Force

# Copiar o NSSM para o servidor
Copy-Item $NssmPath -Destination $DestPath -Force

Write-Host "Arquivos transferidos com sucesso para o servidor!" -ForegroundColor $Green

# 7. Criação/Reinicialização dos Serviços no Servidor
Write-Host "[6/6] Gerando scripts de configuração de serviços no servidor..." -ForegroundColor $Yellow

$ServicosScript = @"
# ==============================================================================
# Script Auxiliar de Instalação de Serviços Windows - Kaptei (Executado no Servidor)
# Executar este script como ADMINISTRADOR no Servidor.
# ==============================================================================

Set-Location -Path "`$PSScriptRoot"

Write-Host "Parando serviços existentes..." -ForegroundColor Cyan
Stop-Service -Name "MSDev_Kaptei_Backend" -ErrorAction SilentlyContinue
Stop-Service -Name "MSDev_Kaptei_Frontend" -ErrorAction SilentlyContinue

# Tempo de respiro para destravar arquivos no disco
Start-Sleep -Seconds 2

# Instalação / Configuração do Backend
Write-Host "Configurando o serviço da API do Backend (Porta $ApiPort)..." -ForegroundColor Yellow
if (Get-Service -Name "MSDev_Kaptei_Backend" -ErrorAction SilentlyContinue) {
    Write-Host "Serviço MSDev_Kaptei_Backend já existe. Atualizando configurações..." -ForegroundColor Cyan
} else {
    .\nssm.exe install MSDev_Kaptei_Backend "`$PSScriptRoot\backend\kaptei.exe"
}
.\nssm.exe set MSDev_Kaptei_Backend AppDirectory "`$PSScriptRoot\backend"
.\nssm.exe set MSDev_Kaptei_Backend DisplayName "MSDev Kaptei API (Porta $ApiPort)"
.\nssm.exe set MSDev_Kaptei_Backend Description "API de Backend do CRM Kaptei rodando na porta $ApiPort"
.\nssm.exe set MSDev_Kaptei_Backend Start SERVICE_AUTO_START
.\nssm.exe set MSDev_Kaptei_Backend AppStdout "`$PSScriptRoot\backend\stdout.log"
.\nssm.exe set MSDev_Kaptei_Backend AppStderr "`$PSScriptRoot\backend\stderr.log"

# Instalação / Configuração do Frontend Web
Write-Host "Configurando o serviço do Frontend Web (Porta $FrontendPort)..." -ForegroundColor Yellow
if (Get-Service -Name "MSDev_Kaptei_Frontend" -ErrorAction SilentlyContinue) {
    Write-Host "Serviço MSDev_Kaptei_Frontend já existe. Atualizando configurações..." -ForegroundColor Cyan
} else {
    .\nssm.exe install MSDev_Kaptei_Frontend "`$PSScriptRoot\frontend\servidor_web.exe" "-port $FrontendPort -dir `$PSScriptRoot\frontend\dist"
}
.\nssm.exe set MSDev_Kaptei_Frontend AppDirectory "`$PSScriptRoot\frontend"
.\nssm.exe set MSDev_Kaptei_Frontend DisplayName "MSDev Kaptei Web (Porta $FrontendPort)"
.\nssm.exe set MSDev_Kaptei_Frontend Description "Servidor Web SPA do CRM Kaptei rodando na porta $FrontendPort"
.\nssm.exe set MSDev_Kaptei_Frontend Start SERVICE_AUTO_START
.\nssm.exe set MSDev_Kaptei_Frontend AppStdout "`$PSScriptRoot\frontend\stdout.log"
.\nssm.exe set MSDev_Kaptei_Frontend AppStderr "`$PSScriptRoot\frontend\stderr.log"

# Iniciar serviços
Write-Host "Iniciando os serviços do Windows..." -ForegroundColor Green
Start-Service -Name "MSDev_Kaptei_Backend"
Start-Service -Name "MSDev_Kaptei_Frontend"

Write-Host "Serviços do Windows MSDev_Kaptei configurados e iniciados com sucesso!" -ForegroundColor Green
Get-Service -Name "MSDev_Kaptei_*"
"@

# Grava o instalador de serviços na raiz da pasta do servidor
$ServicosScript | Out-File -FilePath (Join-Path $DestPath "instalar_servicos.ps1") -Encoding utf8 -Force
Write-Host "Script auxiliar de serviços 'instalar_servicos.ps1' gerado com sucesso em '$DestPath'." -ForegroundColor $Green

# Heurística para detectar se o script atual está rodando direto no servidor de destino com privilégios de Admin
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$hostname = [System.Net.Dns]::GetHostName()
$ips = [System.Net.Dns]::GetHostAddresses($hostname) | ForEach-Object { $_.IPAddressToString }
$isServerLocal = ($ips -contains $ServerIP) -or (Test-Path "D:\msd\MSDev-kaptei")

if ($isServerLocal -and $isAdmin) {
    Write-Host "Servidor de destino detectado localmente. Instalando serviços agora..." -ForegroundColor $Yellow
    Push-Location $DestPath
    .\instalar_servicos.ps1
    Pop-Location
} else {
    Write-Host ""
    Write-Host "ATENÇÃO - INSTRUÇÃO ADICIONAL PARA INSTALAÇÃO DOS SERVIÇOS:" -ForegroundColor $Yellow
    Write-Host "1. A build e transferência dos arquivos foram finalizadas na pasta de destino." -ForegroundColor $Cyan
    Write-Host "2. Como você já está logado no servidor ($ServerIP), abra um terminal do PowerShell" -ForegroundColor $Cyan
    Write-Host "   como ADMINISTRADOR no Servidor e execute o seguinte comando para ativar/reiniciar os serviços:" -ForegroundColor $Cyan
    Write-Host "   powershell.exe -ExecutionPolicy Bypass -File \\$ServerIP\msd\MSDev-kaptei\instalar_servicos.ps1" -ForegroundColor $Yellow
}

Write-Host "Iniciando os serviços remotamente via RPC..." -ForegroundColor $Yellow
sc.exe \\$ServerIP start MSDev_Kaptei_Backend *>$null
sc.exe \\$ServerIP start MSDev_Kaptei_Frontend *>$null
Start-Sleep -Seconds 2

Write-Host "======================================================================" -ForegroundColor $Green
Write-Host "DEPLOY EXECUTADO E PRONTO PARA USO!" -ForegroundColor $Green
Write-Host "======================================================================" -ForegroundColor $Green
