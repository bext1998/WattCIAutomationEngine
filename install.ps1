# Watt installer for Windows PowerShell 5.1 and PowerShell 7+.
# This script intentionally installs only the latest eligible snapshot prerelease.

$ErrorActionPreference = 'Stop'

$script:WattRepository = 'bext1998/WattCIAutomationEngine'
$script:WattReleasesApi = 'https://api.github.com/repos/bext1998/WattCIAutomationEngine/releases'
$script:WattSnapshotTagPattern = '^v[0-9]+\.[0-9]+\.[0-9]+-snapshot(?:\.[0-9]+)?$'

function ConvertTo-WattSnapshotVersion {
    param([Parameter(Mandatory = $true)][string]$Tag)

    if ($Tag -notmatch '^v([0-9]+)\.([0-9]+)\.([0-9]+)-snapshot(?:\.([0-9]+))?$') {
        throw "Release tag '$Tag' is not a valid snapshot version."
    }

    return [pscustomobject]@{
        Major = [int64]$Matches[1]
        Minor = [int64]$Matches[2]
        Patch = [int64]$Matches[3]
        Snapshot = if ($Matches[4]) { [int64]$Matches[4] } else { [int64]0 }
    }
}

function Compare-WattSnapshotVersion {
    param(
        [Parameter(Mandatory = $true)][string]$Left,
        [Parameter(Mandatory = $true)][string]$Right
    )

    $leftVersion = ConvertTo-WattSnapshotVersion -Tag $Left
    $rightVersion = ConvertTo-WattSnapshotVersion -Tag $Right
    foreach ($part in @('Major', 'Minor', 'Patch', 'Snapshot')) {
        if ($leftVersion.$part -lt $rightVersion.$part) { return -1 }
        if ($leftVersion.$part -gt $rightVersion.$part) { return 1 }
    }
    return 0
}

function Select-WattSnapshotRelease {
    param([Parameter(Mandatory = $true)][object[]]$Releases)

    foreach ($release in $Releases) {
        if ($null -eq $release) { continue }
        if ($release.draft -eq $true) { continue }
        if ($release.prerelease -ne $true) { continue }
        if ([string]$release.tag_name -notmatch $script:WattSnapshotTagPattern) { continue }
        return $release
    }

    throw 'No snapshot prerelease is available from the official Watt GitHub Releases list.'
}

function Get-WattReleaseAssets {
    param([Parameter(Mandatory = $true)]$Release)

    $exeAsset = $null
    $checksumAsset = $null
    foreach ($asset in @($Release.assets)) {
        if ($asset.name -eq 'watt.exe') { $exeAsset = $asset }
        if ($asset.name -eq 'watt.exe.sha256') { $checksumAsset = $asset }
    }

    if ($null -eq $exeAsset) {
        throw "Release '$($Release.tag_name)' is missing the required watt.exe asset."
    }
    if ($null -eq $checksumAsset) {
        throw "Release '$($Release.tag_name)' is missing the required watt.exe.sha256 asset."
    }

    foreach ($asset in @($exeAsset, $checksumAsset)) {
        $uri = [Uri]$asset.browser_download_url
        $expectedPathPrefix = "/$script:WattRepository/releases/download/"
        if ($uri.Scheme -ne 'https' -or $uri.Host -ine 'github.com' -or -not $uri.AbsolutePath.StartsWith($expectedPathPrefix, [StringComparison]::Ordinal)) {
            throw "Release '$($Release.tag_name)' supplied an unexpected download URL for $($asset.name)."
        }
    }

    return [pscustomobject]@{ Exe = $exeAsset; Checksum = $checksumAsset }
}

function Invoke-WattDownload {
    param(
        [Parameter(Mandatory = $true)][string]$Uri,
        [Parameter(Mandatory = $true)][string]$OutFile
    )

    $parameters = @{
        Uri = $Uri
        OutFile = $OutFile
        Headers = @{ 'User-Agent' = 'WattInstaller/1.0' }
        ErrorAction = 'Stop'
    }
    if ((Get-Command Invoke-WebRequest).Parameters.ContainsKey('UseBasicParsing')) {
        $parameters.UseBasicParsing = $true
    }
    Invoke-WebRequest @parameters | Out-Null
}

