#!/bin/sh

#alias shosetRun='go run -race test/*.go 5'

sleep 6

./bin/shoset_build 5 D 0 0 rien &
P=$!

#Kill and restart
sleep 11

kill $P

sleep 1

./bin/shoset_build 6 D 0 0 rien &

wait