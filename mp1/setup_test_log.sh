#!/bin/bash

# List of remote hosts

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

# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp1"

# Command to run on each host
for host in "${hosts[@]}"; do
  echo "Connecting to $host"
  
  # SSH into each host, check for existing process, kill it if found, and run new commands
  ssh "$host" "
    cd $repo_dir
    python3 generator.py 300000
  "
  
  echo "Done with $host"
done

