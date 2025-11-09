# AI-Assisted Development Context

This directory contains documentation and reference materials for AI-assisted development on Steampipe.

## Contents

- **reference/testing-conventions.md** - Testing standards and best practices for Steampipe tests

## Purpose

This directory provides context and conventions for AI coding assistants to maintain consistency when contributing to the Steampipe codebase. The materials here help ensure that AI-generated code follows project standards.

## Testing Infrastructure

Steampipe's test infrastructure is located in `pkg/test/`:
- **Mocks** - Mock implementations of core interfaces (db_client, plugin_manager)
- **Helpers** - Test utilities for config, database, and filesystem operations
- **Documentation** - See `pkg/test/README.md` for detailed usage

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/db/db_local/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

When adding tests:
1. Follow conventions in `reference/testing-conventions.md`
2. Use test infrastructure from `pkg/test/`
3. Use testify/assert for assertions
4. Keep tests fast, focused, and maintainable
5. Add integration tests to `tests/integration/` when needed

## Reference

- Testing conventions: `reference/testing-conventions.md`
- Test infrastructure: `pkg/test/README.md`
- Integration tests: `tests/integration/`
