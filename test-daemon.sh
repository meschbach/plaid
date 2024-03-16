#!/bin/bash

if [ -f "$PWD/test.env.sh" ]; then
  source "$PWD/test.env.sh"
  export ENV=systest
fi

set -e
echo
echo "Testing"
echo
export TEST_TIMEOUT="5s"
unit_test_flags="-count 1 -tags sane"
go test -timeout 10s $unit_test_flags ./resources/...
go test -timeout 10s $unit_test_flags ./internal/plaid/...
go test -timeout 10s $unit_test_flags ./ipc/... ./client/... ./service/... ./controllers/...

echo
echo "Building binary artifacts"
echo
#build_flags="-race"
build_flags=""
go build $build_flags -o tests/system/plaid-daemon ./cmd/daemon
go build $build_flags -o tests/system/plaid-client ./cmd/client
go build $build_flags -o tests/fsn-watch ./cmd/fsnwatch

daemon=$PWD/tests/system/plaid-daemon
client=$PWD/tests/system/plaid-client
client_flags="--delete-on-completion"

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
  cd simple && $client up $client_flags && echo "[test-harness] simple done"
)

# One shot test
echo
echo "Running one-shot tests"
echo
(cd deps/one-shot
  export OTEL_SERVICE_NAME="plaid_one-shot"
  export PLAID_CONFIG=$PWD/plaid.json
  (cd job-b && $client up  $client_flags && echo "[test-harness] one-shot dep test complete")
)

echo
echo "Running services"
echo
# Services
(cd deps/services
  rm -f service-a/service client-b/service
  export OTEL_SERVICE_NAME="plaid_services"
  export PLAID_CONFIG=$PWD/plaid.json
  (cd client-b &&  $client up  $client_flags)
)

echo
echo "Validated and ready"
echo
