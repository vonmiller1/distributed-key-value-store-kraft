package raftserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/raft"
)

func (s Server) RaftJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	nodeid := r.FormValue("nodeid")
	addr := r.FormValue("addr")
	if nodeid == "" || addr == "" {
		http.Error(w, "nodeid and addr are required", http.StatusBadRequest)
		return
	}
	if s.raft.State() != raft.Leader {
		http.Error(w, "not leader", http.StatusServiceUnavailable)
		return
	}

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		http.Error(w, "get raft configuration failed", http.StatusInternalServerError)
		return
	}

	f := s.raft.AddVoter(raft.ServerID(nodeid), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		http.Error(w, "add voter failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}

func (s Server) RaftStatus(w http.ResponseWriter, r *http.Request) {
	j, err := json.Marshal(s.raft.Stats())
	if err != nil {
		http.Error(w, "marshal raft status failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (s Server) RaftLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	nodeid := r.FormValue("nodeid")
	if nodeid == "" {
		http.Error(w, "nodeid is required", http.StatusBadRequest)
		return
	}
	if s.raft.State() != raft.Leader {
		http.Error(w, "not the leader", http.StatusServiceUnavailable)
		return
	}

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		http.Error(w, "get raft configuration failed", http.StatusInternalServerError)
		return
	}

	future := s.raft.RemoveServer(raft.ServerID(nodeid), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("remove node %s: %s", nodeid, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("removed successfully"))
}
