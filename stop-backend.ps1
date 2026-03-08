# Script de conveniencia para parar a API Go em execucao no Windows.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$stoppedAny = $false
$repoRoot = Split-Path -Parent $PSCommandPath
$logDir = Join-Path $repoRoot ".logs"
$supervisorPidFile = Join-Path $logDir "backend-supervisor.pid"

if (Test-Path $supervisorPidFile) {
  $rawPid = Get-Content -Path $supervisorPidFile -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($null -ne $rawPid) {
    $pidValue = 0
    $normalizedPid = $rawPid.ToString().Trim()
    if ([int]::TryParse($normalizedPid, [ref]$pidValue)) {
      # Para o supervisor primeiro para evitar reinicio automatico durante o desligamento manual.
      $supervisorProcess = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($supervisorProcess) {
        Stop-Process -Id $pidValue -Force -ErrorAction SilentlyContinue
        $stoppedAny = $true
      }
    }
  }

  Remove-Item -Path $supervisorPidFile -Force -ErrorAction SilentlyContinue
}

# Para processos que estejam ouvindo na porta da API.
$connections = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue
if ($connections) {
  $processIds = $connections | Select-Object -ExpandProperty OwningProcess -Unique
  foreach ($processId in $processIds) {
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    $stoppedAny = $true
  }
}

# Tambem tenta parar processos chamados "api" que possam nao estar na porta esperada.
$apiProcesses = Get-Process -Name "api" -ErrorAction SilentlyContinue
if ($apiProcesses) {
  foreach ($apiProcess in $apiProcesses) {
    Stop-Process -Id $apiProcess.Id -Force -ErrorAction SilentlyContinue
    $stoppedAny = $true
  }
}

if ($stoppedAny) {
  Write-Host "Backend parado com sucesso."
} else {
  Write-Host "Nenhum processo de backend ativo foi encontrado."
}
