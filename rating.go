package main

// Pure rating logic. No host calls in this file, so it is fully unit-testable.

// AlbumInfo describes the current state of an album for one user.
type AlbumInfo struct {
	SongCount int `json:"c"` // number of tracks on the album
	Rating    int `json:"r"` // current album rating of the user (0 = none)
}

// Acc accumulates the track ratings of one album.
type Acc struct {
	Sum   int `json:"s"` // sum of stars
	Count int `json:"n"` // number of rated tracks
}

// Managed tracks which album ratings were set by this plugin, together with the
// number of rated tracks at the time of the last calculation. Key = album ID.
type Managed map[string]int

// RatingSettings are the plugin settings relevant for the calculation.
type RatingSettings struct {
	ThresholdPercent int  // minimum share of rated tracks in percent (1–100)
	RemoveWhenBelow  bool // remove album rating when track ratings are removed and the share drops below the threshold
}

// Action is a change to apply to an album rating.
type Action struct {
	AlbumID   string `json:"id"`
	NewRating int    `json:"r"`           // 0 = remove rating
	PrevRated int    `json:"p,omitempty"` // for removals: previous rated-track count (used to retry on failure)
}

// RoundHalfUp returns the average rounded half up (3.5 → 4), clamped to 1–5.
func RoundHalfUp(sum, count int) int {
	if count <= 0 {
		return 0
	}
	// floor(sum/count + 0.5) using integers only, to avoid float inaccuracies
	r := (2*sum + count) / (2 * count)
	if r < 1 {
		r = 1
	}
	if r > 5 {
		r = 5
	}
	return r
}

// MeetsThreshold reports whether enough tracks are rated.
// Example: 18 tracks at 75 % → at least 14 rated tracks (13.5 rounded up).
func MeetsThreshold(rated, total, percent int) bool {
	if total <= 0 || rated <= 0 {
		return false
	}
	return rated*100 >= percent*total
}

// Plan computes the required changes for all albums of a user.
//
// Rules:
//   - Threshold reached → album rating = rounded average. Always overwrites, including
//     manually set album ratings. The album is marked as managed by the plugin.
//   - Threshold not reached, album is managed and track ratings were removed since the
//     last run → remove the album rating (if RemoveWhenBelow is enabled).
//   - Threshold not reached and album was never rated by the plugin → leave untouched
//     (manual album ratings are preserved).
func Plan(albums map[string]AlbumInfo, acc map[string]Acc, managed Managed, s RatingSettings) ([]Action, Managed) {
	actions := []Action{}
	newManaged := Managed{}

	for id, album := range albums {
		a := acc[id]
		rated := a.Count
		if rated > album.SongCount && album.SongCount > 0 {
			rated = album.SongCount
		}

		if MeetsThreshold(rated, album.SongCount, s.ThresholdPercent) {
			target := RoundHalfUp(a.Sum, a.Count)
			newManaged[id] = rated
			if album.Rating != target {
				actions = append(actions, Action{AlbumID: id, NewRating: target})
			}
			continue
		}

		prev, wasManaged := managed[id]
		if !wasManaged {
			continue
		}

		ratingRemoved := rated < prev
		if ratingRemoved && s.RemoveWhenBelow {
			if album.Rating != 0 {
				actions = append(actions, Action{AlbumID: id, NewRating: 0, PrevRated: prev})
			}
			continue // no longer managed
		}
		// keep the rating; album stays managed with its current state
		newManaged[id] = rated
	}

	return actions, newManaged
}
