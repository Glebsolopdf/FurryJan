param(
    [switch]$SkipPathUpdate
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$srcDir = Join-Path $projectRoot "src"
$binaryName = "furryjan.exe"
$buildOutput = Join-Path $srcDir $binaryName

$localAppData = $env:LOCALAPPDATA
if ([string]::IsNullOrWhiteSpace($localAppData)) {
    throw "LOCALAPPDATA is not set."
}

$installDir = Join-Path $localAppData "Programs\Furryjan"
$targetBinary = Join-Path $installDir $binaryName

Write-Host "Building Furryjan for Windows..."
Push-Location $srcDir
try {
    go mod tidy
    go build -o $binaryName ./cmd/main.go
} finally {
    Pop-Location
}

Write-Host "Installing to $targetBinary..."
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Copy-Item -Path $buildOutput -Destination $targetBinary -Force

if (-not $SkipPathUpdate) {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ([string]::IsNullOrWhiteSpace($currentPath)) {
        $currentPath = ""
    }

    $paths = $currentPath -split ";" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    if ($paths -notcontains $installDir) {
        $newPath = ($paths + $installDir) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added install directory to user PATH. Restart terminal to use command 'furryjan'."
    }
}

Write-Host "Installation complete."
Write-Host "Run: $targetBinary"
