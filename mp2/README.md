# MP2

## Compile the *.proto file
1. Install protoc ([reference](https://protobuf.dev/getting-started/gotutorial/))

```
# when using Linux environment
$ apt install -y protobuf-compiler
$ protoc --version  # Ensure compiler version is 3+

# when using Mac environment
$ brew install protobuf
$ protoc --version  # Ensure compiler version is 3+
```

2. Install plugin
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

3. Add PATH
You can put this command under `~/.bashrc` or  `~/.zshrc`
```
export PATH=$PATH:$(go env GOPATH)/bin
source ~/.bashrc (or source ~/.zshrc)
```

4. Get the protobuf package
```
go get google.golang.org/protobuf/proto
```

5. using Protocol Buffers compiler to generate go file that can be imported to main function
```
protoc --go_out=. --go_opt=paths=source_relative proto/swim.proto
```


## Start & Control services

1. To update the latest version
```
./pull_repo.sh
```
2. To start the services
```
# start all VM machine in global.go files
cd mp2/ && chmod +x setup.sh && ./setup.sh
# start specific VM machine
cd mp2/ && ./start_server 1 #open the first node
cd mp2/ && ./start_server 2 #open the second node
```
3. To send comand to running machine by src/control.go
a. list the specific node membership info
```
go run src/control.go --c ls --s 6 # list the sixth node membership info
```
b. list the specific node suspected membership (after kill the node, only 3.5 seconds can see)
```
go run src/control.go --c lss --s 6 # list the sixth node membership info
```
c. list the specific node gossiplist
```
go run src/control.go --c lsg --s 6 # list the sixth node gossip list
```
d. kill the specific node
```
go run src/control.go --c kill --s 6 # kill sixth node
```
e. update the drop rate
```
go run src/control.go --c drop --s 0.01 # update all nodes' drop rate
```