#!/bin/bash

export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4317
export OTEL_EXPORTER=grpc
export ENV=systest

set -xe
go build  -o tests/fsn-watch ./cmd/fsnwatch
./tests/fsn-watch hosted
