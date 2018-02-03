@echo off
go build -ldflags "-s -w" -o gjpush ./cmd/cli
