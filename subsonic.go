package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
)

type albumJSON struct {
	ID         string `json:"id"`
	SongCount  int    `json:"songCount"`
	UserRating int    `json:"userRating"`
}

type songJSON struct {
	ID         string `json:"id"`
	AlbumID    string `json:"albumId"`
	UserRating int    `json:"userRating"`
}

type subsonicResponse struct {
	Response struct {
		Status string `json:"status"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		AlbumList2 *struct {
			Album []albumJSON `json:"album"`
		} `json:"albumList2"`
		SearchResult3 *struct {
			Song []songJSON `json:"song"`
		} `json:"searchResult3"`
	} `json:"subsonic-response"`
}

func callSubsonic(uri string) (*subsonicResponse, error) {
	raw, err := host.SubsonicAPICall(uri)
	if err != nil {
		return nil, err
	}
	return parseSubsonic(raw)
}

func parseSubsonic(raw string) (*subsonicResponse, error) {
	var resp subsonicResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if resp.Response.Status != "ok" {
		if e := resp.Response.Error; e != nil {
			return nil, fmt.Errorf("subsonic error %d: %s", e.Code, e.Message)
		}
		return nil, errors.New("subsonic response with status " + resp.Response.Status)
	}
	return &resp, nil
}

func fetchAlbums(user string, offset int) ([]albumJSON, error) {
	uri := fmt.Sprintf("getAlbumList2?type=alphabeticalByName&size=%d&offset=%d&u=%s",
		pageSize, offset, url.QueryEscape(user))
	resp, err := callSubsonic(uri)
	if err != nil {
		return nil, err
	}
	if resp.Response.AlbumList2 == nil {
		return nil, nil
	}
	return resp.Response.AlbumList2.Album, nil
}

func fetchSongs(user string, offset, mode int) ([]songJSON, error) {
	query := "%22%22" // ""
	if mode == queryModeEmpty {
		query = ""
	}
	uri := fmt.Sprintf("search3?query=%s&artistCount=0&albumCount=0&songCount=%d&songOffset=%d&u=%s",
		query, pageSize, offset, url.QueryEscape(user))
	resp, err := callSubsonic(uri)
	if err != nil {
		return nil, err
	}
	if resp.Response.SearchResult3 == nil {
		return nil, nil
	}
	return resp.Response.SearchResult3.Song, nil
}

// setRating sets the album rating (0 = remove). Navidrome accepts album IDs for setRating.
func setRating(user, albumID string, value int) error {
	uri := fmt.Sprintf("setRating?id=%s&rating=%d&u=%s",
		url.QueryEscape(albumID), value, url.QueryEscape(user))
	_, err := callSubsonic(uri)
	return err
}
