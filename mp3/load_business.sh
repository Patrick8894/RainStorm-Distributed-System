cd src
for i in {1..100}
do
    go run client.go create --localfilename ~/business/business_${i}.txt --HyDFSfilename business_${i}.txt
    echo "go run client.go create --localfilename ~/business/business_${i}.txt --HyDFSfilename business_${i}.txt"
done