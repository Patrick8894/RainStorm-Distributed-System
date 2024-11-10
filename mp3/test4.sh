# test in node 5

go run client.go append --localfilename ~/business/business10.txt --HyDFSfilename business3.txt && \
go run client.go append --localfilename ~/business/business11.txt --HyDFSfilename business3.txt


go run client.go get --localfilename ~/output.txt --HyDFSfilename business3.txt
echo "Content of business3.txt\n"
more ~/output.txt
