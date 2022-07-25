#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

shosetRun A 1 localhost:8001 rien 0 rien 1 &
#P1=$!

wait
