## Summary

<!-- Why this change exists. Link the issue if there is one. -->

Fixes #

## Scope

- [ ] Same lot, same run loop (no campaign, meta, second strip, or second player)
- [ ] New numbers live in `internal/game/const.go`
- [ ] No `dist/` or `*.exe` in the diff
- [ ] Copy stays on-spec (`PERMIT DENIED`, no Killdozer / Heemeyer / wanted stars)

## Test plan

- [ ] `go test ./...`
- [ ] Control tests still green (`TestForwardVector`, `TestMultTable`, `TestSpawnWOneSecond`, `TestSpawnAHalfSecond`, `TestBladeDownDropsSheriffHP`)
- [ ] Played a run (or N/A — docs-only)

Show-off tests touched by this PR:

- [ ] N/A
- [ ] New player can steer, toggle blade, and die inside 90s
- [ ] Plant pip is enough to judge time
- [ ] Coward (`×1.0`) vs two-target (`×1.6`) still cash out differently
- [ ] Heat / plates readable from the sprite
- [ ] Rubble can still steal a street
