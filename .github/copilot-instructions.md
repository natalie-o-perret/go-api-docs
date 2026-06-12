# Copilot code review instructions for go-api-docs

## General Go guidelines

- Flag any use of `interface{}` or `any` where a concrete typed alternative exists.
- Prefer explicit error handling over panics; every `error` return must be checked.
- Ensure exported functions, types, and constants have doc comments.
- Use table-driven tests; flag test functions that lack sub-test coverage for branching logic.
- Warn on direct use of `os.Exit` or `log.Fatal` outside of `main`.

## OpenAPI / codegen concerns

- Validate that generated code stays in sync with the spec: flag manual edits inside auto-generated files.
- Ensure route handlers return appropriate HTTP status codes for all code paths, including error paths.
- Check that request validation errors return `400 Bad Request`, not `500`.
- Flag missing `Content-Type` header assertions in handler tests.

## Dependency and module hygiene

- Warn if `go.sum` changes without a corresponding `go.mod` change.
- Flag new indirect dependencies that could be promoted to direct.
- Flag use of `replace` directives in `go.mod` unless clearly justified.

## Security

- Flag any handler that reads user-supplied input without sanitisation before passing it to file paths, shell commands, or SQL queries.
- Warn on the use of `math/rand` where `crypto/rand` is appropriate.
- Ensure TLS configuration does not set `InsecureSkipVerify: true` in production paths.
