#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

#sleep 3

shosetRun D 0 localhost:8004 localhost:8003 1 A 0 &
#P4=$!

wait