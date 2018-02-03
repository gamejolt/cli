@echo off
go build -ldflags "-s -w" -tags prod -o gjpush ./cmd/cli
