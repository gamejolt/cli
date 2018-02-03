#!/bin/bash
go build -ldflags "-s -w" -tags prod -o gjpush ./cmd/cli
