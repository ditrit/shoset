#!/bin/sh

#alias shosetRun='go run -race test/*.go 5'

sleep 4

./bin/shoset_build 5 C 0 0 rien &
#P3=$!

wait