# Cursor hooks

Windows PowerShell hooks for this repo. Cursor reloads `.cursor/hooks.json` on save. Inspect them under **Settings → Hooks**, or the **Hooks** output channel.

Launcher flags (`-WindowStyle Hidden`, no `cmd /c`) keep hook windows from stealing focus. Scripts fail open unless they explicitly return `deny`.

| Event | Script | What it does |
|---|---|---|
| `sessionStart` | `session-start.ps1` | Injects a short one-lot briefing |
| `afterFileEdit` / `afterTabFileEdit` | `gofmt-on-edit.ps1` | `gofmt -w` on edited `.go` files; marks Go dirty |
| `afterShellExecution` | `record-go-test.ps1` | Clears dirty when `go test` / `build.bat` passed |
| `beforeShellExecution` | `shell-guard.ps1` | Blocks committing `dist/` / `*.exe`; asks on force-push |
| `stop` | `verify-on-stop.ps1` | One follow-up to run `go test ./...` if Go is still dirty |

Runtime state: `.cursor/hooks/state/verify.json` (not committed).
