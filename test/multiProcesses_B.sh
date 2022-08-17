#!/bin/sh

sleep 2

#binary testNumber Lname receiver sender destination relaunch

./bin/shoset_build 5 B 0 0 rien 0 &
P=$!

#Kill and restart
sleep 20

kill $P

sleep 1

./bin/shoset_build 5 B 0 0 rien 1 &

wait