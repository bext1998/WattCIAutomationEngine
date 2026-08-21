# Run with: powershell -NoProfile -ExecutionPolicy Bypass -File .\tests\install.tests.ps1
# This test suite dot-sources the installer and exercises only pure/test-local helpers.
. (Join-Path $PSScriptRoot '..\install.ps1')

$script:passed = 0
$script:failed = 0

function Assert-Equal {
    param($Actual, $Expected, [string]$Message)
    if ($Actual -ne $Expected) {
        throw "$Message. Expected: [$Expected]; actual: [$Actual]"
    }
}

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw $Message
    }
}

function Assert-Throws {
    param([scriptblock]$Action, [string]$ExpectedMessage, [string]$Message)
    try {
        & $Action
    }
    catch {
        if ($_.Exception.Message -notlike "*$ExpectedMessage*") {
            throw "$Message. Expected error containing [$ExpectedMessage]; actual: [$($_.Exception.Message)]"
        }
        return
    }
    throw "$Message. Expected an error."
}

function Test-Case {
    param([string]$Name, [scriptblock]$Body)
    try {
        & $Body
        $script:passed++
        Write-Host "PASS $Name"
    }
    catch {
        $script:failed++
        Write-Error "FAIL ${Name}: $($_.Exception.Message)"
    }
}

Test-Case 'selects the first eligible snapshot prerelease in API order' {
    $releases = @(
        [pscustomobject]@{ draft = $false; prerelease = $false; tag_name = 'v9.0.0'; assets = @() },
        [pscustomobject]@{ draft = $true; prerelease = $true; tag_name = 'v8.0.0-snapshot'; assets = @() },
        [pscustomobject]@{ draft = $false; prerelease = $true; tag_name = 'v0.0.10-snapshot'; assets = @() },
        [pscustomobject]@{ draft = $false; prerelease = $true; tag_name = 'v0.0.9-snapshot'; assets = @() }
    )

    Assert-Equal (Select-WattSnapshotRelease -Releases $releases).tag_name 'v0.0.10-snapshot' 'Release selection'
}

Test-Case 'rejects release lists without an eligible snapshot' {
    $releases = @([pscustomobject]@{ draft = $false; prerelease = $true; tag_name = 'v1.0.0-rc1'; assets = @() })
    Assert-Throws { Select-WattSnapshotRelease -Releases $releases } 'No snapshot prerelease' 'Release selection failure'
}

Test-Case 'compares snapshot versions numerically rather than lexically' {
    Assert-True ((Compare-WattSnapshotVersion -Left 'v0.0.10-snapshot' -Right 'v0.0.2-snapshot') -gt 0) 'v0.0.10 must be newer than v0.0.2'
    Assert-True ((Compare-WattSnapshotVersion -Left 'v1.2.3-snapshot.10' -Right 'v1.2.3-snapshot.2') -gt 0) 'snapshot sequence must be numeric'
    Assert-Equal (Compare-WattSnapshotVersion -Left 'v1.2.3-snapshot' -Right 'v1.2.3-snapshot.0') 0 'omitted snapshot sequence equals zero'
}

Test-Case 'rejects malformed snapshot versions' {
    Assert-Throws { ConvertTo-WattSnapshotVersion -Tag 'v1.2-snapshot' } 'not a valid snapshot version' 'Version validation'
}

Test-Case 'fails closed on malformed or mismatched checksums without touching the installation file' {
    $tempDirectory = Join-Path ([IO.Path]::GetTempPath()) ('watt-installer-test-' + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tempDirectory | Out-Null
    try {
        $download = Join-Path $tempDirectory 'download.exe'
        $installed = Join-Path $tempDirectory 'watt.exe'
        $checksum = Join-Path $tempDirectory 'watt.exe.sha256'
        [IO.File]::WriteAllText($download, 'downloaded bytes')
        [IO.File]::WriteAllText($installed, 'known good installed bytes')
        [IO.File]::WriteAllText($checksum, ('0' * 64) + '  watt.exe')

        Assert-Throws { Assert-WattChecksum -FilePath $download -ChecksumPath $checksum } 'SHA-256 verification failed' 'Checksum mismatch'
        Assert-Equal ([IO.File]::ReadAllText($installed)) 'known good installed bytes' 'Checksum failure must preserve installed file'

        [IO.File]::WriteAllText($checksum, 'not a checksum')
        Assert-Throws { Assert-WattChecksum -FilePath $download -ChecksumPath $checksum } 'checksum file format is invalid' 'Malformed checksum'
    }
    finally {
        Remove-Item -LiteralPath $tempDirectory -Force -Recurse -ErrorAction SilentlyContinue
    }
}

Test-Case 'adds the install directory once while preserving existing User PATH entries and order' {
    $path = 'C:\Tools;C:\Other\\'
    $result = Add-WattPathEntry -PathValue $path -Entry 'C:\Users\Alice\AppData\Local\Programs\Watt'
    Assert-Equal $result.Value 'C:\Tools;C:\Other\\;C:\Users\Alice\AppData\Local\Programs\Watt' 'New User PATH value'
    Assert-True $result.Changed 'New entry should be marked changed'

    $existing = Add-WattPathEntry -PathValue 'C:\TOOLS;C:\Users\Alice\AppData\Local\Programs\WATT\\;C:\Other' -Entry 'C:\Users\Alice\AppData\Local\Programs\Watt'
    Assert-Equal $existing.Value 'C:\TOOLS;C:\Users\Alice\AppData\Local\Programs\WATT\\;C:\Other' 'Existing User PATH entries must not be rewritten'
    Assert-True (-not $existing.Changed) 'Case and trailing slash variants must be idempotent'
}

Test-Case 'requires both official release assets' {
    $release = [pscustomobject]@{
        assets = @([pscustomobject]@{ name = 'watt.exe'; browser_download_url = 'https://example.invalid/watt.exe' })
    }
    Assert-Throws { Get-WattReleaseAssets -Release $release } 'watt.exe.sha256' 'Missing checksum asset'
}

Write-Host "`n$script:passed passed; $script:failed failed"
if ($script:failed -ne 0) {
    exit 1
}
