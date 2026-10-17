package main

import (
	"fmt"
	"strconv"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
)

const (
	pageSize         = 500 // items per page (getAlbumList2 / search3)
	pagesPerCallback = 2   // page fetches per callback
	ratingsPerStep   = 25  // setRating calls per callback
	lockTTLSeconds   = 600 // a run without progress is considered dead after 10 minutes

	phaseAlbums = "albums"
	phaseSongs  = "songs"
	phaseApply  = "apply"

	queryModeQuoted = 0 // search3?query=""  (two double quotes)
	queryModeEmpty  = 1 // search3?query=    (empty) – fallback
)

// scanState is the intermediate state of a run for one user.
type scanState struct {
	User       string               `json:"user"`
	Phase      string               `json:"phase"`
	Offset     int                  `json:"offset"`
	QueryMode  int                  `json:"queryMode"`
	Albums     map[string]AlbumInfo `json:"albums,omitempty"`
	Acc        map[string]Acc       `json:"acc,omitempty"`
	Actions    []Action             `json:"actions,omitempty"`
	NewManaged Managed              `json:"newManaged,omitempty"`
	Set        int                  `json:"set"`
	Removed    int                  `json:"removed"`
	Failed     int                  `json:"failed"`
}

// onTick starts a run for every authorized user unless one is already in progress.
func onTick() error {
	s := loadSettings()

	// interval changed in the configuration? → reschedule the tick
	if data, ok, err := host.KVStoreGet(kvInterval); err == nil && (!ok || string(data) != strconv.Itoa(s.IntervalMinutes)) {
		if err := scheduleTick(s.IntervalMinutes); err != nil {
			logWarn("could not update interval: %v", err)
		} else {
			logInfo("interval set to %d min", s.IntervalMinutes)
		}
	}

	users, err := host.UsersGetUsers()
	if err != nil {
		return fmt.Errorf("could not read users: %w", err)
	}
	if len(users) == 0 {
		logWarn("no users authorized – assign users in the plugin's User Access settings")
	}

	for _, u := range users {
		if locked, _ := host.KVStoreHas(kvLockPrefix + u.UserName); locked {
			logDebug("run for %s still in progress", u.UserName)
			continue
		}
		st := &scanState{
			User:   u.UserName,
			Phase:  phaseAlbums,
			Albums: map[string]AlbumInfo{},
			Acc:    map[string]Acc{},
		}
		if err := saveState(st); err != nil {
			logError("could not save state for %s: %v", u.UserName, err)
			continue
		}
		refreshLock(u.UserName)
		logDebug("run for %s started", u.UserName)
		scheduleContinuation(u.UserName)
	}
	return nil
}

// continueScan executes a bounded number of steps and schedules a continuation if needed.
func continueScan(user string) error {
	st, ok := loadState(user)
	if !ok {
		return nil
	}
	refreshLock(user)

	for i := 0; i < pagesPerCallback; i++ {
		done, err := step(st)
		if err != nil {
			logError("run for %s aborted: %v", user, err)
			finishRun(user)
			return nil
		}
		if done {
			finishRun(user)
			if st.Set > 0 || st.Removed > 0 || st.Failed > 0 {
				logInfo("%s – %d albums rated, %d ratings removed, %d errors",
					user, st.Set, st.Removed, st.Failed)
			} else {
				logDebug("%s – no changes", user)
			}
			return nil
		}
		if st.Phase == phaseApply {
			break // apply steps are already batched
		}
	}

	if err := saveState(st); err != nil {
		logError("could not save state for %s: %v", user, err)
		finishRun(user)
		return nil
	}
	scheduleContinuation(user)
	return nil
}

func scheduleContinuation(user string) {
	if _, err := host.SchedulerScheduleOneTime(1, payloadScanPrefix+user, ""); err != nil {
		logError("could not schedule continuation for %s: %v", user, err)
		finishRun(user)
	}
}

// step executes one unit of work. Returns true when the run is complete.
func step(st *scanState) (bool, error) {
	switch st.Phase {
	case phaseAlbums:
		albums, err := fetchAlbums(st.User, st.Offset)
		if err != nil {
			return false, err
		}
		logDebug("%s – albums from %d: received %d", st.User, st.Offset, len(albums))
		for _, a := range albums {
			if a.SongCount > 0 {
				st.Albums[a.ID] = AlbumInfo{SongCount: a.SongCount, Rating: a.UserRating}
			}
		}
		if len(albums) < pageSize {
			st.Phase, st.Offset = phaseSongs, 0
		} else {
			st.Offset += pageSize
		}
		return false, nil

	case phaseSongs:
		songs, err := fetchSongs(st.User, st.Offset, st.QueryMode)
		if err != nil {
			return false, err
		}
		logDebug("%s – tracks from %d: received %d", st.User, st.Offset, len(songs))
		// fallback in case the server does not treat query="" as "everything"
		if st.Offset == 0 && len(songs) == 0 && len(st.Albums) > 0 && st.QueryMode == queryModeQuoted {
			st.QueryMode = queryModeEmpty
			return false, nil
		}
		for _, s := range songs {
			if s.UserRating > 0 && s.AlbumID != "" {
				a := st.Acc[s.AlbumID]
				a.Sum += s.UserRating
				a.Count++
				st.Acc[s.AlbumID] = a
			}
		}
		if len(songs) < pageSize {
			managed := loadManaged(st.User)
			st.Actions, st.NewManaged = Plan(st.Albums, st.Acc, managed, loadSettings().RatingSettings)
			logDebug("%s – %d albums checked, %d changes planned", st.User, len(st.Albums), len(st.Actions))
			st.Albums, st.Acc = nil, nil
			st.Phase = phaseApply
		} else {
			st.Offset += pageSize
		}
		return false, nil

	case phaseApply:
		if len(st.Actions) == 0 {
			if err := saveManaged(st.User, st.NewManaged); err != nil {
				return false, err
			}
			return true, nil
		}
		n := ratingsPerStep
		if n > len(st.Actions) {
			n = len(st.Actions)
		}
		for _, act := range st.Actions[:n] {
			if err := setRating(st.User, act.AlbumID, act.NewRating); err != nil {
				st.Failed++
				logWarn("setRating for album %s failed: %v", act.AlbumID, err)
				if act.NewRating == 0 && act.PrevRated > 0 {
					st.NewManaged[act.AlbumID] = act.PrevRated // retry on the next run
				}
				continue
			}
			if act.NewRating == 0 {
				st.Removed++
			} else {
				st.Set++
			}
		}
		st.Actions = st.Actions[n:]
		return false, nil
	}
	return false, fmt.Errorf("unknown phase %q", st.Phase)
}
