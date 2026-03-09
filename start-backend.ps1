# Script de conveniencia para subir a API Go em segundo plano no Windows.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Resolve caminhos absolutos a partir da raiz do repositrio.
$repoRoot = Split-Path -Parent $PSCommandPath
$backendDir = Join-Path $repoRoot "backend"
$binaryDir = Join-Path $backendDir "bin"
$binaryPath = Join-Path $binaryDir "api.exe"
$supervisorScriptPath = Join-Path $repoRoot "backend-supervisor.ps1"
$healthURL = "http://localhost:8080/health"
$logDir = Join-Path $repoRoot ".logs"
$supervisorPidFile = Join-Path $logDir "backend-supervisor.pid"

function Get-SupervisorProcess {
  if (-not (Test-Path $supervisorPidFile)) {
    return $null
  }

  $rawPid = Get-Content -Path $supervisorPidFile -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($null -eq $rawPid) {
    return $null
  }

  $normalizedPid = $rawPid.ToString().Trim()
  if ($normalizedPid -eq "") {
    return $null
  }

  $pidValue = 0
  if (-not [int]::TryParse($normalizedPid, [ref]$pidValue)) {
    return $null
  }

  return Get-Process -Id $pidValue -ErrorAction SilentlyContinue
}

function Test-BackendHealth {
  # Verifica rapidamente se a API ja esta respondendo.
  try {
    $response = Invoke-WebRequest -UseBasicParsing -Uri $healthURL -TimeoutSec 3
    return $response.Content.Trim().ToLower() -eq "ok"
  } catch {
    return $false
  }
}

function Restart-ListeningBackendProcess {
  # Forca recarga do binario recem-compilado quando ja existe API ativa na porta 8080.
  $connections = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue
  if (-not $connections) {
    return
  }

  $processIds = $connections | Select-Object -ExpandProperty OwningProcess -Unique
  foreach ($processId in $processIds) {
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    Write-Host ("Reiniciando API em execucao para aplicar o novo build (PID encerrado: {0})" -f $processId)
  }
}

if (-not (Test-Path $binaryDir)) {
  # Garante a pasta de binarios antes da compilacao.
  New-Item -ItemType Directory -Path $binaryDir | Out-Null
}

Write-Host "Compilando backend..."
Push-Location $backendDir
try {
  go build -o $binaryPath .\cmd\api
} finally {
  Pop-Location
}

if (-not (Test-Path $supervisorScriptPath)) {
  Write-Error "Arquivo nao encontrado: $supervisorScriptPath"
  exit 1
}

if (-not (Test-Path $logDir)) {
  New-Item -Path $logDir -ItemType Directory | Out-Null
}

$supervisorProcess = Get-SupervisorProcess
if ($null -eq $supervisorProcess) {
  Write-Host "Iniciando supervisor do backend..."
  $supervisorProcess = Start-Process `
    -FilePath "powershell.exe" `
    -ArgumentList @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", $supervisorScriptPath, "-RepoRoot", $repoRoot) `
    -WorkingDirectory $repoRoot `
    -WindowStyle Hidden `
    -PassThru
  Set-Content -Path $supervisorPidFile -Value $supervisorProcess.Id -Encoding ascii
} else {
  Write-Host ("Supervisor do backend ja esta ativo (PID: {0})" -f $supervisorProcess.Id)
}

# Garante aplicacao imediata do binario atualizado quando o backend ja estava de pe.
Restart-ListeningBackendProcess

# Aguarda a API ficar pronta antes de retornar sucesso.
for ($attempt = 0; $attempt -lt 20; $attempt++) {
  Start-Sleep -Seconds 1
  if (Test-BackendHealth) {
    $connection = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue |
      Select-Object -First 1
    if ($connection) {
      Write-Host ("API ativa em http://localhost:8080 (PID: {0})" -f $connection.OwningProcess)
    } else {
      Write-Host "API ativa em http://localhost:8080"
    }
    exit 0
  }
}

Write-Error "A API nao respondeu no tempo esperado. Verifique .logs\\backend-supervisor.log e .logs\\backend.err.log."
exit 1
