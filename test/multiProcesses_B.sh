#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

#sleep 1

shosetRun B 0 localhost:8002 localhost:8001 0 rien 0 &
#P2=$!

wait