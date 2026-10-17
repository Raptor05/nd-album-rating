# Details

## Calculation

- **Album rating** = average of all rated tracks of the album, rounded half up, clamped to 1–5.
  Examples: 3.25 → 3, 3.5 → 4, 3.75 → 4.
- **Unrated tracks** are ignored in the average; they only count towards the share.
- **Share** = rated tracks / total tracks of the album (all discs).
  The threshold is reached when `rated × 100 ≥ threshold × total`.

## Decision table

| Situation | Result |
|---|---|
| Threshold reached | Album rating is set or updated – including albums with a manual rating |
| Threshold not reached, album never rated by the plugin | Nothing happens |
| Plugin-rated album, track ratings removed, share below threshold, removal enabled | Album rating is removed |
| Plugin-rated album, track ratings removed, share below threshold, removal disabled | Album rating stays |
| Plugin-rated album, only new tracks added (rescan), no rating removed | Album rating stays |
| Threshold raised in the settings, album now below it | Album rating stays (no track rating was removed) |

The plugin remembers which albums it rated and how many tracks were rated at that time.
Only those album ratings can be removed by the plugin.

## Runtime behavior

1. A recurring scheduler tick fires at the configured interval. When the interval changes in
   the settings, the tick is rescheduled on its next run.
2. For each authorized user, a run is started unless one is already in progress.
3. A run has three phases, each split across multiple scheduler callbacks:
   - **albums:** all albums via `getAlbumList2` (500 per page) – ID, track count, current rating
   - **songs:** all tracks via `search3` with an empty query (500 per page) – album ID and rating
   - **apply:** only album ratings that differ from the calculated value are written via `setRating`
4. Each callback performs at most 2 page fetches or 25 `setRating` calls, then schedules its
   continuation one second later. This keeps every callback well below Navidrome's 30-second limit.

## State and locking

All state lives in the plugin's KVStore (`<DataFolder>/plugins/album-rating/kvstore.db`):

| Key | Content |
|---|---|
| `state:<user>` | intermediate state of the current run |
| `lock:<user>` | marks a running run; expires after 10 minutes without progress |
| `managed:<user>` | albums rated by the plugin and their rated-track count |
| `interval` | interval the tick is currently scheduled with |

The plugin does not use its own clock, because time inside the WASM sandbox is not reliable.
Expiry of the lock is measured by Navidrome (KVStore TTL). If Navidrome is restarted during a
run, the lock expires and the next tick starts a fresh run.

## Load

Each run reads all albums and tracks of each authorized user. These are internal calls without
network traffic. For very large libraries, increase the check interval to reduce load.
