#!/bin/sh

# Installation instructions :


# Class diagram :
# Install goplantuml :(Make sure /home/"user"/go/bin is include in your PATH)
# go install github.com/jfeliu007/goplantuml/cmd/goplantuml@latest

# Install plantuml :
# (Make sure java is installed)
# Download the jar file from : https://plantuml.com/en/download (Don't forget to change the path to it in the command.)


# Call diagram :
# Install go-callvis :
# git clone https://github.com/ofabry/go-callvis.git
# cd go-callvis && make install

# Install graphviz (Not sure if it's necessary) :
# sudo apt install graphviz

goplantuml -recursive -aggregate-private-members -show-aggregations -show-aliases -show-compositions -show-connection-labels -show-implementations -title "Shoset class diagram"  ./.. > class_diagram_shoset.puml
java -jar '/media/partagé/DitRit/GoPlantUML/plantuml-1.2022.6.jar' -svg -v ./class_diagram_shoset.puml

go-callvis -nostd -nointer --group pkg,type -minlen 5 -limit github.com/ditrit/shoset -focus github.com/ditrit/shoset -file ./call_graph/shoset ./../test

go-callvis -nostd -nointer --group pkg,type -minlen 5 -limit github.com/ditrit/shoset -focus github.com/ditrit/shoset/msg -file ./call_graph/msg ./../test

go-callvis -nostd -nointer --group pkg,type -minlen 5 -limit github.com/ditrit/shoset -focus github.com/ditrit/shoset/files -file ./call_graph/files ./../test #Not working