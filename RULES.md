# Monostack Rules

These are project-specific rules for Monostack. For shared workspace rules, see [../RULES.md](../RULES.md).

## AWS-specific
- Use AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2/...`) exclusively.
- All AWS clients must support dynamic endpoint routing via `aws.EndpointResolverWithOptionsFunc` for LocalStack/MiniStack compatibility.
- Pass explicit `context.Context` from adapters to SDK calls for cancellation and timeout control.
- Return typed/wrapped domain errors from adapters so the TUI layer can display styled error banners.

## Project-specific
- AWS adapters live in `internal/adapters/aws/`.
- TUI adapters live in `internal/adapters/tui/`.
- Config storage lives in `internal/pkg/config/`.
- Shared UI styles live in `internal/pkg/ui/`.
- Build: `make build` produces `./monostack`. Install locally: `make install-local` (outputs to `~/bin/monostack`).
- Release: `.goreleaser.yaml` handles multi-platform builds, Homebrew tap, and changelog generation.
