#!/bin/bash
go build -ldflags "-s -w" -tags "forceposix" ./cmd/gjpush
