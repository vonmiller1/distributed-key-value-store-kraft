#!/bin/bash

case $1 in
    get)
        KEY=${2:?key not given}
        curl -G "localhost:3000/get" --data-urlencode "key=$KEY"
        echo
        ;;
    put)
        KEY=${2:?key not given}
        VALUE=${3:?value not given}
        curl -G "localhost:3000/put" --data-urlencode "key=$KEY" --data-urlencode "val=$VALUE"
        echo
        ;;
    delete)
        KEY=${2:?key not given}
        curl -G "localhost:3000/delete" --data-urlencode "key=$KEY"
        echo
        ;;
    status)
        curl "localhost:3000/status"
        echo
        ;;
    *)
        echo "Unknown usage: $1"
        echo "Valid commands:"
        echo "  get [key]"
        echo "  put [key] [value]"
        echo "  delete [key]"
        echo "  status"
        exit 1
        ;;
esac
