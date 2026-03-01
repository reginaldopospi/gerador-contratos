# Script de conveniencia para subir a API Go em segundo plano no Windows.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Resolve caminhos absolutos a partir da raiz do repositrio.
$repoRoot = Split-Path -Parent $PSCommandPath
$backendDir = Join-Path $repoRoot "backend"
$binaryDir = Join-Path $backendDir "bin"
$binaryPath = Join-Path $binaryDir "api.exe"
$healthURL = "http://localhost:8080/health"

function Test-BackendHealth {
  # Verifica rapidamente se a API ja esta respondendo.
  try {
    $response = Invoke-WebRequest -UseBasicParsing -Uri $healthURL -TimeoutSec 3
    return $response.Content.Trim().ToLower() -eq "ok"
  } catch {
    return $false
  }
}

if (Test-BackendHealth) {
  Write-Host "API ja esta ativa em http://localhost:8080"
  exit 0
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

# Libera a porta 8080 caso exista processo antigo travado.
$connections = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue
if ($connections) {
  $processIds = $connections | Select-Object -ExpandProperty OwningProcess -Unique
  foreach ($processId in $processIds) {
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
  }
}

Write-Host "Iniciando backend..."
$process = Start-Process -FilePath $binaryPath -WorkingDirectory $backendDir -PassThru

# Aguarda a API ficar pronta antes de retornar sucesso.
for ($attempt = 0; $attempt -lt 10; $attempt++) {
  Start-Sleep -Seconds 1
  if (Test-BackendHealth) {
    Write-Host ("API ativa em http://localhost:8080 (PID: {0})" -f $process.Id)
    exit 0
  }
}

Write-Error "A API nao respondeu no tempo esperado. Verifique logs e configuracao."
exit 1
