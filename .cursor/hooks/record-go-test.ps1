# afterShellExecution — clear the dirty-Go flag when tests pass.
. "$PSScriptRoot\_lib.ps1"

try {
    $data = Read-HookInput
    if ($null -eq $data) {
        Write-HookJson @{}
        exit 0
    }

    $command = [string]$data.command
    $output = [string]$data.output
    $isTest = $command -match '(?i)go\s+test' -or $command -match '(?i)build\.bat'
    if (-not $isTest) {
        Write-HookJson @{}
        exit 0
    }

    $failed = $output -match '(?m)^FAIL\b' -or $output -match '(?i)tests failed'
    if (-not $failed) {
        $state = Read-VerifyState
        Write-VerifyState -DirtyGo $false -LastGoFile $state.lastGoFile
    }
} catch {
    # fail open
}

Write-HookJson @{}
exit 0
