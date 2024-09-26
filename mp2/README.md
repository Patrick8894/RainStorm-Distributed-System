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