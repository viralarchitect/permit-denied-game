# Shared helpers for PERMIT DENIED Cursor hooks.
# Dot-source from other scripts:  . "$PSScriptRoot\_lib.ps1"

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$script:RepoRoot = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$script:VerifyStatePath = Join-Path $PSScriptRoot 'state\verify.json'

function Read-HookInput {
    $raw = [Console]::In.ReadToEnd()
    if ([string]::IsNullOrWhiteSpace($raw)) {
        return $null
    }
    try {
        return $raw | ConvertFrom-Json
    } catch {
        return $null
    }
}

function Write-HookJson {
    param($Object)
    if ($null -eq $Object) {
        Write-Output '{}'
        return
    }
    Write-Output ($Object | ConvertTo-Json -Compress -Depth 6)
}

function Get-HookFilePath {
    param($Data)
    if ($null -eq $Data) {
        return $null
    }
    foreach ($name in @('file_path', 'filePath', 'path')) {
        $value = $Data.$name
        if (-not [string]::IsNullOrWhiteSpace([string]$value)) {
            return [string]$value
        }
    }
    return $null
}

function Get-RepoRootPrefix {
    return ([System.IO.Path]::GetFullPath($script:RepoRoot).TrimEnd('\') + '\')
}

function Test-PathUnderRepo {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) {
        return $false
    }
    try {
        $full = [System.IO.Path]::GetFullPath($Path)
        return $full.StartsWith((Get-RepoRootPrefix), [System.StringComparison]::OrdinalIgnoreCase)
    } catch {
        return $false
    }
}

function Get-RepoRelativePath {
    param([string]$Path)
    try {
        $full = [System.IO.Path]::GetFullPath($Path)
        $prefix = Get-RepoRootPrefix
        if ($full.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
            return $full.Substring($prefix.Length).Replace('\', '/')
        }
        return $full
    } catch {
        return $Path
    }
}

function Read-VerifyState {
    if (-not (Test-Path -LiteralPath $script:VerifyStatePath)) {
        return [pscustomobject]@{ dirtyGo = $false; lastGoFile = $null }
    }
    try {
        return Get-Content -LiteralPath $script:VerifyStatePath -Raw | ConvertFrom-Json
    } catch {
        return [pscustomobject]@{ dirtyGo = $false; lastGoFile = $null }
    }
}

function Write-VerifyState {
    param(
        [bool]$DirtyGo,
        [string]$LastGoFile = $null
    )
    $dir = Split-Path $script:VerifyStatePath -Parent
    if (-not (Test-Path -LiteralPath $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
    $payload = @{
        dirtyGo     = $DirtyGo
        lastGoFile  = $LastGoFile
        updatedAtMs = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
    } | ConvertTo-Json -Compress
    [System.IO.File]::WriteAllText($script:VerifyStatePath, $payload)
}
