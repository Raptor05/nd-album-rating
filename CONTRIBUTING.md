# Contributing

## Project layout

| File | Purpose |
|---|---|
| `main.go` | Entry point, scheduler callback routing, logging |
| `init_wasm.go` | `nd_on_init` export (WASM-only, excluded from native tests) |
| `scan.go` | Run state machine (albums → songs → apply) |
| `rating.go` | Pure calculation logic |
| `config.go` | Reading and validating the plugin settings |
| `subsonic.go` | Subsonic API calls and response parsing |
| `store.go` | KVStore helpers |
| `manifest.json` | Plugin metadata, permissions and configuration form |
| `.github/workflows/` | CI build and release workflows |

## Development

```sh
make test      # unit tests (native Go)
make package   # build album-rating.ndp
```

Tests run natively because the Navidrome Go PDK ships stubs for non-WASM builds. Code that is only
valid for the WASM target belongs in files with the `//go:build wasip1` constraint.

## Updating the Navidrome PDK

The PDK has no tagged releases and is referenced by a pseudo-version:

```sh
go get github.com/navidrome/navidrome/plugins/pdk/go@master
go mod tidy
```

## Releasing

1. Commit your changes to `main` and make sure the **Build** workflow is green.
2. Create and push a tag:

   ```sh
   git tag v1.2.0
   git push origin v1.2.0
   ```

3. The **Release** workflow sets the manifest version from the tag, builds the plugin and
   publishes a GitHub release with `album-rating.ndp` attached.
