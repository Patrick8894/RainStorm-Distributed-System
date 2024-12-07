cd src
# <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>
go run client.go app1_op1.go app1_op2.go Traffic_Sign1000.txt business_3.txt 1 1 stateful
go run client.go app1_op1.go app1_op2.go Traffic_Sign1000.txt business_3.txt 1 1 stateful