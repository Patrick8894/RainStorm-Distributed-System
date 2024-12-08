cd ../mp3/src
go run ../../mp3/src/client.go create --localfilename ~/Traffic_Signs_1000.txt --HyDFSfilename Traffic_Signs_1000.txt && \
cd ../../mp4/src
# <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>
go run client.go app1_op1.go app1_op2.go Traffic_Signs_1000.txt test1_1.txt 3 "No Outlet" stateless && \
# go run client.go app2_op1.go app2_op2.go Traffic_Signs_1000.txt test2_2.txt 3 "Punched Telespar" stateful
echo "All done"