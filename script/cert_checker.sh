#!/bin/bash

# openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt

# macros
REPERTORY=~/.shoset
LOGFILE=$REPERTORY/cert_checker.txt

i=0

for SHOSET in $REPERTORY/*;
do 
    PORT=`expr 8080 + $i`
    echo $SHOSET
    cd $SHOSET/cert;
    for FILE in *;
    do 
        echo $FILE
    done
    echo -n "$SHOSET : " >> $LOGFILE

    # run openssl command server and check if it worked
    openssl s_server -accept $PORT -www -cert cert.crt -key privateKey.key -CAfile CAcert.crt -naccept 1 | grep ACCEPT >> $LOGFILE & 
    
    sleep 0.5
    
    # run openssl command client and quit when connected
    # https://stackoverflow.com/questions/25760596/how-to-terminate-openssl-s-client-after-connection
    openssl s_client -connect localhost:$PORT -showcerts -CAfile CAcert.crt <<< "Q"

    echo ""
    i=`expr $i + 1`
done