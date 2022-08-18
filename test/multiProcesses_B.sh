#!/bin/sh

#binary testNumber Lname receiver sender destination relaunch

sleep 1

./bin/shoset_build 5 B 0 0 rien 0 &
P=$!

#Kill and restart
sleep 30

kill $P

sleep 1

./bin/shoset_build 5 B 0 0 rien 1 &

wait