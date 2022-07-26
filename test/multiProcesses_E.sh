#!/bin/sh

alias shosetRun='go run -race test/*.go 5'

sleep 4

# lname pki IP remote IP sender destination receiver

shosetRun D 0 localhost:8004 localhost:8003 1 A 0 &
#P4=$!

wait