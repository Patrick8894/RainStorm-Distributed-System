#!/bin/bash

# List of remote hosts

hosts=(
  "bohaowu2@fa24-cs425-6605.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6601.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6602.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6603.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6604.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6606.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6607.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6608.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6609.cs.illinois.edu"
  "bohaowu2@fa24-cs425-6610.cs.illinois.edu"
)

# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp2/src"

# Command to run on each host
for i in "${!hosts[@]}"; do
    host="${hosts[$i]}"
    echo "Connecting to $host"
    
    # SSH into each host, check for existing process, kill it if found, and run new commands
    ssh "$host" "
    PORT=8081
    PID=\$(lsof -t -i :\${PORT})
    
    # If a PID is found, kill the process
    if [ ! -z \"\$PID\" ]; then
        echo \"Stopping old server with PID \$PID...\"
        kill \$PID
        # If needed, force kill
        # sudo kill -9 \$PID
    else
        echo \"No process found on port \${PORT}.\"
    fi
    
    cd $repo_dir
    go build node.go
    if [ $i -eq 0 ]; then
        nohup ./node --introducer > ../data/mp2${i}.log 2>&1 &
    else
        nohup ./node > ../data/mp2${i}.log 2>&1 &
    fi
  "
  echo "Done with $host"
done
