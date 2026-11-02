# Contributing to MongoDB Query Exporter

Thanks for your interest in contributing! This document describes how to set up
your environment, the development workflow, and the conventions this project
follows.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer (see `go.mod` and the CI workflow)
- [Docker](https://docs.docker.com/get-docker/) (for building images and running the demo)
- [golangci-lint](https://golangci-lint.run/usage/install/) (used in CI and recommended locally)

> Mocks are generated with [mockery](https://github.com/vektra/mockery) v3, which
> `make generate-mocks` fetches on demand via `go run` &mdash; no separate install needed.

## Getting Started

Clone the repository and download the dependencies:

```bash
git clone https://github.com/ppussar/mongodb_exporter.git
cd mongodb_exporter
go mod download
```

## Development Workflow

The project uses a `Makefile` for common tasks:

| Command | Description |
|---------|-------------|
| `make test` | Run the unit tests |
| `make cover` | Run tests with race detector and generate `coverage.txt` |
| `make build` | Run tests and build the binary into `bin/mongodb_exporter` |
| `make run` | Run the exporter locally against `configuration.yaml` |
| `make image` | Build the Docker image |
| `make start-demo` | Start MongoDB + exporter via Docker Compose and scrape metrics |
| `make stop-demo` | Stop the demo containers |
| `make generate-mocks` | Regenerate the test mocks with mockery v3 (config in `.mockery.yaml`) |
| `make clean` | Remove the `bin/` directory |

### Running the demo

```bash
make start-demo
curl localhost:9090/prometheus
make stop-demo
```

## Code Style

- Format your code with `gofmt` (run `go fmt ./...` before committing).
- Keep the code passing `golangci-lint run`; the CI pipeline lints every push
  and pull request and will fail on lint errors.
- Add or update tests for any behavioral change. Test files live next to the
  code they cover (`*_test.go`).

### Regenerating mocks

Mocks for the MongoDB wrapper interfaces are generated with
[mockery](https://github.com/vektra/mockery) v3 and live in `internal/mocks`.
The interfaces to mock and the output layout are configured in `.mockery.yaml`.
After changing an interface in `internal/wrapper`, regenerate them:

```bash
make generate-mocks
```

The mockery version is pinned via `MOCKERY_VERSION` in the `Makefile`; the tool
is fetched on demand with `go run`, so no local installation is required. The
generated files carry a `DO NOT EDIT` header &mdash; do not edit them by hand;
change the interface (and `.mockery.yaml` if needed) and regenerate instead.

## Commit Messages

This project uses
[go-semantic-release](https://github.com/go-semantic-release/action), which
determines versions and generates releases from
[Conventional Commits](https://www.conventionalcommits.org/). Please format
commit messages accordingly:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Common types:

- `feat:` a new feature (triggers a minor release)
- `fix:` a bug fix (triggers a patch release)
- `docs:` documentation only changes
- `test:` adding or fixing tests
- `refactor:` code change that neither fixes a bug nor adds a feature
- `chore:` build process, tooling, or dependency changes

Breaking changes must include `BREAKING CHANGE:` in the footer or a `!` after
the type/scope (e.g. `feat!:`), which triggers a major release.

## Pull Requests

1. Fork the repository and create a topic branch from `main`.
2. Make your changes, including tests and documentation updates.
3. Ensure the full pipeline passes locally:
   ```bash
   go mod tidy
   golangci-lint run
   make build
   make cover
   ```
   Confirm `go mod tidy` leaves `go.mod` and `go.sum` unchanged, since CI checks
   this.
4. Push your branch and open a pull request against `main`.
5. Keep the PR title concise and the description focused on what changed, why,
   and how it was tested.

The CI pipeline runs lint, build, coverage, and a Docker Compose smoke test on
every pull request. Releases and image publishing happen automatically when
changes land on `main`.

## Reporting Issues

Please open an issue on GitHub with:

- A clear description of the problem or feature request
- Steps to reproduce (for bugs), including your configuration where relevant
- The exporter version and Go/Docker versions if applicable

## License

By contributing, you agree that your contributions will be licensed under the
same license as this project (see [LICENSE](LICENSE)).
