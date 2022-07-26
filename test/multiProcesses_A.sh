#!/bin/sh

go build -v -race -gcflags=all="-N -l" test/*.go

#alias shosetRun='go run -race -gcflags=all="-N -l" test/*.go 5'

./shoset_build 5 A 1 0 rien &
#P1=$!

wait
