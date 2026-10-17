package main

import "testing"

func TestParseSubsonicAlbums(t *testing.T) {
	raw := `{"subsonic-response":{"status":"ok","albumList2":{"album":[
		{"id":"al1","songCount":12,"userRating":4},
		{"id":"al2","songCount":8}
	]}}}`
	resp, err := parseSubsonic(raw)
	if err != nil {
		t.Fatal(err)
	}
	albums := resp.Response.AlbumList2.Album
	if len(albums) != 2 || albums[0].UserRating != 4 || albums[1].UserRating != 0 {
		t.Fatalf("unexpected albums: %+v", albums)
	}
}

func TestParseSubsonicSongs(t *testing.T) {
	raw := `{"subsonic-response":{"status":"ok","searchResult3":{"song":[
		{"id":"s1","albumId":"al1","userRating":5}
	]}}}`
	resp, err := parseSubsonic(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s := resp.Response.SearchResult3.Song; len(s) != 1 || s[0].AlbumID != "al1" {
		t.Fatalf("unexpected songs: %+v", s)
	}
}

func TestParseSubsonicError(t *testing.T) {
	raw := `{"subsonic-response":{"status":"failed","error":{"code":50,"message":"not authorized"}}}`
	if _, err := parseSubsonic(raw); err == nil {
		t.Fatal("expected error")
	}
}
