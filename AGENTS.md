## Learned User Preferences

- Do not commit built binaries (`dist/` or `*.exe`); CI produces the Windows package.
- Do not retune wet-concrete `Solid` or HP unless explicitly approved; leave the known spawn-vs-writer mismatch in place.
- When asked to put changes on GitHub, include all remaining uncommitted work except secrets, binaries, and artifacts.
- PRs auto-merge when the Windows package CI is green (`go test ./...` plus the GUI build). Do not merge by hand unless auto-merge is stuck.

## Learned Workspace Facts

- GitHub repo is `viralarchitect/permit-denied-game` with default branch `main`.
- This worktree is an untrusted git directory; use a per-command `safe.directory` override (`C:/Users/viral/PERMIT-DENIED`) instead of changing git config.
- Local `build.bat` writes a console `dist\permitdenied.exe`; CI builds with `-H windowsgui`, uploads artifact `permitdenied-windows-amd64`, and attaches the exe on `v*` tags.
