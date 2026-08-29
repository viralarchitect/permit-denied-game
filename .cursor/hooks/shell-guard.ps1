# beforeShellExecution — keep binaries out of git and ask on force-push.
. "$PSScriptRoot\_lib.ps1"

function Deny([string]$UserMessage, [string]$AgentMessage) {
    Write-HookJson @{
        permission    = 'deny'
        user_message  = $UserMessage
        agent_message = $AgentMessage
    }
    exit 0
}

function Ask([string]$UserMessage, [string]$AgentMessage) {
    Write-HookJson @{
        permission    = 'ask'
        user_message  = $UserMessage
        agent_message = $AgentMessage
    }
    exit 0
}

try {
    $data = Read-HookInput
    $command = if ($data) { [string]$data.command } else { '' }
    if ([string]::IsNullOrWhiteSpace($command)) {
        Write-HookJson @{ permission = 'allow' }
        exit 0
    }

    $isGit = $command -match '(?i)\bgit\b'
    if ($isGit -and $command -match '(?i)\b(add|commit)\b' -and $command -match '(?i)(\.exe\b|[\\/]dist[\\/]|\bdist\\|\bdist/)') {
        Deny `
            'Do not stage or commit built binaries. dist\ and *.exe stay out of git (see .gitignore).' `
            'Hook denied git add/commit of dist/ or *.exe. Build with build.bat; CI uploads the Windows artifact. Do not commit the exe.'
    }

    if ($isGit -and $command -match '(?i)\badd\b' -and $command -match '(?i)(\s\.(?:\s|$)|-A\b|--all\b)') {
        $exeHits = @()
        $rootExe = Join-Path $script:RepoRoot 'permitdenied.exe'
        $distExe = Join-Path $script:RepoRoot 'dist\permitdenied.exe'
        if (Test-Path -LiteralPath $rootExe) { $exeHits += 'permitdenied.exe' }
        if (Test-Path -LiteralPath $distExe) { $exeHits += 'dist\permitdenied.exe' }
        if ($exeHits.Count -gt 0) {
            Ask `
                "This git add may pick up $($exeHits -join ', '). Confirm they stay untracked." `
                "A hook flagged git add of the whole tree while $($exeHits -join ', ') exists. Leave binaries untracked."
        }
    }

    if ($isGit -and $command -match '(?i)\bpush\b' -and $command -match '(?i)(--force\b|-f\b)') {
        Ask `
            'Force-push requires your review.' `
            'A hook is asking the user to confirm this force-push.'
    }
} catch {
    Write-HookJson @{ permission = 'allow' }
    exit 0
}

Write-HookJson @{ permission = 'allow' }
exit 0
