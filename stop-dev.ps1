# Script de conveniencia para parar backend e frontend locais no Windows.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Resolve caminhos absolutos dos scripts auxiliares.
$repoRoot = Split-Path -Parent $PSCommandPath
$backendStopScript = Join-Path $repoRoot "stop-backend.ps1"
$frontendPort = 5173
$stoppedFrontend = $false

# Encerra o processo que estiver ouvindo a porta do frontend.
$frontendConnections = Get-NetTCPConnection -State Listen -LocalPort $frontendPort -ErrorAction SilentlyContinue
if ($frontendConnections) {
  $frontendProcessIds = $frontendConnections | Select-Object -ExpandProperty OwningProcess -Unique
  foreach ($frontendProcessId in $frontendProcessIds) {
    Stop-Process -Id $frontendProcessId -Force -ErrorAction SilentlyContinue
    $stoppedFrontend = $true
  }
}

if ($stoppedFrontend) {
  Write-Host "Frontend parado com sucesso."
} else {
  Write-Host "Nenhum frontend ativo na porta $frontendPort."
}

if (Test-Path $backendStopScript) {
  # Reaproveita o script existente para desligar a API.
  powershell -ExecutionPolicy Bypass -File $backendStopScript
} else {
  Write-Host "Arquivo nao encontrado: $backendStopScript"
}