function Assert-WattChecksum {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string]$ChecksumPath
    )

    $checksumContents = [IO.File]::ReadAllText($ChecksumPath).Trim()
    $match = [regex]::Match($checksumContents, '^([0-9A-Fa-f]{64})\s+\*?watt\.exe\s*$')
    if (-not $match.Success) {
        throw 'The watt.exe.sha256 checksum file format is invalid.'
    }

    $expectedHash = $match.Groups[1].Value
    $actualHash = (Get-FileHash -LiteralPath $FilePath -Algorithm SHA256).Hash
    if (-not [string]::Equals($expectedHash, $actualHash, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'SHA-256 verification failed for downloaded watt.exe; the existing installation was not changed.'
    }
}

function Get-WattInstalledVersion {
    param([Parameter(Mandatory = $true)][string]$ExePath)

    try {
        $versionOutput = & $ExePath --version 2>&1
        $exitCode = $LASTEXITCODE
    }
    catch {
        throw "Existing Watt executable '$ExePath' could not be executed. Remove or repair it manually before retrying."
    }

    $versionText = ($versionOutput | Out-String).Trim()
    if ($exitCode -ne 0 -or $versionText -notmatch '^watt\s+(v[0-9]+\.[0-9]+\.[0-9]+-snapshot(?:\.[0-9]+)?)$') {
        throw "Existing Watt executable '$ExePath' did not return a valid snapshot version. Remove or repair it manually before retrying."
    }
    return $Matches[1]
}

function Normalize-WattPathEntry {
    param([AllowEmptyString()][string]$Value)

    $normalized = $Value.Trim()
    if ($normalized -match '^[A-Za-z]:\\+$') {
        return $normalized.Substring(0, 3)
    }
    return $normalized.TrimEnd([char]92)
}

function Add-WattPathEntry {
    param(
        [AllowNull()][string]$PathValue,
        [Parameter(Mandatory = $true)][string]$Entry
    )

    $normalizedEntry = Normalize-WattPathEntry -Value $Entry
    foreach ($existingEntry in @($PathValue -split ';')) {
        if ([string]::Equals((Normalize-WattPathEntry -Value $existingEntry), $normalizedEntry, [StringComparison]::OrdinalIgnoreCase)) {
            return [pscustomobject]@{ Value = $PathValue; Changed = $false }
        }
    }

    if ([string]::IsNullOrEmpty($PathValue)) {
        return [pscustomobject]@{ Value = $Entry; Changed = $true }
    }
    return [pscustomobject]@{ Value = "$PathValue;$Entry"; Changed = $true }
}

function Ensure-WattUserPath {
    param([Parameter(Mandatory = $true)][string]$InstallDirectory)

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $userPathResult = Add-WattPathEntry -PathValue $userPath -Entry $InstallDirectory
    if ($userPathResult.Changed) {
        try {
            [Environment]::SetEnvironmentVariable('Path', $userPathResult.Value, 'User')
        }
        catch {
            throw "Watt was installed, but the User PATH could not be updated: $($_.Exception.Message)"
        }
    }

    $processPathResult = Add-WattPathEntry -PathValue $env:Path -Entry $InstallDirectory
    if ($processPathResult.Changed) {
        $env:Path = $processPathResult.Value
    }
    return $userPathResult.Changed
}

