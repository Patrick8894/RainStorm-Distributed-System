# Directory where the repo is located on remote machines
repo_dir="cs425g66/mp4/src"

# Restart the kiled from VMs test2.sh
host="bohaowu2@fa24-cs425-6610.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8091
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping worker with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No worker found on port ${PORT}."
fi

cd $repo_dir

go build worker.go
nohup ./worker > ../../mp1/data/${log_files[0]} 2>&1 &
EOF
echo "Execute ssh $host and restart worker 



host="bohaowu2@fa24-cs425-6609.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8091
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping worker with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No worker found on port ${PORT}."
fi

cd $repo_dir

go build worker.go
nohup ./worker > ../../mp1/data/${log_files[0]} 2>&1 &
EOF
echo "Execute ssh $host and restart worker 


# Waiting for stablize
echo "Wait for 5 seconds for the system to stablize"
sleep 5

# Run applications 2
echo "Run application 2"
go run src/client.go app2_op1.go app2_op2.go Traffic_Signs_1000.txt test3_2.txt 3 "No Outlet" stateful

# wait to the middle of the steam processing
sleep 1.5

# fail two applications
# must kill the statful application task
host="bohaowu2@fa24-cs425-6610.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8091
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping worker with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No worker found on port ${PORT}."
fi
EOF
echo "Execute ssh $host and kill worker with port" 


host="bohaowu2@fa24-cs425-6609.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8091
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping worker with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No worker found on port ${PORT}."
fi
EOF
echo "Execute ssh $host and kill worker with port" 
