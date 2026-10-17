//go:build wasip1

package main

import xpdk "github.com/extism/go-pdk"

// nd_on_init is called once after the plugin has been loaded (Lifecycle capability).
// This file is only compiled for the WASM target, so `go test` runs natively.
//
//go:wasmexport nd_on_init
func ndOnInit() int32 {
	if err := onStart(); err != nil {
		logError("%v", err)
		_ = xpdk.OutputJSON(map[string]string{"error": err.Error()})
		return 0
	}
	_ = xpdk.OutputJSON(map[string]string{})
	return 0
}
