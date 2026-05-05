# cronwatch

Lightweight daemon that monitors cron job execution and sends alerts on missed or failed runs.

## Installation

```bash
go install github.com/cronwatch/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/cronwatch/cronwatch.git && cd cronwatch && make build
```

## Usage

Define your monitored jobs in `cronwatch.yaml`:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert:
      slack: "#alerts"
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap an existing cron job to report its exit status:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will send an alert if the job does not execute within its expected window or exits with a non-zero status.

## Configuration

| Field | Description |
|-------|-------------|
| `schedule` | Standard cron expression defining expected run time |
| `timeout` | Maximum allowed duration before the job is considered failed |
| `alert` | Notification target (`email`, `slack`, or `webhook`) |

## License

MIT