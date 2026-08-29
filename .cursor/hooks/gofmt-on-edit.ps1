# afterFileEdit / afterTabFileEdit — gofmt changed Go files and mark them dirty.
. "$PSScriptRoot\_lib.ps1"

try {
    $data = Read-HookInput
    $filePath = Get-HookFilePath $data
    if (-not $filePath) {
        Write-HookJson @{}
        exit 0
    }

    if ($filePath -notmatch '\.go$') {
        Write-HookJson @{}
        exit 0
    }

    if (-not (Test-PathUnderRepo $filePath) -or -not (Test-Path -LiteralPath $filePath)) {
        Write-HookJson @{}
        exit 0
    }

    $gofmt = Get-Command gofmt -ErrorAction SilentlyContinue
    if ($gofmt) {
        $prev = $ErrorActionPreference
        $ErrorActionPreference = 'SilentlyContinue'
        try {
            & gofmt -w $filePath | Out-Null
        } finally {
            $ErrorActionPreference = $prev
        }
    }

    Write-VerifyState -DirtyGo $true -LastGoFile (Get-RepoRelativePath $filePath)
} catch {
    # fail open
}

Write-HookJson @{}
exit 0
