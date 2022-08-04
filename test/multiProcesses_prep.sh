#!/bin/sh

# To anable attachin to process  : /proc/sys/kernel/yama/ptrace_scope to 0
#code /proc/sys/kernel/yama/ptrace_scope

#go tool pprof -http=":" ./shoset_build "./profiler_save/cpu_C_normal.prof"

#-gcflags=all="-N -l" Disable optimizations for debugging

go build -v -o shoset_build -race -gcflags=all="-N -l" test/*.go

rm -rf ~/.shoset

rm -rf ./profiler/*

#alias shosetRun='timeout 30s go run -race test/*.go 5'

wait