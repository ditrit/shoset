#!/bin/sh

sleep 6

#binary testNumber Lname receiver sender destination relaunch

./bin/shoset_build 5 D 0 0 rien 0 &
P=$!

#Kill and restart
sleep 30

kill $P

sleep 1

./bin/shoset_build 5 D 0 0 rien 1 &

wait