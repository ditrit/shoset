#!/bin/sh

sleep 8

#binary testNumber Lname receiver sender destination relaunch

./bin/shoset_build 5 E 0 1 A 0 &
P=$!

#Kill and restart
sleep 35

kill $P

sleep 1

./bin/shoset_build 5 E 0 1 A 1 &

wait