function Install-Watt {
    if ($env:OS -ne 'Windows_NT') {
        throw 'Watt installation is supported only on Windows.'
    }

    # Windows PowerShell 5.1 may otherwise negotiate an unsupported TLS version with GitHub.
    if ($PSVersionTable.PSVersion.Major -lt 6) {
        [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
    }

    try {
        $releases = Invoke-RestMethod -Uri $script:WattReleasesApi -Headers @{ 'User-Agent' = 'WattInstaller/1.0' } -ErrorAction Stop
    }
    catch {
        throw "Could not query the official Watt GitHub Releases list: $($_.Exception.Message)"
    }
    $release = Select-WattSnapshotRelease -Releases @($releases)
    $assets = Get-WattReleaseAssets -Release $release
    $targetVersion = [string]$release.tag_name

    $localAppData = [Environment]::GetFolderPath([Environment+SpecialFolder]::LocalApplicationData)
    if ([string]::IsNullOrWhiteSpace($localAppData)) {
        throw 'Could not determine LOCALAPPDATA for the current user.'
    }
    $installDirectory = Join-Path $localAppData 'Programs\Watt'
    $installPath = Join-Path $installDirectory 'watt.exe'

    $existingVersion = $null
    $existingHash = $null
    if (Test-Path -LiteralPath $installPath -PathType Leaf) {
        $existingVersion = Get-WattInstalledVersion -ExePath $installPath
        try {
            $existingHash = (Get-FileHash -LiteralPath $installPath -Algorithm SHA256).Hash
        }
        catch {
            throw "Existing Watt executable '$installPath' could not be hashed. Close processes using it or repair it manually before retrying."
        }

        $comparison = Compare-WattSnapshotVersion -Left $existingVersion -Right $targetVersion
        if ($comparison -gt 0) {
            throw "Installed Watt $existingVersion is newer than release $targetVersion; refusing to downgrade."
        }
    }

    $temporaryDirectory = Join-Path ([IO.Path]::GetTempPath()) ('WattInstall-' + [Guid]::NewGuid().ToString('N'))
    try {
        New-Item -ItemType Directory -Path $temporaryDirectory -Force | Out-Null
        $downloadedExe = Join-Path $temporaryDirectory 'watt.exe'
        $downloadedChecksum = Join-Path $temporaryDirectory 'watt.exe.sha256'
        Invoke-WattDownload -Uri $assets.Exe.browser_download_url -OutFile $downloadedExe
        Invoke-WattDownload -Uri $assets.Checksum.browser_download_url -OutFile $downloadedChecksum
        Assert-WattChecksum -FilePath $downloadedExe -ChecksumPath $downloadedChecksum
        $downloadedHash = (Get-FileHash -LiteralPath $downloadedExe -Algorithm SHA256).Hash

        $needsInstall = $true
        if ($null -ne $existingVersion -and $existingVersion -eq $targetVersion -and [string]::Equals($existingHash, $downloadedHash, [StringComparison]::OrdinalIgnoreCase)) {
            $needsInstall = $false
            Write-Host "Watt $targetVersion is already installed and its SHA-256 checksum matches the release."
        }

        if ($needsInstall) {
            if ($null -ne $existingVersion) {
                Write-Host "Updating Watt $existingVersion to $targetVersion..."
            }
            else {
                Write-Host "Installing Watt $targetVersion..."
            }

            New-Item -ItemType Directory -Path $installDirectory -Force | Out-Null
            $replacementPath = Join-Path $installDirectory ('.watt.exe.' + [Guid]::NewGuid().ToString('N') + '.new')
            try {
                [IO.File]::Copy($downloadedExe, $replacementPath, $true)
                if (Test-Path -LiteralPath $installPath -PathType Leaf) {
                    [IO.File]::Replace($replacementPath, $installPath, $null)
                }
                else {
                    [IO.File]::Move($replacementPath, $installPath)
                }
            }
            catch {
                throw "Could not replace '$installPath'. It may be in use by another process; close processes using Watt and retry. $($_.Exception.Message)"
            }
            finally {
                Remove-Item -LiteralPath $replacementPath -Force -ErrorAction SilentlyContinue
            }
        }
    }
    finally {
        Remove-Item -LiteralPath $temporaryDirectory -Force -Recurse -ErrorAction SilentlyContinue
    }

    $pathChanged = Ensure-WattUserPath -InstallDirectory $installDirectory
    try {
        $installedVersionOutput = & $installPath --version 2>&1
        $installedVersionExitCode = $LASTEXITCODE
    }
    catch {
        throw "Watt was installed but could not be verified at '$installPath': $($_.Exception.Message)"
    }
    $installedVersionText = ($installedVersionOutput | Out-String).Trim()
    if ($installedVersionExitCode -ne 0 -or $installedVersionText -ne "watt $targetVersion") {
        throw "Watt was installed but '$installPath --version' returned '$installedVersionText' instead of 'watt $targetVersion'."
    }

    Write-Host "Watt $targetVersion is ready at $installPath."
    if ($pathChanged) {
        Write-Host 'Added Watt to the User PATH. Restart other open terminals before using watt there.'
    }
    else {
        Write-Host 'Watt is already present in the User PATH. Restart other open terminals if they do not see it yet.'
    }
}

if ($MyInvocation.InvocationName -ne '.') {
    try {
        Install-Watt
    }
    catch {
        Write-Error "Watt installation failed: $($_.Exception.Message)"
        exit 1
    }
}
