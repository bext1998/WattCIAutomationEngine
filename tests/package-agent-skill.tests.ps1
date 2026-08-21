# Run with: powershell -NoProfile -ExecutionPolicy Bypass -File .\tests\package-agent-skill.tests.ps1
$ErrorActionPreference = 'Stop'

$repositoryRoot = Split-Path -Path $PSScriptRoot -Parent
$packageScript = Join-Path $repositoryRoot 'scripts\package-agent-skill.ps1'
$expectedFiles = @(
    'watt/SKILL.md',
    'watt/references/authoring.md',
    'watt/references/execution.md'
)

$tempDirectory = Join-Path ([IO.Path]::GetTempPath()) ('watt-agent-skill-test-' + [Guid]::NewGuid().ToString('N'))
try {
    $outputDirectory = Join-Path $tempDirectory 'dist'
    & $packageScript -SourceDirectory (Join-Path $repositoryRoot 'skills\watt') -OutputDirectory $outputDirectory | Out-Null

    $zipPath = Join-Path $outputDirectory 'watt-agent-skill.zip'
    $checksumPath = Join-Path $outputDirectory 'watt-agent-skill.zip.sha256'
    if (-not (Test-Path -LiteralPath $zipPath -PathType Leaf)) {
        throw 'Agent Skill ZIP was not created.'
    }
    if (-not (Test-Path -LiteralPath $checksumPath -PathType Leaf)) {
        throw 'Agent Skill checksum was not created.'
    }

    $extractDirectory = Join-Path $tempDirectory 'extract'
    Expand-Archive -LiteralPath $zipPath -DestinationPath $extractDirectory -Force
    $actualFiles = @(Get-ChildItem -LiteralPath $extractDirectory -File -Recurse | ForEach-Object {
        $_.FullName.Substring($extractDirectory.Length).TrimStart([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar).Replace('\', '/')
    } | Sort-Object)
    if (($actualFiles -join "`n") -ne (($expectedFiles | Sort-Object) -join "`n")) {
        throw "Unexpected ZIP file list: $($actualFiles -join ', ')"
    }

    foreach ($archivePath in $expectedFiles) {
        $relativePath = $archivePath.Substring('watt/'.Length).Replace('/', [IO.Path]::DirectorySeparatorChar)
        $sourceBytes = [IO.File]::ReadAllBytes((Join-Path $repositoryRoot "skills\watt\$relativePath"))
        $archiveBytes = [IO.File]::ReadAllBytes((Join-Path $extractDirectory $archivePath.Replace('/', [IO.Path]::DirectorySeparatorChar)))
        if ($sourceBytes.Length -ne $archiveBytes.Length) {
            throw "ZIP content differs for $archivePath."
        }
        for ($index = 0; $index -lt $sourceBytes.Length; $index++) {
            if ($sourceBytes[$index] -ne $archiveBytes[$index]) {
                throw "ZIP content differs for $archivePath."
            }
        }
    }

    $checksum = [IO.File]::ReadAllText($checksumPath)
    $expectedHash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($checksum -ne "$expectedHash  watt-agent-skill.zip`n") {
        throw 'Agent Skill checksum does not match the ZIP.'
    }

    Write-Host 'PASS agent skill package structure, content, and checksum'
}
finally {
    Remove-Item -LiteralPath $tempDirectory -Force -Recurse -ErrorAction SilentlyContinue
}
