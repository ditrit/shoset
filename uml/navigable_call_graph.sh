#!/bin/sh

#Open the call graph svg in a browser and generate new graph when you click on a package.
go-callvis -nostd -nointer --group pkg,type -minlen 5 -limit github.com/ditrit/shoset -focus github.com/ditrit/shoset ./../test