#!/bin/sh

rm -rf ~/.shoset

alias shosetRun='timeout 30s go run -race test/*.go 5'

wait