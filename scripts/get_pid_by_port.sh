#!/bin/bash

PORT=${1:?usage: $0 <port>}

if command -v ss >/dev/null 2>&1; then
    sudo ss -lptn "sport = :$PORT"
else
    echo "ss not found; install iproute2 or use lsof -i :$PORT" >&2
    exit 1
fi
