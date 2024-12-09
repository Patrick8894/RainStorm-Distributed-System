cd ../mp3/src
go run client.go create --localfilename ~/TrafficSigns_4.txt --HyDFSfilename TrafficSigns_4.txt
go run client.go create --localfilename ~/test1_1.txt --HyDFSfilename test1_1.txt
cd ../../mp4/src
# <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>
go run client.go app1_op1 app1_op2 TrafficSigns_10.txt h.txt 3 "No Outlet" stateless && 
go run client.go app2_op1 app2_op2 TrafficSigns_50.txt b.txt 3 "Punched Telespar" stateful 
echo "All done"