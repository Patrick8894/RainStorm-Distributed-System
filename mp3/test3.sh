# business 2 store in 4
ssh fa24-cs425-6605.cs.illinois.edu (
cd ~/cs425g66/mp2/src
go run src/control.go --c kill --s 4 && go run src/control.go --c kill --s 6 
sleep 3
cd ~/cs425g66/mp3/src
go run client.go ls -HyDFSfilename business_2.txt
go run client.go store
go run client.go list_mem_ids )
