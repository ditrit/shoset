#!/bin/bash

# openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt

# macros
REPERTORY=~/.shoset
LOGFILE=$REPERTORY/cert_checker.txt

i=0

for SHOSET in $REPERTORY/*; do
    #PORT=$(expr 8080 + $i)
    #echo $SHOSET
    for CONN in $SHOSET/*; do
        if [ -d "${CONN}" ]; then
            echo "${CONN}"

            PORT=$(expr 8080 + $i)
            cd $CONN/cert
            echo "$PWD"
            ls
            #echo -n "$SHOSET : " >> $LOGFILE ??

            echo $i
            echo $PORT

            # run openssl command server and check if it worked
            echo "### server :"
            openssl s_server -accept $PORT -www -cert ./cert.crt -key ./privateKey.key -CAfile ./CAcertificate.crt -naccept 1 | grep ACCEPT >>$LOGFILE &

            # sleep 0.5

            # run openssl command client and quit when connected
            # https://stackoverflow.com/questions/25760596/how-to-terminate-openssl-s-client-after-connection
            echo "### client :"
            openssl s_client -connect localhost:$PORT -showcerts -CAfile ./CAcertificate.crt <<<"Q"

            # sleep 0.5

            # openssl pkey -in ./privateKey.key -pubout -outform pem | sha256sum
            # openssl x509 -in ./cert.crt -pubkey -noout -outform pem | sha256sum
            # openssl verify -verbose -CAfile ./CAcertificate.crt ./cert.crt

            #openssl req -in CSR.csr -pubkey -noout -outform pem | sha256sum

            echo DONE
            i=$(expr $i + 1)
        fi
    done
done
