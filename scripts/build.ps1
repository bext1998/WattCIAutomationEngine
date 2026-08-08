[CmdletBinding()]
param(
    [string]$OutputPath = (Join-Path $PSScriptRoot '..\dist\watt.exe'),
    [string]$Version = 'dev'
)

$resolvedOutputPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputPath)
$outputDirectory = Split-Path -Path $resolvedOutputPath -Parent
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

$originalGoOS = [System.Environment]::GetEnvironmentVariable('GOOS', 'Process')
$originalGoArch = [System.Environment]::GetEnvironmentVariable('GOARCH', 'Process')
$originalCgoEnabled = [System.Environment]::GetEnvironmentVariable('CGO_ENABLED', 'Process')

try {
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
}
finally {
    [System.Environment]::SetEnvironmentVariable('GOOS', $originalGoOS, 'Process')
    [System.Environment]::SetEnvironmentVariable('GOARCH', $originalGoArch, 'Process')
    [System.Environment]::SetEnvironmentVariable('CGO_ENABLED', $originalCgoEnabled, 'Process')
}
