#!/bin/sh

sleep 4

#binary testNumber Lname receiver sender destination relaunch

./bin/shoset_build 5 C 0 0 rien 0 &
P=$!

#Kill and restart
sleep 25

kill $P

sleep 1

./bin/shoset_build 5 C 0 0 rien 1 &

wait