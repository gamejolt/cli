#!/bin/bash
go build -ldflags "-s -w" -tags "prod forceposix" ./cmd/gjpush
