# Supervisor do backend: mantem a API ativa e reinicia automaticamente em caso de queda.
param(
  [Parameter(Mandatory = $true)]
  [string]$RepoRoot
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$backendDir = Join-Path $RepoRoot "backend"
$binaryPath = Join-Path $backendDir "bin\\api.exe"
$healthURL = "http://localhost:8080/health"
$logDir = Join-Path $RepoRoot ".logs"
$supervisorLog = Join-Path $logDir "backend-supervisor.log"
$backendOutLog = Join-Path $logDir "backend.out.log"
$backendErrLog = Join-Path $logDir "backend.err.log"

function Write-SupervisorLog {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  if (-not (Test-Path $logDir)) {
    New-Item -Path $logDir -ItemType Directory | Out-Null
  }

  $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
  Add-Content -Path $supervisorLog -Value "[$timestamp] $Message"
}

function Test-BackendHealth {
  try {
    $response = Invoke-WebRequest -UseBasicParsing -Uri $healthURL -TimeoutSec 3
    return $response.Content.Trim().ToLower() -eq "ok"
  } catch {
    return $false
  }
}

function Get-ListeningBackendProcessIds {
  $connections = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue
  if (-not $connections) {
    return @()
  }
  return $connections | Select-Object -ExpandProperty OwningProcess -Unique
}

function Stop-StaleBackendProcesses {
  # Se existe listener na porta mas healthcheck falhou, o processo pode estar travado.
  $processIds = Get-ListeningBackendProcessIds
  foreach ($processId in $processIds) {
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    Write-SupervisorLog "Processo antigo na porta 8080 encerrado (PID: $processId)."
  }
}

function Start-BackendProcess {
  if (-not (Test-Path $binaryPath)) {
    Write-SupervisorLog "Binario do backend nao encontrado em $binaryPath"
    return
  }

  if (-not (Test-Path $logDir)) {
    New-Item -Path $logDir -ItemType Directory | Out-Null
  }

  # Inicia a API em background e registra saida para facilitar diagnostico.
  $process = Start-Process `
    -FilePath $binaryPath `
    -WorkingDirectory $backendDir `
    -RedirectStandardOutput $backendOutLog `
    -RedirectStandardError $backendErrLog `
    -PassThru

  Write-SupervisorLog "Backend iniciado (PID: $($process.Id))."
}

Write-SupervisorLog "Supervisor iniciado."

while ($true) {
  try {
    if (-not (Test-BackendHealth)) {
      Stop-StaleBackendProcesses
      Start-BackendProcess
    }
  } catch {
    Write-SupervisorLog ("Falha no ciclo de supervisao: {0}" -f $_.Exception.Message)
  }

  Start-Sleep -Seconds 10
}
