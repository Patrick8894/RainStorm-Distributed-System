./multiappend.sh business_4.txt 7 8 9 10 ~/business/business_7.txt ~/business/business_8.txt ~/business/business_9.txt ~/business/business_10.txt
cd src
go run client.go merge --HyDFSfilename business_4.txt
go run client.go getfromreplica --VMaddress fa24-cs425-6605.cs.illinois.edu --HyDFSfilename business_4.txt --localfilename ~/test5.txt
cat ~/test5.txt
