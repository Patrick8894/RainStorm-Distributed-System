cd src/ 
go run mp3_client.go create --localfilename ~/TrafficSigns_50.txt --HyDFSfilename TrafficSigns_50.txt && \
go run client.go app1_op1 app1_op2 TrafficSigns_50.txt 50.txt 3 "No Outlet" stateless &
cd ..

# wait to the middle of the steam processing
sleep 1.5
# fail two applications
# applicatoins use port 8091
host="bohaowu2@fa24-cs425-6604.cs.illinois.edu"
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

host="bohaowu2@fa24-cs425-6604.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8081
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping mp2 with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No mp2 found on port ${PORT}."
fi
EOF
echo "Execute ssh $host and kill mp2 with port" 


host="bohaowu2@fa24-cs425-6607.cs.illinois.edu"
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


host="bohaowu2@fa24-cs425-6607.cs.illinois.edu"
ssh "$host" << 'EOF'
PORT=8081
PID=$(lsof -t -i :${PORT})
# If a PID is found, kill the process
if [ ! -z "$PID" ]; then
    echo "Stopping mp2 with PID $PID..."
    kill $PID
    # If needed, force kill
    # sudo kill -9 $PID
else
    echo "No mp2 found on port ${PORT}."
fi
EOF
echo "Execute ssh $host and kill mp2 with port" 

# Print to console any duplicate records that were discarded by a task. 
# Print the tasks that were rescheduled to the console.