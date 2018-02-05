@echo off
go build -ldflags "-s -w" -tags prod ./cmd/gjpush
