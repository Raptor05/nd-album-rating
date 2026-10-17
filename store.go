package main

import (
	"encoding/json"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
)

// KVStore keys
const (
	kvInterval      = "interval"
	kvStatePrefix   = "state:"
	kvManagedPrefix = "managed:"
	kvLockPrefix    = "lock:"
)

func loadState(user string) (*scanState, bool) {
	data, ok, err := host.KVStoreGet(kvStatePrefix + user)
	if err != nil || !ok {
		return nil, false
	}
	var st scanState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, false
	}
	if st.Albums == nil {
		st.Albums = map[string]AlbumInfo{}
	}
	if st.Acc == nil {
		st.Acc = map[string]Acc{}
	}
	if st.NewManaged == nil {
		st.NewManaged = Managed{}
	}
	return &st, true
}

func saveState(st *scanState) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return host.KVStoreSet(kvStatePrefix+st.User, data)
}

func loadManaged(user string) Managed {
	m := Managed{}
	data, ok, err := host.KVStoreGet(kvManagedPrefix + user)
	if err != nil || !ok {
		return m
	}
	_ = json.Unmarshal(data, &m)
	return m
}

func saveManaged(user string, m Managed) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return host.KVStoreSet(kvManagedPrefix+user, data)
}

func refreshLock(user string) {
	_ = host.KVStoreSetWithTTL(kvLockPrefix+user, []byte("1"), lockTTLSeconds)
}

// finishRun removes the run state and the lock for a user.
func finishRun(user string) {
	_ = host.KVStoreDelete(kvStatePrefix + user)
	_ = host.KVStoreDelete(kvLockPrefix + user)
}
