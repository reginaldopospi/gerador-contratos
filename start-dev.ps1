# Script de conveniencia para iniciar backend e frontend juntos em segundo plano no Windows.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Resolve os caminhos da raiz do repositorio e dos servicos.
$repoRoot = Split-Path -Parent $PSCommandPath
$backendStartScript = Join-Path $repoRoot "start-backend.ps1"
$frontendDir = Join-Path $repoRoot "frontend"
$frontendPort = 5173
$frontendURL = "http://localhost:$frontendPort"
$logDir = Join-Path $repoRoot ".logs"
$frontendOutLog = Join-Path $logDir "frontend-dev.out.log"
$frontendErrLog = Join-Path $logDir "frontend-dev.err.log"

function Test-PortListening {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port
  )

  # Verifica se existe processo escutando na porta informada.
  $connection = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue
  return $null -ne $connection
}

function Ensure-FrontendDependencies {
  # Instala dependencias somente quando node_modules ainda nao existe.
  $nodeModulesPath = Join-Path $frontendDir "node_modules"
  if (Test-Path $nodeModulesPath) {
    return
  }

  Write-Host "Instalando dependencias do frontend..."
  Push-Location $frontendDir
  try {
    npm.cmd install
  } finally {
    Pop-Location
  }
}

function Start-Frontend {
  if (Test-PortListening -Port $frontendPort) {
    Write-Host "Frontend ja esta ativo em $frontendURL"
    return
  }

  # Garante pasta de logs e limpa arquivos antigos para facilitar diagnostico.
  if (-not (Test-Path $logDir)) {
    New-Item -Path $logDir -ItemType Directory | Out-Null
  }
  if (Test-Path $frontendOutLog) {
    Remove-Item $frontendOutLog -Force
  }
  if (Test-Path $frontendErrLog) {
    Remove-Item $frontendErrLog -Force
  }

  Write-Host "Iniciando frontend..."
  $frontendProcess = Start-Process `
    -FilePath "npm.cmd" `
    -ArgumentList @("run", "dev") `
    -WorkingDirectory $frontendDir `
    -RedirectStandardOutput $frontendOutLog `
    -RedirectStandardError $frontendErrLog `
    -PassThru

  # Aguarda o Vite abrir a porta 5173.
  for ($attempt = 0; $attempt -lt 25; $attempt++) {
    Start-Sleep -Milliseconds 600
    if (Test-PortListening -Port $frontendPort) {
      Write-Host ("Frontend ativo em {0} (PID: {1})" -f $frontendURL, $frontendProcess.Id)
      return
    }
  }

  Write-Error ("Frontend nao iniciou na porta {0}. Consulte: {1} e {2}" -f $frontendPort, $frontendOutLog, $frontendErrLog)
}

if (-not (Test-Path $backendStartScript)) {
  Write-Error "Arquivo nao encontrado: $backendStartScript"
  exit 1
}

if (-not (Test-Path $frontendDir)) {
  Write-Error "Diretorio nao encontrado: $frontendDir"
  exit 1
}

Write-Host "Subindo backend..."
powershell -ExecutionPolicy Bypass -File $backendStartScript
if ($LASTEXITCODE -ne 0) {
  Write-Error "Falha ao iniciar backend."
  exit 1
}

Ensure-FrontendDependencies
Start-Frontend

Write-Host ""
Write-Host "Sistema pronto:"
Write-Host " - Frontend: $frontendURL"
Write-Host " - API: http://localhost:8080/health"
