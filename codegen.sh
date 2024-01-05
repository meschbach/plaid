#!/bin/bash

set -e
function gen() {
  protoc --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      $1
}

gen ./internal/plaid/daemon/wire/resources.proto
gen ./ipc/grpc/logger/logger.proto
