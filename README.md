# Distributed Key-Value Store (Raft)

A Go implementation of a sharded, in-memory key-value store that uses the [Hashicorp Raft](https://github.com/hashicorp/raft) library for replicated consensus on each shard and on the configuration service.

## Overview

The system splits keys across fixed data shards using Murmur3 hashing. Each shard runs a three-node Raft cluster for durable, consistent writes. A separate three-node Raft cluster stores shard metadata (leader Raft addresses and shard count). An HTTP **router** resolves the correct shard leader via the config service and forwards client get, put, and delete requests.

This repository is structured as a course-style distributed systems project: shard membership is registered statically at startup (see `scripts/run_system.sh`), not discovered dynamically at runtime.

## Features

- **Sharded key space** — Keys are mapped to one of N shards via Murmur3 and fixed bucket sizing (`HASH_MODULO = 16384`).
- **Per-shard Raft replication** — Data nodes (`dataserver`) replicate state through a shared finite-state machine (FSM) supporting `PUT`, `GET`, and `DEL` log commands.
- **Replicated config service** — Config nodes (`configserver`) store shard ID → leader Raft address and expose `/config` for the router.
- **HTTP router** — Single entry point on port `3000` (default) for `/get`, `/put`, `/delete`, and `/status`.
- **Cluster join helpers** — `/raft/join`, `/raft/leave`, and `/raft/status` on Raft-enabled nodes.
- **Leader notification** — Data shard leaders POST to the config service when leadership changes.

## Tech stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.20+ |
| Consensus | [hashicorp/raft](https://github.com/hashicorp/raft) v1.6.0 |
| Raft log storage | [hashicorp/raft-boltdb](https://github.com/hashicorp/raft-boltdb) |
| Hashing | [twmb/murmur3](https://github.com/twmb/murmur3) |

## Prerequisites

- **Go** 1.20 or newer (`go version`)
- **Bash**, **curl**, and **ss** (used by `scripts/run_system.sh` and `scripts/get_pid_by_port.sh`; the run script uses `sudo` for port cleanup on router restart)
- A Unix-like environment (Linux/macOS/WSL) to run the shell scripts as written

## Getting started

```bash
git clone https://github.com/vonmiller1/distributed-key-value-store-kraft.git
cd distributed-key-value-store-kraft
go build ./...
```

No environment variables are required for the default local demo. Optional flags on each binary are documented below under **Configuration**.

## Running locally

Build all packages, then start the full stack (three config nodes, three shards × three data nodes, router):

```bash
go build ./...
./scripts/run_system.sh
```

Log files are written under `logs/`. Follow a component, for example:

```bash
tail -F logs/router.log
```

**Default ports (demo layout)**

| Service | HTTP | Raft (internal) |
|---------|------|-----------------|
| Router | 3000 | — |
| Config nodes | 3001–3003 | 13001–13003 |
| Data shard 1 | 8001–8003 | 18011–18013 |
| Data shard 2 | 8011–8013 | 18021–18023 |
| Data shard 3 | 8021–8023 | 18031–18033 |

**Client helper**

```bash
./scripts/client.sh put mykey myvalue
./scripts/client.sh get mykey
./scripts/client.sh delete mykey
./scripts/client.sh status
```

**Status endpoints**

```bash
curl http://localhost:3001/config
curl http://localhost:8001/raft/status
```

Stop the stack with `Ctrl+C` in the terminal running `run_system.sh`.

## Testing

This project does not ship a separate automated test suite in-tree. After starting the cluster, exercise paths with `./scripts/client.sh` and the curl examples above. To run Go tests if you add them later:

```bash
go test ./...
```

## Project structure

```
.
├── configserver/     # Replicated shard metadata (Raft + HTTP)
├── dataserver/       # Replicated key-value shard (Raft + HTTP)
├── router/           # Murmur3 sharding and request proxy
├── raftserver/       # Shared HTTP handlers and Raft admin APIs
│   └── fsm/          # In-memory KV finite-state machine
├── scripts/          # run_system.sh, client.sh, get_pid_by_port.sh
├── go.mod
└── LICENSE
```

## Configuration

Common flags (see `flag` definitions in each `main.go`):

| Flag | Used by | Purpose |
|------|---------|---------|
| `-port` | All HTTP servers | HTTP listen port |
| `-node_id` | configserver, dataserver | Raft server ID |
| `-raft_addr` | configserver, dataserver | Raft TCP bind address |
| `-store_dir` | configserver, dataserver | BoltDB / snapshot directory (temp dir if empty) |
| `-shard_id` | dataserver | Shard number reported to config on leader change |
| `-configserverport1/2/3` | router | Fallback config HTTP ports |

Raft snapshot interval and threshold are constants in `configserver/main.go` and `dataserver/main.go` (`snapInterval`, `snapThreshold`).

## Deployment

There is no Dockerfile or CI pipeline in this repository. For production you would run compiled binaries instead of `go run`, persist `-store_dir` on disk, secure Raft and HTTP with TLS and authentication, and replace static shard registration with a controlled config API.

## Contributing

1. Fork the repository and create a feature branch from `main`.
2. Keep commits focused; use imperative messages (e.g. `fix(router): …`).
3. Open a pull request describing behavior and any operational impact.

Please do not commit secrets, real hostnames, or credentials.

## License

[MIT License](LICENSE) — Copyright (c) 2023 Ved.
