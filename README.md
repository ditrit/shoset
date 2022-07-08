# shoset

A smart multi-ends socket library.

## Documentation

Missing. -- Find a usage example [in the gandalf project](https://github.com/ditrit/gandalf-core/blob/master/aggregator/aggregator.go).

## Design principles

For now, over to the [team's taiga](https://taiga.orness.com/project/xavier-namt/wiki/shoset)

## Running tests

You can run multiple tests in Shoset. There are three types of tests; unit test, script test or functional tests. 

### Unit tests
To run unit tests, you need to go in your editor (I did it on VSCode) in the file with a name corresponding to `*_test.go`, then you can simply click on `run test` or `debug test` button. There is also a dedicated tab in the navigation bar with a flask as icon to run each test one after the other.

### Script tests
To run script tests, you can find two scripts in the folder `shoset/script/` which are `cert_checker.sh` and `shoset_checker.sh`. These tests can be run in a Linux terminal as follows :

```txt
./script/cert_checker
./script/shoset_checker number_of_files_expected
```

#### cert_checker.sh
`cert_checker.sh` is a script done to check the validity of certificates generated for each Shoset after those have run a functional test. It will run a server in background and then run a client who will connect to this server by showing its certificates. If the certificates are valid, then you must see in the file `~/.shoset/cert_checker.txt` the word `ACCEPTED` for each Shoset.

#### shoset_checker.sh
`shoset_checker.sh` is a script done to launch multiple functional tests on after the other to check if there are no errors occurring when multiples test are run. You need to add the argument `number_of_files_expected` which is an integer representing the number of files you want to be obtained after each test. For example, if you have a network composed of 4 Shoset (1 PKI and 3 classic Shoset), then you will obtain 17 files (4 for the PKI, 3 for each classic Shoset and 1 config file for each Shoset - depending on your network complexity, there is not always a config file). To precisely know the number of file you need, you can simply run a functional test one time and see how many files you have by running this command :
```txt
tree ~/.shoset/
```
After the script is run, you can watch if errors occurred in the file `shoset/log_error.txt`, if this file is empty, then it means that you are good to go !

### Functional tests

You can run multiple functional tests in the file `shoset/test/test.go` with this command : 
```txt
go run test/test.go arg
```
The arg argument corresponds to the test you want to run. Take a look at the test file and adapt the test if you want. Don't forget to remove existing certificates and config files in `~/.shoset/folder`.

You can set an alias to run tests easily : `alias shoset='rm -rf ~/.shoset && timeout 30s go run -race test/test.go 4 > log.txt 2>&1'`. It removes existing files, timeouts the test at 30 seconds (it shouldn't last more than 15 seconds except if the network is large), runs the test with -race argument to detect data races and print the output in a log.txt file that you can find in the `shoset/` folder.

If you need to kill a Shoset at running time for testing. Then I advise you to launch two terminals. In the first one you will run your program with 2 as arg which will run the simpleCluster() function that simply creates a PKI Shoset. In the other terminal, run your program with 4 as arg which will run the testJoin4() function that creates multiple classic Shosets. Then you can `CTRL+C` in one of the two terminal to kill the Shoset(s). Of course, you need to adapt the call to the function corresponding to your arg in the main() function of the test file.

If you want to create your own test, don't forget that the network needs that the first Shoset is initialized as a PKI with InitPki() function and the other with the Protocol() protocol - most of the function are deprecated because they do not use InitPKI() neither Protocol(), please don't use them.
