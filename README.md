# go-unraid-clean

Typed CLI to generate a reviewable cleanup list from Plex activity (via Tautulli) and media inventory (via Sonarr/Radarr), then apply deletions after review.

## Why Cobra

Cobra gives you:
- Discoverable UX (subcommands, structured help, examples)
- Maintainable command structure as this grows (scan/apply/list/verify)
- Built-in shell completion later if you want it

## Quick Start

```bash
cp configs/config.example.yaml config.yaml

go build ./cmd/go-unraid-clean
./go-unraid-clean scan --config config.yaml --out review.json --sort size --order desc
./go-unraid-clean scan --config config.yaml --csv review.csv --table --sort gap --order desc
./go-unraid-clean apply --config config.yaml --in review.json
./go-unraid-clean apply --config config.yaml --in review.json --confirm
./go-unraid-clean interactive --config config.yaml --in review.json
./go-unraid-clean scan --config config.yaml --csv review.csv --table -v
```

## Config

See `configs/config.example.yaml`.

### Rules

- `activity_min_percent`: minimum percent complete in Tautulli history to count as activity.
- `inactivity_days_after_watch`: if any watch activity is older than this, the item is eligible.
- `never_watched_days_since_added`: if no watch activity and added is older than this, the item is eligible.
- `low_watch_min_added_days`: only consider low-watch rule if added this many days ago.
- `low_watch_max_hours`: max total watch hours to qualify as low-watch.
- `low_watch_require`: when true, only include items that match low-watch (can still use other reasons for labeling).
- `series_ended_only`: only include series whose status is `ended`.

### Exceptions

Use `exceptions` to keep favorites from ever being listed. You can exclude by IDs, titles, or path prefixes.
IDs are most reliable; titles are matched case-insensitively after normalization.

### Sorting

Use `--sort` to control ordering in the report and `--order` for direction.

Supported sort keys:
- `size` (size on disk)
- `added` (date added)
- `gap` (days between added and first watch; if never watched, uses age since added)
- `last_activity` (timestamp of last watch activity)
- `inactivity` (days since last activity; if never watched, uses age since added)

### Interactive Review

Use `interactive` to step through items one-by-one and choose actions:
- skip (no changes for this item)
- always-ignore (adds to exceptions in config)
- delete entirely
- delete files only (keep movie/show entry)
- keep last season (series only)

Each item shows top viewers (up to 2) with combined watch hours.

### Verbose Logging

Use `-v` to enable verbose logging:

```bash
./go-unraid-clean scan --config config.yaml --csv review.csv --table -v
```

## Status

- `scan` produces a review report based on Tautulli + Sonarr/Radarr data.
- `apply` prints a summary and requires `--confirm` to delete items.

Next step is to wire the API clients and rule engine.
