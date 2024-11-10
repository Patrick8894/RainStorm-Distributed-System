cd src && go run client.go create --localfilename ~/business/business_1.txt --HyDFSfilename business_1.txt && \
cd src && go run client.go create --localfilename ~/business/business_2.txt --HyDFSfilename business_2.txt && \
cd src && go run client.go create --localfilename ~/business/business_3.txt --HyDFSfilename business_3.txt && \
cd src && go run client.go create --localfilename ~/business/business_4.txt --HyDFSfilename business_4.txt && \
cd src && go run client.go create --localfilename ~/business/business_5.txt --HyDFSfilename business_5.txt

sleep 5
go run src/client.go get --localfilename ~/output.txt --HyDFSfilename business_1.txt
echo "Content of business_1.txt\n"
more ~/output.txt
echo "Comparing 1.txt with business_1.txt\n"
diff ~/business/business_1.txt ~/output.txt