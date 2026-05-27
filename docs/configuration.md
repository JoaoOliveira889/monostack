# Configuration

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MONOSTACK_ACCESS_KEY` | Mock access key for endpoint mode |
| `MONOSTACK_SECRET_KEY` | Mock secret key for endpoint mode |
| `AWS_ACCESS_KEY_ID` | Standard AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Standard AWS secret key |
| `AWS_REGION` | Default AWS region |

## Config File

Located at `~/.config/monostack/config.json`. Written with `0600` permissions for security.

```json
{
  "service_name": "MiniStack",
  "endpoint_url": "http://localhost:4566",
  "region": "us-east-1",
  "access_key_id": "",
  "secret_access_key": "",
  "use_mock": false,
  "left_panel_ratio": 0.3
}
```

### Fields

| Field | Description |
|-------|-------------|
| `service_name` | Friendly name for this connection profile |
| `endpoint_url` | AWS endpoint URL. Empty for real AWS, custom URL for LocalStack |
| `region` | AWS region (e.g., `us-east-1`) |
| `access_key_id` | AWS access key ID |
| `secret_access_key` | AWS secret access key |
| `use_mock` | When `true`, all AWS calls return simulated data |
| `left_panel_ratio` | Width ratio of the left panel in the TUI (0.0 to 1.0) |

## YAML Subscription Import

Subscription YAML is stored per SNS topic. The editor opens in the context of the selected topic, and `topic` can be omitted from entries that belong to that topic.

```yaml
version: 1

subscriptions:
  - name: pix
    topic: dev-webapi-pix-sns
    event_type:
      - pix_received
```

`queue` is optional per entry. Monostack uses `queue`, then `default_queue`, then a sibling queue inferred from the topic name (`-sns` → `-sqs`). `default_filter_scope` is optional and falls back to `message_body` when omitted.

When `default_filter_scope` is set to `message_attributes`, the filter policy applies to SNS message attributes. When set to `message_body`, the filter checks within the JSON body.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--version` | `false` | Show version information |

## Editor Integration

Monostack uses your default browser for presigned S3 URLs via the `open` command (macOS) or `xdg-open` (Linux).
