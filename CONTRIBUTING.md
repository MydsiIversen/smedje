# Contributing to Smedje

Thanks for your interest in contributing.

## Getting started

```bash
git clone https://github.com/smedje/smedje.git
cd smedje
make test
```

## Development

- Go 1.23+ required
- Run `make test` before submitting a PR
- Run `make lint` to check for issues
- Follow conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`

## Adding a generator

1. Create a new file under `pkg/forge/<category>/<name>.go`
2. Implement the `forge.Generator` interface
3. Call `forge.Register()` in your `init()` function
4. Add a corresponding CLI command in `cmd/smedje/`
5. Add tests in `<name>_test.go`

The generator self-registers via `init()`, so it appears in the CLI, library,
and benchmark suite automatically.

## Code style

- `gofmt` is the law
- Comments explain why, not what
- Doc comments on every public symbol
- No new dependencies without justification

## License

By contributing, you agree that your contributions will be licensed under
the AGPL-3.0 license.
