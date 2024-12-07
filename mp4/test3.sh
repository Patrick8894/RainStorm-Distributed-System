# Restart the kill VMs
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
echo "Execute ssh $host and restart worker 