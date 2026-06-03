# Getting Started

Current release: [0.0.5](https://github.com/JoaoOliveira889/monostack)

## Installation

Choose one of the installation methods from the [README](../README.md). There are four options available:

1. **Homebrew** (recommended for macOS/Linux)
2. **Pre-built binary** from GitHub Releases
3. **`go install`** for Go developers
4. **Build locally and keep it on your PATH** for custom builds

> Requires Go 1.26.3 or later.

## First Run

Run `monostack` in your terminal. The first launch creates a default config at `~/.config/monostack/config.json` pointing to `http://localhost:4566` (LocalStack).

If you choose the local build option, install the binary into a directory on your `PATH` once and rebuild it locally whenever you update the source. A simple workflow is `make install-local`, which writes `monostack` to `~/bin/monostack`.

## Configuration

Press `5` to open the Settings panel. You can configure:

1. **Profile name** — A friendly name for this connection profile (e.g., "MiniStack", "Production")
2. **Endpoint URL** — The AWS endpoint URL. Leave empty for real AWS, or set to `http://localhost:4566` for LocalStack
3. **Region** — The AWS region (e.g., `us-east-1`)
4. **Access Key ID** — Your AWS access key
5. **Secret Access Key** — Your AWS secret key
6. **Mock mode** — Toggle online/offline mode. When enabled, all AWS calls return simulated data without any network connection
7. **Enabled Services** — Comma-separated list of panels to show, for example `s3,sqs,sns,secrets`
8. **Panel layout** — Each service panel remembers its own split ratio, and new or re-enabled panels open at `50/50`

## Connecting to AWS

### Real AWS

1. Open the Settings panel (tab `5`)
2. Clear the Endpoint URL field
3. Enter your AWS region, access key ID, and secret access key
4. Alternatively, leave credentials empty to use the default AWS SDK credential chain (environment variables, `~/.aws/credentials`, IAM roles)

### LocalStack / MiniStack

1. Open the Settings panel (tab `5`)
2. Set the Endpoint URL to `http://localhost:4566` (or your LocalStack endpoint)
3. Enter `test` as the access key ID and any value as the secret key

### Mock Mode

Toggle Mock Mode to `true` in Settings to run entirely offline with simulated AWS responses. Useful for development, demos, and testing without network access or AWS credentials.

## Selecting Services

Use the number keys (`1`-`5`) or click on tab labels to switch between services:

| Tab | Service |
|-----|---------|
| `1` | S3 Explorer — Buckets and objects |
| `2` | SQS Queues — Queues and messages |
| `3` | SNS Topics — Topics and subscriptions |
| `4` | Secrets Manager — Secrets and versions |
| `5` | Settings — Connection profiles |
