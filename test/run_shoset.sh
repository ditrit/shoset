# command  : ./test/run_shoset.sh 4

### To use aliases instead :
# code -n ~/.bash_aliases
# source ~/.bash_aliases

rm -rf ~/.shoset 
go run -race test/test.go $1
wait