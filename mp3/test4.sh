# test in node 5
cd src
go run client.go append --localfilename ~/business/business_10.txt --HyDFSfilename business_3.txt && \
go run client.go append --localfilename ~/business/business_11.txt --HyDFSfilename business_3.txt


go run client.go get --localfilename ~/output.txt --HyDFSfilename business_3.txt
echo "Content of business3.txt\n"
more ~/output.txt
