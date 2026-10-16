package raftserver

import (
	"github.com/hashicorp/raft"
)

type Server struct {
	raft *raft.Raft
	fsm  raft.FSM
}

func New(raftInstance *raft.Raft, fsm raft.FSM) *Server {
	if raftInstance == nil || fsm == nil {
		panic("raftserver.New: raft and fsm must be non-nil")
	}
	return &Server{
		raft: raftInstance,
		fsm:  fsm,
	}
}
