// Album Rating – a Navidrome plugin
//
// Automatically calculates album ratings from the track ratings of each user.
//
// Navidrome does not (yet) offer a plugin event for rating changes, so the plugin checks
// the ratings at a configurable interval. Each run is split into small steps with a fixed
// number of API calls, so no callback hits Navidrome's 30-second limit. Intermediate state
// is kept in the plugin's KVStore.
//
// The plugin intentionally does not measure time itself (time.Now), because the clock
// inside the WASM sandbox is not reliable. Intervals use the Navidrome scheduler, and
// locks use KVStore entries with a TTL (the host measures expiry).
//
// Capabilities:  Lifecycle (nd_on_init, see init_wasm.go), SchedulerCallback
// Host services: Scheduler, KVStore, SubsonicAPI, Users
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/navidrome/navidrome/plugins/pdk/go/scheduler"
)

const (
	logPrefix = "album-rating: "

	tickScheduleID    = "album-rating-tick"
	payloadTick       = "tick"
	payloadScanPrefix = "scan:"
)

type albumRatingPlugin struct{}

func init() {
	scheduler.Register(&albumRatingPlugin{})
}

func main() {}

// OnCallback handles scheduler events.
func (p *albumRatingPlugin) OnCallback(req scheduler.SchedulerCallbackRequest) error {
	switch {
	case req.Payload == payloadTick:
		return onTick()
	case strings.HasPrefix(req.Payload, payloadScanPrefix):
		return continueScan(strings.TrimPrefix(req.Payload, payloadScanPrefix))
	default:
		logWarn("unknown payload %q", req.Payload)
		return nil
	}
}

// onStart is called from nd_on_init once the plugin is loaded.
func onStart() error {
	s := loadSettings()
	if err := scheduleTick(s.IntervalMinutes); err != nil {
		return fmt.Errorf("could not start scheduler: %w", err)
	}
	// start the first run shortly after loading
	if _, err := host.SchedulerScheduleOneTime(15, payloadTick, ""); err != nil {
		logWarn("could not schedule first run: %v", err)
	}
	logInfo("started (interval %d min, threshold %d %%, remove when below %v)",
		s.IntervalMinutes, s.ThresholdPercent, s.RemoveWhenBelow)
	return nil
}

// scheduleTick (re)starts the recurring tick with the configured interval.
func scheduleTick(minutes int) error {
	_ = host.SchedulerCancelSchedule(tickScheduleID)
	if _, err := host.SchedulerScheduleRecurring(fmt.Sprintf("@every %dm", minutes), payloadTick, tickScheduleID); err != nil {
		return err
	}
	_ = host.KVStoreSet(kvInterval, []byte(strconv.Itoa(minutes)))
	return nil
}

// Logging helpers. All messages are prefixed so they are easy to grep in the Navidrome log.
func logDebug(format string, args ...any) { pdk.Log(pdk.LogDebug, logPrefix+fmt.Sprintf(format, args...)) }
func logInfo(format string, args ...any)  { pdk.Log(pdk.LogInfo, logPrefix+fmt.Sprintf(format, args...)) }
func logWarn(format string, args ...any)  { pdk.Log(pdk.LogWarn, logPrefix+fmt.Sprintf(format, args...)) }
func logError(format string, args ...any) { pdk.Log(pdk.LogError, logPrefix+fmt.Sprintf(format, args...)) }
