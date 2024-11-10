#!/bin/bash

hosts=(
    "no"
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

# Check if the number of arguments is at least 3
if [ "$#" -lt 3 ]; then
    echo "Usage: $0 filename VM1 [VM2 ... VMN] localfile1 [localfile2 ... localfileN]"
    exit 1
fi

filename=$1
shift

# Calculate the number of remaining arguments
arg_count=$#

# Ensure there are an even number of arguments (equal VMs and local filenames)
if (( arg_count % 2 != 0 )); then
    echo "Error: The number of VMs and local filenames must be equal."
    exit 1
fi

# Number of VMs/local filenames
N=$(( arg_count / 2 ))

# Split the arguments into arrays of VMs and local filenames
VMs=("${@:1:$N}")

localfiles=("${@:$((N+1)):$N}")

# Loop through each VM and local filename pair and run the client.go script in the background
for (( i=0; i<N; i++ )); do
    index="${VMs[$i]}"
    host="${hosts[$index]}"
    localfilename="${localfiles[$i]}"
    ssh "$host" "cd cs425g66/mp3/src && go run client.go append  --localfilename $localfilename  --HyDFSfilename $filename" &
    echo "ssh $host" "go run client.go append  --localfilename $localfilename  --HyDFSfilename $filename" &
done

# Wait for all background processes to complete
wait


