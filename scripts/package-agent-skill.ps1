[CmdletBinding()]
param(
    [string]$SourceDirectory = (Join-Path $PSScriptRoot '..\skills\watt'),
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\dist')
)

$ErrorActionPreference = 'Stop'

$expectedFiles = @(
    'SKILL.md',
    'references/authoring.md',
    'references/execution.md'
)

function Get-NormalizedRelativePath {
    param(
        [Parameter(Mandatory = $true)][string]$BasePath,
        [Parameter(Mandatory = $true)][string]$Path
    )

    return $Path.Substring($BasePath.Length).TrimStart([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar).Replace('\', '/')
}

function Assert-ByteForByteEqual {
    param(
        [Parameter(Mandatory = $true)][string]$ExpectedPath,
        [Parameter(Mandatory = $true)][string]$ActualPath
    )

    $expected = [IO.File]::ReadAllBytes($ExpectedPath)
    $actual = [IO.File]::ReadAllBytes($ActualPath)
    if ($expected.Length -ne $actual.Length) {
        throw "Skill package content differs for '$ExpectedPath'."
    }
    for ($index = 0; $index -lt $expected.Length; $index++) {
        if ($expected[$index] -ne $actual[$index]) {
            throw "Skill package content differs for '$ExpectedPath'."
        }
    }
}

$resolvedSourceDirectory = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($SourceDirectory)
$resolvedOutputDirectory = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputDirectory)
if (-not (Test-Path -LiteralPath $resolvedSourceDirectory -PathType Container)) {
    throw "Agent Skill source directory does not exist: $resolvedSourceDirectory"
}

foreach ($relativePath in $expectedFiles) {
    if (-not (Test-Path -LiteralPath (Join-Path $resolvedSourceDirectory $relativePath) -PathType Leaf)) {
        throw "Agent Skill source is missing required file: $relativePath"
    }
}

$stagingDirectory = Join-Path ([IO.Path]::GetTempPath()) ('watt-agent-skill-stage-' + [Guid]::NewGuid().ToString('N'))
$extractionDirectory = Join-Path ([IO.Path]::GetTempPath()) ('watt-agent-skill-extract-' + [Guid]::NewGuid().ToString('N'))
$zipPath = Join-Path $resolvedOutputDirectory 'watt-agent-skill.zip'
$checksumPath = Join-Path $resolvedOutputDirectory 'watt-agent-skill.zip.sha256'

try {
    New-Item -ItemType Directory -Path $resolvedOutputDirectory -Force | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $stagingDirectory 'watt\references') -Force | Out-Null

    foreach ($relativePath in $expectedFiles) {
        $sourcePath = Join-Path $resolvedSourceDirectory $relativePath
        $stagedPath = Join-Path (Join-Path $stagingDirectory 'watt') $relativePath
        Copy-Item -LiteralPath $sourcePath -Destination $stagedPath -Force
    }

    Remove-Item -LiteralPath $zipPath -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $checksumPath -Force -ErrorAction SilentlyContinue
    Compress-Archive -LiteralPath (Join-Path $stagingDirectory 'watt') -DestinationPath $zipPath -Force

    New-Item -ItemType Directory -Path $extractionDirectory -Force | Out-Null
    Expand-Archive -LiteralPath $zipPath -DestinationPath $extractionDirectory -Force

    $actualFiles = @(Get-ChildItem -LiteralPath $extractionDirectory -File -Recurse | ForEach-Object {
        Get-NormalizedRelativePath -BasePath $extractionDirectory -Path $_.FullName
    } | Sort-Object)
    $expectedArchiveFiles = @($expectedFiles | ForEach-Object { "watt/$_" } | Sort-Object)
    if (($actualFiles -join "`n") -ne ($expectedArchiveFiles -join "`n")) {
        throw "Agent Skill ZIP has an unexpected file list: $($actualFiles -join ', ')"
    }

    foreach ($relativePath in $expectedFiles) {
        Assert-ByteForByteEqual -ExpectedPath (Join-Path $resolvedSourceDirectory $relativePath) -ActualPath (Join-Path (Join-Path $extractionDirectory 'watt') $relativePath)
    }

    $hash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    [IO.File]::WriteAllText($checksumPath, "$hash  watt-agent-skill.zip`n", [Text.Encoding]::ASCII)
}
finally {
    Remove-Item -LiteralPath $stagingDirectory -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $extractionDirectory -Force -Recurse -ErrorAction SilentlyContinue
}

[pscustomobject]@{
    ZipPath = $zipPath
    ChecksumPath = $checksumPath
}
