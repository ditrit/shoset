#!/bin/sh

# To anable attaching to process  : /proc/sys/kernel/yama/ptrace_scope to 0
#code -n /proc/sys/kernel/yama/ptrace_scope

#-gcflags=all="-N -l" Disable optimizations for debugging

go build -v -o ./bin/shoset_build -race -gcflags=all="-N -l" ./test/test.go

rm -rf ~/.shoset

rm -rf ./profiler/*

wait