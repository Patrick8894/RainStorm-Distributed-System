go run src/client.go app1_op1.go app1_op2.go Traffic_Signs_1000.txt app1.txt 3 "No Outlet" stateless

# wait to the middle of the steam processing
sleep 1.5
# fail two applications
# applicatoins use port 8091
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


# Print to console any duplicate records that were discarded by a task. 
# Print the tasks that were rescheduled to the console.