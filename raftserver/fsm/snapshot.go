package fsm

import (
	"encoding/json"
	"sync"

	"github.com/hashicorp/raft"
)

type kvSnapshot struct {
	store *sync.Map
}

func newKVSnapshot(store *sync.Map) (raft.FSMSnapshot, error) {
	return &kvSnapshot{store: store}, nil
}

func (s *kvSnapshot) Persist(sink raft.SnapshotSink) error {
	data := make(map[string]string)
	s.store.Range(func(key, value interface{}) bool {
		k, okK := key.(string)
		v, okV := value.(string)
		if okK && okV {
			data[k] = v
		}
		return true
	})

	encoded, err := json.Marshal(data)
	if err != nil {
		sink.Cancel()
		return err
	}
	if _, err := sink.Write(encoded); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *kvSnapshot) Release() {}

// Legacy no-op snapshot type kept for compatibility with older code paths.
type snapshot struct{}

func (s snapshot) Persist(_ raft.SnapshotSink) error { return nil }
func (s snapshot) Release()                          {}

func newSnapshot() (raft.FSMSnapshot, error) {
	return &snapshot{}, nil
}
