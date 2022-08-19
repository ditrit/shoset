#!/bin/sh

#binary testNumber Lname receiver sender destination relaunch

./bin/shoset_build 5 A 1 0 rien 0 &
P=$!

#Kill and restart
sleep 35

kill $P

sleep 1

./bin/shoset_build 5 A 1 0 rien 1 &

wait
