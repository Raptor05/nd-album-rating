# nd-album-rating

[![Build](https://github.com/Raptor05/nd-album-rating/actions/workflows/build.yml/badge.svg)](https://github.com/Raptor05/nd-album-rating/actions/workflows/build.yml)
[![Latest release](https://img.shields.io/github/v/release/Raptor05/nd-album-rating)](https://github.com/Raptor05/nd-album-rating/releases/latest)

A [Navidrome](https://www.navidrome.org/) plugin that automatically calculates album ratings from track ratings.

## What it does

- Sets the album rating to the **average of the track ratings, rounded half up** (3.5 → 4).
- Only rates an album once a **minimum share of its tracks is rated** (default 75 %).
  For example, with 3 of 18 tracks rated (17 %) the album is not rated; with 18 tracks at 75 %, at least 14 rated tracks are required.
- Optionally **removes the album rating** when track ratings are removed and the share drops below the minimum (default on).
- Works **per user**, because ratings in Navidrome are user-specific.
- Writes ratings through Navidrome's internal Subsonic API (`setRating`).

## Requirements

- Navidrome with the plugin system enabled (`ND_PLUGINS_ENABLED=true`). Use a current release; the plugin system is still marked experimental and its API evolves.

## Installation

1. Download `album-rating.ndp` from the [latest release](https://github.com/Raptor05/nd-album-rating/releases/latest).
2. Copy it into Navidrome's plugins folder (default: `<DataFolder>/plugins`).
3. In the Navidrome web UI, open the plugin, grant **User Access**, adjust the settings and enable it.

See [`HOW_TO_SETUP.md`](HOW_TO_SETUP.md) for detailed instructions, including Docker.

## Configuration

| Setting | Default | Description |
|---|---|---|
| Minimum share of rated tracks (%) | 75 | Album rating is only calculated above this share |
| Remove the album rating when track ratings are removed … | on | Removes plugin-set album ratings when the share drops below the minimum due to removed track ratings |
| Check interval (minutes) | 5 | How often ratings are checked (1–1440) |

## Important notes

- **Polling, not events:** Navidrome does not currently offer plugins an event for rating changes. The plugin checks at the configured interval, so changes appear with a short delay.
- **Manual album ratings are overwritten** as long as the threshold is reached. If you change such an album rating by hand, the next run resets it to the calculated value.
- Manual album ratings on albums **below** the threshold are never touched.
- Each user must be granted access in the plugin's **User Access** settings.

For the detailed behavior, see [`DETAILS.md`](DETAILS.md).

## Troubleshooting

- Plugin log lines contain `plugin=album-rating`. Set `ND_PLUGINS_LOGLEVEL=debug` for step-by-step output.
- `no users authorized` → grant users under **User Access**.
- `setRating … subsonic error 50` → the user is not authorized for the plugin.
- `Scheduler callback failed … context deadline exceeded` → a single API call took too long; please open an issue with the debug log.

## Building from source

```sh
make test      # run unit tests natively
make package   # build album-rating.ndp (requires Go and zip)
make docker    # same, but inside a golang container
```

Pushing a tag like `v1.2.0` triggers the release workflow, which publishes the `.ndp` file as a GitHub release. See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

GPL-3.0 (the Navidrome PDK this plugin builds on is GPL-3.0 licensed).
