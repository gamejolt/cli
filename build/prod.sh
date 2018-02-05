#!/bin/bash
go build -ldflags "-s -w" -tags prod ./cmd/gjpush
