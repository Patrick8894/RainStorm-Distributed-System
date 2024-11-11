rm -rf ~/test
mkdir ~/test
cd src
# random uniform distribution
# for i in {1..100}
# do
#     j=$((RANDOM % 100 + 1))
#     go run client.go get --localfilename ~/test/output$j.txt --HyDFSfilename business_$j.txt
#     # wait until the file is downloaded
#     while [ ! -f ~/test/output$j.txt ]; do
#         sleep 0.1
#     done
# done

# frequent
# for i in {1..100}
# do
#     go run client.go get --localfilename ~/test/business_$i.txt --HyDFSfilename business_1.txt
#     while [ ! -f ~/test/business_$i.txt ]; do
#         sleep 0.1
#     done
# done


# mixed operation
# for i in {1..100}
# do
#     # j chooses the operation
#     j=$((RANDOM % 100 + 1))
#     # k chooses the file
#     k=$((RANDOM % 100 + 1))
#     if [ $j -lt 90 ];
#     then
#         go run client.go get --localfilename ~/test/business_$k.txt --HyDFSfilename business_$k.txt
#         while [ ! -f ~/test/business_$k.txt ]; do
#             sleep 0.1
#         done
#     else
#         go run client.go append --localfilename ~/business/business_$k.txt --HyDFSfilename business_$k.txt
#     fi
# done
