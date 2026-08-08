[CmdletBinding()]
param(
    [string]$OutputPath = (Join-Path $PSScriptRoot '..\dist\watt.exe'),
    [string]$Version = 'dev'
)

$resolvedOutputPath = [System.IO.Path]::GetFullPath($OutputPath)
$outputDirectory = Split-Path -Path $resolvedOutputPath -Parent
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'

Push-Location -LiteralPath (Join-Path $PSScriptRoot '..')
try {
    go build -trimpath -ldflags "-X main.version=$Version" -o $resolvedOutputPath ./cmd/watt
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}
