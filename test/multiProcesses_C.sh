#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

#sleep 2

# lname pki IP remote IP sender destination receiver

shosetRun C 0 localhost:8003 localhost:8002 0 rien 0 &
#P3=$!

wait