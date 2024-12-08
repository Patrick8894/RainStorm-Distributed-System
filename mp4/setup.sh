#!/bin/bash

# List of remote hosts

hosts=(
    "bohaowu2@fa24-cs425-6605.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6601.cs.illinois.edu"
    "bohaowu2@fa24-cs425-6602.cs.illinois.edu"
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
    "machine.4.log"
    "machine.6.log"
    "machine.7.log"
    "machine.8.log"
    "machine.9.log"
    "machine.10.log"
)

# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp4/src"

# Command to run on each leader
host="${hosts[0]}"
echo "Connecting to $host"
ssh "$host" "
    PORT=8090
    PID=\$(lsof -t -i :\${PORT})

    
    # If a PID is found, kill the process
    if [ ! -z \"\$PID\" ]; then
        echo \"Stopping old server with PID \$PID...\"
        kill \$PID
        #
        # sudo kill -9 \$PID
    else
        echo \"No process found on port \${PORT}.\"
    fi

    cd $repo_dir

    go build leader.go
    nohup ./leader > ../../mp1/data/${log_files[0]} 2>&1 &

    echo \"Done with leader $host\"
"

# Command to run  worker on each host
for i in "${!hosts[@]}"; do
    host="${hosts[$i]}"
    echo "Connecting to $host"
    
    # SSH into each host, check for existing process, kill it if found, and run new commands
    ssh "$host" "
    PORT=8091
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

    go build worker.go
    nohup ./worker > ../../mp1/data/${log_files[$i]} 2>&1 &
  "
  echo "Done with worker $host"
done