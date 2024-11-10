ssh fa24-cs425-6605.cs.illinois.edu << 'EOF'
go run client.go ls -HyDFSfilename business_3.txt && \
go run client.go list_mem_ids && \
cd ~/cs425g66/mp3/src && \
go run client.go store
EOF

