#!/bin/sh

#Open the call graph svg in a browser and generate new graph when you click on a package.
go-callvis -graphviz -nostd -nointer --group pkg,type -limit github.com/ditrit/shoset -focus github.com/ditrit/shoset ./../test #-minlen 50 -nodesep 5

#go-callvis -nostd -nointer --group pkg,type -minlen 5 ./../test