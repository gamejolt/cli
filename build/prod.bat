@echo off
go build -ldflags "-s -w" -tags "prod forceposix" ./cmd/gjpush
