#!/bin/bash

# openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt

REPERTORY=~/.shoset
LOGFILE=$REPERTORY/log.txt
i=0

for SHOSET in $REPERTORY/*;
do 
    PORT=`expr 8080 + $i`
    echo $SHOSET
    cd $SHOSET/cert;
    # echo $CERT
    for FILE in *;
    do 
        echo $FILE
    done
    echo -n "$SHOSET : " >> $LOGFILE
    openssl s_server -accept $PORT -www -cert cert.crt -key privateKey.key -CAfile CAcert.crt -naccept 1 | grep ACCEPT >> $LOGFILE & 
    sleep 2
    openssl s_client -connect localhost:$PORT -showcerts -CAfile CAcert.crt <<< "Q"
    echo ""
    i=`expr $i + 1`
done