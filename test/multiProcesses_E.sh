#!/bin/sh

#binary testNumber Lname receiver sender destination relaunch

sleep 4

./bin/shoset_build 5 E 0 1 A 0 &
P=$!

#Kill and restart
sleep 15

kill $P

sleep 1

./bin/shoset_build 5 E 0 1 A 1 &

wait