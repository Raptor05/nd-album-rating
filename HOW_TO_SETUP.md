# How to set up

## 1. Enable the plugin system

### Docker Compose

```yaml
services:
  navidrome:
    image: deluan/navidrome:latest
    environment:
      ND_PLUGINS_ENABLED: "true"
      # optional, default is <DataFolder>/plugins:
      # ND_PLUGINS_FOLDER: /data/plugins
      # optional, for troubleshooting:
      # ND_PLUGINS_LOGLEVEL: debug
    volumes:
      - ./data:/data
      - /path/to/music:/music:ro
```

### navidrome.toml

```toml
[Plugins]
Enabled = true
# Folder = "/path/to/plugins"
# LogLevel = "debug"
```

## 2. Install the plugin

1. Download `album-rating.ndp` from the latest GitHub release
   (or from the artifacts of a successful **Build** workflow run).
2. Copy it into the plugins folder – with the Compose example above: `./data/plugins/`.
3. Restart Navidrome, or click **Rescan** in the plugins section of the web UI.

> Keep the file name `album-rating.ndp`. Navidrome uses the file name as the plugin ID and
> stores configuration, user access and plugin data under that ID.

## 3. Configure the plugin

In the Navidrome web UI, go to the admin area → **Plugins** → **Album Rating**:

1. **User Access:** enable *Allow all users* or select the users whose albums should be rated.
   The plugin cannot be enabled without this.
2. **Configuration:** adjust the minimum share, the removal checkbox and the check interval.
3. **Enable** the plugin.

The first run starts about 15 seconds after enabling. After that, the plugin runs at the
configured interval.

## 4. Verify

With `ND_PLUGINS_LOGLEVEL=debug` you should see lines like:

```
album-rating: started (interval 5 min, threshold 75 %, remove when below true)
album-rating: alice – albums from 0: received 500
album-rating: alice – tracks from 0: received 500
album-rating: alice – 812 albums checked, 37 changes planned
album-rating: alice – 37 albums rated, 0 ratings removed, 0 errors
```

## Updating

Replace `album-rating.ndp` in the plugins folder with the new version and restart Navidrome
(or click **Rescan**). Settings and plugin data are kept because the plugin ID stays the same.
