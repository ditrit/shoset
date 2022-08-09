#!/bin/sh

#alias shosetRun='go run -race test/*.go 5'

sleep 8

./bin/shoset_build 5 E 0 1 A &
#P4=$!

wait