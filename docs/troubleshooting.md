# Troubleshooting

## Blank Screen

Ensure your terminal supports ANSI escape codes and is at least 60 columns wide.

## AWS Connection Errors

- Check your config in the Settings panel (tab `5`)
- For LocalStack, ensure it's running on the configured endpoint (`http://localhost:4566`)
- For real AWS, verify your credentials in `~/.aws/credentials` or environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- Try enabling Mock Mode for offline testing
- Verify the region matches your AWS resources

## Config File Issues

If the config file is corrupted, delete it and restart Monostack:
```bash
rm ~/.config/monostack/config.json
```

Monostack will create a fresh default config on the next launch.

## Permission Denied Errors

Monostack stores config files with `0600` permissions. If you encounter permission issues:
```bash
chmod 600 ~/.config/monostack/config.json
```

## FAQ

### How do I connect to LocalStack?
Set the Endpoint URL to `http://localhost:4566` in the Settings panel (tab `5`). Use any value for the access key and secret key.

### How do I connect to real AWS?
Clear the Endpoint URL field in Settings and enter your AWS credentials. You can also leave credentials blank to use the default AWS SDK credential chain.

### How do I add subscriptions between topics and queues?
In the SNS Topics tab, press `b` to batch subscribe topics to queues. You can also press `i` to import a YAML subscription file.

### How many messages can I peek at once?
The SQS peek command retrieves up to 5 messages at a time.

## Getting Help

Open a [GitHub Issue](https://github.com/JoaoOliveira889/monostack/issues) with:
- Your OS and terminal emulator
- Steps to reproduce the issue
- Any error messages or logs
