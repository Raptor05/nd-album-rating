package main

import "testing"

func TestRoundHalfUp(t *testing.T) {
	cases := []struct{ sum, count, want int }{
		{7, 2, 4},  // 3.5 → 4
		{13, 4, 3}, // 3.25 → 3
		{15, 4, 4}, // 3.75 → 4
		{5, 1, 5},  // 5
		{3, 2, 2},  // 1.5 → 2
		{10, 4, 3}, // 2.5 → 3
		{0, 0, 0},  // no ratings
	}
	for _, c := range cases {
		if got := RoundHalfUp(c.sum, c.count); got != c.want {
			t.Errorf("RoundHalfUp(%d,%d) = %d, want %d", c.sum, c.count, got, c.want)
		}
	}
}

func TestMeetsThreshold(t *testing.T) {
	cases := []struct {
		rated, total, pct int
		want              bool
	}{
		{3, 18, 75, false},  // 16.7 %
		{13, 18, 75, false}, // 72.2 %
		{14, 18, 75, true},  // 77.8 %
		{3, 4, 75, true},    // exactly 75 %
		{0, 10, 1, false},
		{1, 0, 75, false},
		{10, 10, 100, true},
	}
	for _, c := range cases {
		if got := MeetsThreshold(c.rated, c.total, c.pct); got != c.want {
			t.Errorf("MeetsThreshold(%d,%d,%d) = %v, want %v", c.rated, c.total, c.pct, got, c.want)
		}
	}
}

var defaultRatingSettings = RatingSettings{ThresholdPercent: 75, RemoveWhenBelow: true}

func TestPlanSetsRatingWhenThresholdReached(t *testing.T) {
	albums := map[string]AlbumInfo{"a": {SongCount: 4, Rating: 0}}
	acc := map[string]Acc{"a": {Sum: 11, Count: 3}} // 3.67 → 4
	actions, managed := Plan(albums, acc, Managed{}, defaultRatingSettings)
	if len(actions) != 1 || actions[0].NewRating != 4 {
		t.Fatalf("expected rating 4, got %+v", actions)
	}
	if managed["a"] != 3 {
		t.Fatalf("album should be managed, got %+v", managed)
	}
}

func TestPlanOverwritesManualRating(t *testing.T) {
	albums := map[string]AlbumInfo{"a": {SongCount: 4, Rating: 1}}
	acc := map[string]Acc{"a": {Sum: 20, Count: 4}}
	actions, _ := Plan(albums, acc, Managed{}, defaultRatingSettings)
	if len(actions) != 1 || actions[0].NewRating != 5 {
		t.Fatalf("manual rating should be overwritten, got %+v", actions)
	}
}

func TestPlanNoChangeWhenRatingMatches(t *testing.T) {
	albums := map[string]AlbumInfo{"a": {SongCount: 4, Rating: 4}}
	acc := map[string]Acc{"a": {Sum: 16, Count: 4}}
	actions, _ := Plan(albums, acc, Managed{}, defaultRatingSettings)
	if len(actions) != 0 {
		t.Fatalf("expected no change, got %+v", actions)
	}
}

func TestPlanLeavesUnmanagedManualRatingAlone(t *testing.T) {
	// 3 of 18 rated, album manually rated 5 → leave untouched
	albums := map[string]AlbumInfo{"a": {SongCount: 18, Rating: 5}}
	acc := map[string]Acc{"a": {Sum: 6, Count: 3}}
	actions, managed := Plan(albums, acc, Managed{}, defaultRatingSettings)
	if len(actions) != 0 || len(managed) != 0 {
		t.Fatalf("manual rating must not be removed: %+v %+v", actions, managed)
	}
}

func TestPlanRemovesWhenTrackRatingRemoved(t *testing.T) {
	albums := map[string]AlbumInfo{"a": {SongCount: 4, Rating: 4}}
	acc := map[string]Acc{"a": {Sum: 8, Count: 2}} // previously 3 rated
	actions, managed := Plan(albums, acc, Managed{"a": 3}, defaultRatingSettings)
	if len(actions) != 1 || actions[0].NewRating != 0 || actions[0].PrevRated != 3 {
		t.Fatalf("expected removal, got %+v", actions)
	}
	if _, ok := managed["a"]; ok {
		t.Fatalf("album should no longer be managed")
	}
}

func TestPlanKeepsRatingWhenRemovalDisabled(t *testing.T) {
	s := RatingSettings{ThresholdPercent: 75, RemoveWhenBelow: false}
	albums := map[string]AlbumInfo{"a": {SongCount: 4, Rating: 4}}
	acc := map[string]Acc{"a": {Sum: 8, Count: 2}}
	actions, managed := Plan(albums, acc, Managed{"a": 3}, s)
	if len(actions) != 0 {
		t.Fatalf("expected no change, got %+v", actions)
	}
	if managed["a"] != 2 {
		t.Fatalf("album should stay managed with current state: %+v", managed)
	}
}

func TestPlanDoesNotRemoveWhenOnlySongCountGrew(t *testing.T) {
	// album gained tracks after a rescan, no rating was removed → keep
	albums := map[string]AlbumInfo{"a": {SongCount: 8, Rating: 4}}
	acc := map[string]Acc{"a": {Sum: 12, Count: 3}}
	actions, managed := Plan(albums, acc, Managed{"a": 3}, defaultRatingSettings)
	if len(actions) != 0 {
		t.Fatalf("expected no removal, got %+v", actions)
	}
	if managed["a"] != 3 {
		t.Fatalf("album should stay managed: %+v", managed)
	}
}

func TestPlanDropsDeletedAlbums(t *testing.T) {
	_, managed := Plan(map[string]AlbumInfo{}, map[string]Acc{}, Managed{"gone": 5}, defaultRatingSettings)
	if len(managed) != 0 {
		t.Fatalf("deleted album should be dropped from managed set: %+v", managed)
	}
}
