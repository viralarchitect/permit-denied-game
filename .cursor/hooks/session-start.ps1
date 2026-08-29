# sessionStart — inject a short operational briefing.
. "$PSScriptRoot\_lib.ps1"

$data = Read-HookInput
$mode = if ($data -and $data.composer_mode) { [string]$data.composer_mode } else { 'agent' }

$context = @"
PERMIT DENIED ($mode session). One-lot Go + Ebitengine gauntlet. Same strip every run.

Commands: ``go test ./...`` then ``go run ./cmd/permitdenied``. Windows package: ``build.bat`` -> ``dist\permitdenied.exe`` (do not commit the exe).

Tune numbers in ``internal/game/const.go``. Design = ``PERMIT_DENIED.md``. Numbers/APIs = ``PERMIT_DENIED_GO_IMPLEMENTATION.md``. Art contract = ``assets/usable/ASSETS.md``.

Stay on the strip: no campaign, no between-run meta, no second lot, no title KILLDOZER. A threat that does not change steering does not ship.
"@

Write-HookJson @{ additional_context = $context }
exit 0
