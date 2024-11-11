#!/bin/bash

# Check if the correct number of arguments is provided
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <iterations> <localfilename> <HyDFSfilename>"
    exit 1
fi

# Assign arguments to variables
iterations=$1
localfilename=$2
HyDFSfilename=$3

# Change directory to the source directory
cd ~/cs425g66/mp3/src

# Loop for the specified number of iterations
for ((i=0; i<iterations; i++))
do
    go run client.go append --localfilename "$localfilename" --HyDFSfilename "$HyDFSfilename"
done