#!/bin/bash

hosts=(
    "bohaowu2@fa24-cs425-6601.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6602.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6603.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6604.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6605.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6606.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6607.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6608.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6609.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6610.cs.illinois.edu"
)

machine_count=$1
localfilename=$2
HyDFSfilename=$3

# Calculate the outcome of 1000/machine_count
iterations=$((1000 / machine_count))

# Loop through each VM and local filename pair and run the client.go script in the background
for (( i=0; i<machine_count; i++ )); do
    host="${hosts[$i]}"
    ssh "$host" "cd cs425g66/mp3 && ./many_append.sh $iterations $localfilename $HyDFSfilename" &
    echo "ssh $host" "go run client.go append  --localfilename $localfilename  --HyDFSfilename $HyDFSfilename" &
done

# Wait for all background processes to complete
wait


