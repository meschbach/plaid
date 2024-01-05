#!/bin/bash

set -e
echo
echo "Testing"
echo
export ENV=systest
unit_test_flags="-count 1 -tags sane"
go test -timeout 10s $unit_test_flags ./resources/...
go test -timeout 10s $unit_test_flags ./internal/plaid/...
go test -timeout 10s $unit_test_flags ./ipc/... ./client/... ./service/...

echo
echo "Building binary artifacts"
echo
#build_flags="-race"
build_flags=""
go build $build_flags -o tests/system/plaid-daemon ./cmd/daemon
go build $build_flags -o tests/system/plaid-client ./cmd/client
go build $build_flags -o tests/system/deps/services/service-a/service ./tests/system/deps/services/service-a/cmd
go build $build_flags -o tests/system/deps/services/service-b/service ./tests/system/deps/services/service-b/cmd

daemon=$PWD/tests/system/plaid-daemon
client=$PWD/tests/system/plaid-client

cd tests/system

echo
echo "Spawning Daemon"
echo
export PLAID_SOCKET=$PWD/plaid.socket
$daemon run &
daemon_pid=$!
function cleanup {
  echo "[test-harness] Cleanup requested, shutting down daemon"
  kill $daemon_pid
  echo "[test-harness] Daemon asked to terminate"
}

trap cleanup EXIT

# ensure we can run a single command
echo
echo "Simple"
echo
(
  export OTEL_SERVICE_NAME="plaid_simple"
  cd simple && $client up  && echo "[test-harness] simple done"
)

# One shot test
echo
echo "Running one-shot tests"
echo
(cd deps/one-shot
  export OTEL_SERVICE_NAME="plaid_one-shot"
  export PLAID_CONFIG=$PWD/plaid.json
  (cd job-b && $client up && echo "[test-harness] one-shot dep test complete")
)

echo
echo "Running services"
echo
# Services
(cd deps/services
  export OTEL_SERVICE_NAME="plaid_services"
  export PLAID_CONFIG=$PWD/plaid.json
  (cd service-b &&  $client up )
)

echo
echo "Validated and ready"
echo
