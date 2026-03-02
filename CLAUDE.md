# Project Guidelines

## Branch Workflow

- Create a **separate branch** for every fix/feature before making changes.
- Branch naming: `fix/<short-description>` for bugs, `feat/<short-description>` for features.
- Base branches off `v6` (the main development branch).
- Each branch should be committed and a PR created against `v6`.

## Code Review

- After completing each fix, run a review agent to check for:
  - Redundant or duplicated code
  - Elegance and idiomatic Go style
  - Correctness and edge cases
  - Whether the change would pass review by a senior/staff engineer
- Do not consider a fix done until it passes review.

## Project Context

- Performance-optimized fork of `vmihailenco/msgpack/v5`, maintained by Basekick Labs.
- Built for Arc, a high-performance time-series database.
- Upstream module path (`github.com/vmihailenco/msgpack/v5`) preserved for drop-in compatibility.
- Go version: 1.26+

## Testing

- Run `go test ./...` to validate changes.
- Run `go test -race ./...` for race condition checks.
- Test files use testify/suite pattern (`MsgpackTest` suite).
- Type tests in `types_test.go` use `typeTest` structs.

## Code Style

- Keep changes minimal and focused.
- Avoid unnecessary refactoring around the fix.
- Add test cases for every bug fix.
