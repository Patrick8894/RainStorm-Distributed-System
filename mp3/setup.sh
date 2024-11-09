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

log_files=(
    "machine.5.log"
    "machine.1.log"
    "machine.2.log"
    "machine.3.log"
    "machine.4.log"
    "machine.6.log"
    "machine.7.log"
    "machine.8.log"
    "machine.9.log"
    "machine.10.log"
)

# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp3/src"

# Command to run on each host
for i in "${!hosts[@]}"; do
    host="${hosts[$i]}"
    echo "Connecting to $host"
    
    # SSH into each host, check for existing process, kill it if found, and run new commands
    ssh "$host" "
    PORT=8085
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

    # Delete the mp3/data before starting the new server
    rm -rf mp3/data

    # Create the mp3/data directory
    mkdir -p mp3/data

    cd $repo_dir
    go build client.go
    go build server.go
    nohup ./server > ../../mp1/data/${log_files[$i]} 2>&1 &
  "
  echo "Done with $host"
done