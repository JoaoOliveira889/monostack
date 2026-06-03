# Features

## S3 Explorer

Browse S3 buckets and their objects in a split-panel view. Features include:
- List all buckets with virtual-scrolled navigation
- Browse objects with size, last modified date, and storage class
- Download objects to `~/Downloads/monostack/`
- Delete individual objects or entire buckets
- Open object URLs in your browser via presigned URLs
- Create new buckets
- Upload files to buckets
- Preview file contents in a resizable modal

## SQS Queues

Manage SQS queues with real-time message counts:
- List queues with available, delayed, and in-flight message counts
- Inspect queue details with recent messages, routing links, and related errors
- Peek messages (up to 5 at a time)
- Send custom JSON payloads to queues
- Purge entire queues
- Create and delete queues
- View SNS subscriptions targeting each queue

## SNS Topics

Publish and manage SNS topics:
- List topics with ARNs
- Inspect topic routes with incoming/outgoing subscription context and filter summaries
- Publish messages with customizable subjects and message attributes
- Create and delete topics
- Subscribe topics to SQS queues or other SNS topics
- Batch subscribe multiple topics at once
- Import topic-scoped YAML subscriptions for `SNS → SQS` routing, with optional `default_queue` and `default_filter_scope`
- Edit subscription filter policies

## Secrets Manager

Inspect and manage AWS Secrets Manager secrets:
- List secrets with metadata and version counts
- Inspect versions and current value
- Create, update, delete, and restore secrets
- Promote specific versions (`r` key)
- Keep secret values masked until explicitly revealed
- Clipboard copy requires confirmation to prevent secret leakage

## Configuration Profiles

Save and switch between AWS connection profiles:
- LocalStack / MiniStack endpoints
- Real AWS via SDK credential chain
- Mock mode for offline testing
- Persistent JSON config at `~/.config/monostack/config.json` with `0600` permissions

## Panel Layout Persistence

Each service panel remembers its own split ratio independently:
- New or re-enabled service panels open at `50/50`
- Resizing one panel only updates that service's stored ratio
- Saving Settings preserves the existing per-panel layout preferences
- The Secrets panel now returns with its own last-used split instead of borrowing a global ratio

## Mock Mode

Run entirely offline with simulated AWS responses. Useful for:
- Development without AWS access
- Demos and presentations
- Testing TUI behavior without network calls

## Snapshot Export/Import

Export your entire environment to a YAML file for later restoration:
- Export includes S3 buckets (with full object contents), SQS queues (with attributes), SNS topics (with subscriptions), secrets metadata and values, and managed subscriptions
- Import restores all resources to the configured AWS endpoint
- Useful for recreating development environments after restarting LocalStack
- **Security note:** Snapshots contain sensitive data including S3 object contents and secret values. Snapshots are stored with `0600` permissions. Credentials (access_key_id and secret_access_key) are stripped from snapshot exports. Treat snapshot YAML files as sensitive.

## Command Log

Toggle a dedicated panel (`o` key) to inspect:
- A temporary in-memory history of every executed AWS command
- Raw command output
- Useful for debugging AWS API calls and responses
