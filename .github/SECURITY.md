# Security

This game is a local Ebitengine binary. There is no game server, account system, or network protocol to attack.

## Supported builds

| Build | Supported |
|---|---|
| Latest `main` / `master` | Yes |
| Latest `v*` GitHub Release | Yes |
| Older tags and local forks | No |

## Report a vulnerability

Use **Security → Report a vulnerability** on this repository (GitHub private advisory). Do not open a public issue for:

- Secrets in the repo or in CI (`GITHUB_TOKEN`, release tooling)
- A crash or file-path issue that looks exploitable in the Windows package

Include OS, Go version, how you built or downloaded the binary, and steps to reproduce.

Engine bugs in Ebitengine belong upstream at [hajimehoshi/ebiten](https://github.com/hajimehoshi/ebiten).
