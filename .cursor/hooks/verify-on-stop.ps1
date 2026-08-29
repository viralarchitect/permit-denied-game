# stop — if Go was edited this session and tests did not pass, run them once.
. "$PSScriptRoot\_lib.ps1"

try {
    $data = Read-HookInput
    if ($null -eq $data -or $data.status -ne 'completed') {
        Write-HookJson @{}
        exit 0
    }

    $loopCount = 0
    if ($null -ne $data.loop_count) {
        $loopCount = [int]$data.loop_count
    }
    if ($loopCount -ge 1) {
        Write-HookJson @{}
        exit 0
    }

    $state = Read-VerifyState
    $dirty = $false
    if ($state -and $state.dirtyGo) {
        $dirty = [bool]$state.dirtyGo
    }
    if (-not $dirty) {
        Write-HookJson @{}
        exit 0
    }

    $file = [string]$state.lastGoFile
    $where = if ($file) { " Last Go edit: $file." } else { '' }
    $msg = "Go files were edited this session and ``go test ./...`` has not passed since then.$where Run ``go test ./...`` now. Keep TestForwardVector, TestMultTable, TestSpawnWOneSecond, TestSpawnAHalfSecond, and TestBladeDownDropsSheriffHP green."
    Write-HookJson @{ followup_message = $msg }
    exit 0
} catch {
    Write-HookJson @{}
    exit 0
}
