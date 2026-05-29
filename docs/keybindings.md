# Keybindings

## Global

| Key | Action |
|-----|--------|
| `1` | S3 Explorer tab |
| `2` | SQS Queues tab |
| `3` | SNS Topics tab |
| `4` | Secrets Manager tab |
| `5` | Settings tab |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `←` / `h` | Switch to left panel |
| `→` / `l` | Switch to right panel |
| `tab` | Cycle between visible panels |
| `q` / `Ctrl+C` | Quit |
| `Ctrl+P` / `?` | Toggle Help Menu |
| `o` | Toggle command logs |
| `space` | Start or extend text selection |
| `y` | Copy selected text |
| `esc` | Back / Cancel / Close |

## S3 Explorer

| Key | Action |
|-----|--------|
| `Enter` / `→` | Select bucket / enter object view |
| `Esc` / `←` | Back to bucket list |
| `b` | Open file in browser (presigned URL) |
| `d` | Download file to `~/Downloads/monostack/` |
| `x` / `Delete` | Delete file or bucket |
| `c` | Create bucket |
| `f` | Create folder prefix |
| `u` | Upload object |

## SQS Queues

| Key | Action |
|-----|--------|
| `Enter` | Inspect selected queue |
| `→` / `l` | Open queue routes (SNS subscriptions) |
| `v` | Peek messages |
| `s` | Send message |
| `p` | Purge queue |
| `P` | Purge all loaded queues |
| `c` | Create queue |
| `x` / `Delete` | Delete queue |

## SNS Topics

| Key | Action |
|-----|--------|
| `Enter` | Inspect selected topic or route |
| `→` / `l` | Open subscriptions |
| `s` | Publish event |
| `i` | Import subscription YAML |
| `c` | Create topic |
| `x` / `Delete` | Delete topic |
| `b` | Batch subscribe SNS to SQS |

## Secrets Manager

| Key | Action |
|-----|--------|
| `Enter` / `v` | Reveal selected secret value |
| `c` | Create secret |
| `u` | Update secret value |
| `r` | Restore secret |
| `d` | Delete secret |

## Settings

| Key | Action |
|-----|--------|
| `Enter` | Edit selected field |
| `Esc` | Stop editing |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `s` | Save profile |

## Profile Actions

| Key | Action |
|-----|--------|
| `E` | Export profile snapshot |
| `L` | Load/import profile snapshot |

## YAML Import

YAML imports are topic-scoped. When a subscription entry includes `queue`, Monostack creates an `SNS → SQS` subscription. When `queue` is omitted, Monostack uses `default_queue` if one is provided, then falls back to the sibling queue inferred from the topic name. Use `default_filter_scope: message_body` when the event type lives inside the message JSON body, or omit it and let Monostack default it for you.
