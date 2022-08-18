#!/bin/bash

#!!! NOT WORKING !!#
#Problems with connection refused and Adresses already in use.

# macros
REPERTORY=~/.shoset
LOGFILE=$REPERTORY/cert_checker.txt

i=0

for SHOSET in $REPERTORY/*; do
    for CONN in $SHOSET/*; do
        if [ -d "${CONN}" ]; then
            echo "${CONN}"

            # Different ports for server et client ?
            # Connection refused ???
            PORT=$(expr 8080 + $i)
            echo
            echo
            cd $CONN/cert
            echo "#### Folder currently checked : $PWD"

            #echo -n "$SHOSET : " >> $LOGFILE ??

            # run openssl command server and check if it worked
            openssl s_server -accept $PORT -www -cert ./cert.crt -key ./privateKey.key -CAfile ./CAcertificate.crt -naccept 1 | grep ACCEPT & #>>$LOGFILE

            # sleep 0.5

            # run openssl command client and quit when connected
            # https://stackoverflow.com/questions/25760596/how-to-terminate-openssl-s-client-after-connection
            openssl s_client -connect localhost:$PORT -showcerts -CAfile ./CAcertificate.crt <<<"Q"

            # sleep 0.5

            # Checks hashes of the keys.
            openssl pkey -in ./privateKey.key -pubout -outform pem | sha256sum
            openssl x509 -in ./cert.crt -pubkey -noout -outform pem | sha256sum
            openssl verify -verbose -CAfile ./CAcertificate.crt ./cert.crt

            echo DONE
            i=$(expr $i + 1)
        fi
    done
done
