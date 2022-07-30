#!/bin/sh

# To anable attachin to process  : /proc/sys/kernel/yama/ptrace_scope to 1 (0?)

#-gcflags=all="-N -l" Disable optimizations for debugging

go build -v -o shoset_build -race -gcflags=all="-N -l" test/*.go

rm -rf ~/.shoset

#alias shosetRun='timeout 30s go run -race test/*.go 5'

wait