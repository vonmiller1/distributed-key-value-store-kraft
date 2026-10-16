package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hashicorp/raft"
)

const (
	PUT = "PUT"
	GET = "GET"
	DEL = "DEL"
)

type FSM struct {
	kv_store *sync.Map
}

func (fsm FSM) Put(key string, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value is not a string")
	}

	fsm.kv_store.Store(key, strValue)
	return nil
}

func (fsm *FSM) Get(key string) (interface{}, error) {
	value, ok := fsm.kv_store.Load(key)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	return value, nil
}

func (fsm *FSM) Delete(key string) error {
	_, ok := fsm.kv_store.Load(key)
	if !ok {
		return fmt.Errorf("key not found")
	}

	fsm.kv_store.Delete(key)
	return nil
}

type Payload struct {
	OP    string
	Key   string
	Value interface{}
}

type ApplyResponse struct {
	Error error
	Data  interface{}
}

func (fsm FSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var payload Payload
		if err := json.Unmarshal(log.Data, &payload); err != nil {
			fmt.Fprintf(os.Stderr, "error unmarshalling payload: %s\n", err.Error())
			return &ApplyResponse{Error: err}
		}

		switch payload.OP {
		case PUT:
			if err := fsm.Put(payload.Key, payload.Value); err != nil {
				return &ApplyResponse{Error: err}
			}
			return &ApplyResponse{
				Error: nil,
				Data:  payload.Value,
			}
		case GET:
			value, err := fsm.Get(payload.Key)
			if err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}
			return &ApplyResponse{
				Error: nil,
				Data:  value,
			}
		case DEL:
			if err := fsm.Delete(payload.Key); err != nil {
				return &ApplyResponse{Error: err}
			}
			return &ApplyResponse{
				Error: nil,
				Data:  nil,
			}
		default:
			return &ApplyResponse{Error: fmt.Errorf("unknown op: %s", payload.OP)}
		}
	}
	fmt.Fprintf(os.Stderr, "ignored raft log type: %d\n", log.Type)
	return nil
}

func (fsm FSM) Snapshot() (raft.FSMSnapshot, error) {
	return newKVSnapshot(fsm.kv_store)
}

func (fsm *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var data map[string]string
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		return err
	}
	fsm.kv_store = &sync.Map{}
	for k, v := range data {
		fsm.kv_store.Store(k, v)
	}
	return nil
}

func NewFSM() raft.FSM {
	return &FSM{
		kv_store: &sync.Map{},
	}
}
