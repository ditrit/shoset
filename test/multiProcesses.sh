#!/bin/sh

#connection refused
#/home/mint20/Documents/DitRit/shoset/shoset.go:303

rm -rf ~/.shoset

alias shosetRun='timeout 30s go run -race test/*.go 5'

shosetRun A 1 localhost:8001 rien 0 rien 1 &
P1=$!

#sleep 1

shosetRun B 0 localhost:8002 localhost:8001 0 rien 0 &
P2=$!

#sleep 1
# kill P2

shosetRun C 0 localhost:8003 localhost:8002 1 A 0 &
P3=$!

# #sleep 1

# shosetRun D 0 localhost:8004 localhost:8003 1 A 0 &
# P4=$!

wait