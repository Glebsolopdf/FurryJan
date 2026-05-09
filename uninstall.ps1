$ErrorActionPreference = "Stop"

$appData = $env:APPDATA
$localAppData = $env:LOCALAPPDATA

if ([string]::IsNullOrWhiteSpace($appData) -or [string]::IsNullOrWhiteSpace($localAppData)) {
    throw "APPDATA or LOCALAPPDATA is not set."
}

$configDir = Join-Path $appData "furryjan"
$installDir = Join-Path $localAppData "Programs\Furryjan"

Write-Host "Removing config: $configDir"
if (Test-Path $configDir) {
    Remove-Item -Path $configDir -Recurse -Force
}

Write-Host "Removing installation: $installDir"
if (Test-Path $installDir) {
    Remove-Item -Path $installDir -Recurse -Force
}

Write-Host "Uninstall complete."
