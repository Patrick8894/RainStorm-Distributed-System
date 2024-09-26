#!/bin/bash

apt install -y protobuf-compiler
protoc --version
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Define the line to add
line_to_add='export PATH=$PATH:$(go env GOPATH)/bin'

# Check if the line already exists in ~/.bashrc
if ! grep -Fxq "$line_to_add" ~/.bashrc; then
    # If the line is not found, add it to ~/.bashrc
    echo "$line_to_add" >> ~/.bashrc
    echo "Line added to ~/.bashrc"
else
    echo "Line already exists in ~/.bashrc"
fi

go get google.golang.org/protobuf/proto
