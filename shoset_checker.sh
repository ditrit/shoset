#!/bin/bash

REPERTORY=~/.shoset
LOGFILE=$REPERTORY/shoset_checker.txt
EXPECTED_FILES_NUMBER=$1

number_of_errors=0

for i in {1..10}
do
    timeout 60s go run ./test/test.go 4
    number_of_files=`find $REPERTORY -type f | wc -l`
    if [ $EXPECTED_FILES_NUMBER -eq $number_of_files ]
        then
            number_of_errors=`expr $number_of_errors + 0`
        else
            number_of_errors=`expr $number_of_errors + 1`
    fi
    rm -rf $REPERTORY
    echo "errors : $number_of_errors"
    echo "loop number : $i"

done

echo $number_of_errors
