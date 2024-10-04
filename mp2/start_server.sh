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

log_files=(
    "machine.1.log"
    "machine.2.log"
    "machine.3.log"
    "machine.4.log"
    "machine.5.log"
    "machine.6.log"
    "machine.7.log"
    "machine.8.log"
    "machine.9.log"
    "machine.10.log"
)
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

log_files=(
    "machine.1.log"
    "machine.2.log"
    "machine.3.log"
    "machine.4.log"
    "machine.5.log"
    "machine.6.log"
    "machine.7.log"
    "machine.8.log"
    "machine.9.log"
    "machine.10.log"
)

# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp2/src"

# Check if an argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <server_number>"
    exit 1
fi

# Get the server number from the argument
server_number=$1

# Validate the server number
if [ "$server_number" -lt 1 ] || [ "$server_number" -gt ${#hosts[@]} ]; then
    echo "Invalid server number. Please provide a number between 1 and ${#hosts[@]}."
    exit 1
fi

# Adjust index to be zero-based
index=$((server_number - 1))

# Get the corresponding host and log file
host="${hosts[$index]}"
log_file="${log_files[$index]}"

echo "Connecting to $host"

# SSH into the host, check for existing process, kill it if found, and run new commands
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
go build control.go
if [ $index -eq 4 ]; then
    nohup ./node --introducer > ../../mp1/data/$log_file 2>&1 &
else
    nohup ./node > ../../mp1/data/$log_file 2>&1 &
fi
"

echo "Done with $host"
