#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

sleep 2

shosetRun C 0 localhost:8003 localhost:8002 1 A 0 &
#P3=$!

wait