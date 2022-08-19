#!/bin/bash

# !!! NOT WORKING !!!
# Need updating and test.

# macros
REPERTORY=~/.shoset
LOGFILE=$REPERTORY/shoset_checker.txt
EXPECTED_FILES_NUMBER=$1

number_of_errors=0

# allow ctrl+C while running
set -m

# clean files
echo "" > log.txt
echo "" > log_error.txt

# test program n times
for i in {1..10}
do
    # run command
    rm -rf $REPERTORY && timeout 45s go run -race test/test.go 4 >> log.txt 2>&1 && grep ERROR log.txt >> log_error.txt

    # count number of files in the repertory
    number_of_files=`find $REPERTORY -type f | wc -l`

    if [ $EXPECTED_FILES_NUMBER -eq $number_of_files ]
        then
            number_of_errors=`expr $number_of_errors + 0`
        else
            echo "expected $EXPECTED_FILES_NUMBER, got $number_of_files" 
            number_of_errors=`expr $number_of_errors + 1`
    fi
    echo "errors : $number_of_errors"
    echo "loop number : $i"
    sleep 5
done

echo $number_of_errors
