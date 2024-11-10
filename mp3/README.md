# MP3


## Command
first enter to ``mp3/src`` folder
```
cd cs425g66/mp3/src
```
1. create localfilename HyDFSfilename
```
go run client.go create --localfilename ~/business/business_1.txt --HyDFSfilename business_1.txt
```
2. fetch HyDFSfile to local
```
go run client.go get --localfilename ~/a.txt --HyDFSfilename foo1.txt
```
3. check the HyDFS file location
```
go run client.go ls -HyDFSfilename business_3.txt
```
4. print the all membership's node id and membership
```
go run client.go list_mem_ids
```
5. list the HyDFS file store in local machine
```
go run client.go store
```
6. append local file content to the end of HyDFSfile
```
go run client.go append --localfilename foo1.txt --HyDFSfilename foo1.txt
```
7. get a file from particular replica indicated by VMaddress
```
go run client.go getfromreplica --VMaddress fa24-cs425-6610.cs.illinois.edu --HyDFSfilename foo1.txt --localfilename ~/out.txt
```
8. multiappend script, make machine 7 append businese_2.txt to the end of append.txt and so on
```
./multiappend append.txt 7 8 9 10 business_2.txt business_2.txt business_2.txt business_2.txt
```
9. merge all the replica files 
```
go run client.go merge --HyDFSfilename foo1.txt
```