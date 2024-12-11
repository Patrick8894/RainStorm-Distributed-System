# MP4


## Spark install
1. Install Spark file  from ![Apache](https://spark.apache.org/downloads.html)
```
tar xvf spark-<version>-bin-hadoop<version>.tgz
```
2. move the VM
```
scp -r spark-<version>-bin-hadoop<version> user@location
```
3. move to disired directory
```
mv spark-<version>-bin-hadoop<version> ~/
```
3. Add env variable, can use ```which java```  or ```readlink -f /usr/bin/java```to see java location
```
vim ~/.bashrc
# then add the following content into the file
export SPARK_HOME=~/spark-3.5.3-bin-hadoop3  # Replace with your Spark installation directory
export PATH=$SPARK_HOME/bin:$SPARK_HOME/sbin:$PATH
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-17.0.13.0.11-4.el9.x86_64
export PATH=$JAVA_HOME/bin:$PATH
```

4. activate
```
source ~/.bashrc
```

## Spark setting cluster

1. 
```
vim $SPARK_HOME/conf/spark-env.sh
```

2. The setting for spark-env.sh 
```
# The IP address or hostname of the master
SPARK_MASTER_HOST=172.22.156.220

# The IP address the worker should bind to (use the same as the master in this case)
SPARK_LOCAL_IP=172.22.156.220

# The URL of the master (used for worker registration)
SPARK_MASTER_URL=spark://172.22.156.220:7077

# Total CPU cores to allocate for the worker on this VM
SPARK_WORKER_CORES=4

# Total memory to allocate for the worker on this VM
SPARK_WORKER_MEMORY=8g

# Directory where worker logs are stored
SPARK_WORKER_DIR=/tmp/spark-worker

# Port configurations (optional, Spark picks defaults if not set)
SPARK_MASTER_PORT=7077
SPARK_MASTER_WEBUI_PORT=8080
SPARK_WORKER_WEBUI_PORT=8081
```


example setting
```
# The IP address or hostname of the master
SPARK_MASTER_HOST=172.22.156.220

# The IP address the worker should bind to (use the same as the master in this case)
SPARK_LOCAL_IP=172.22.156.220

# The URL of the master (used for worker registration)
SPARK_MASTER_URL=spark://172.22.156.220:7077

# Total CPU cores to allocate for the worker on this VM
SPARK_WORKER_CORES=4

# Total memory to allocate for the worker on this VM
SPARK_WORKER_MEMORY=8g

# Directory where worker logs are stored
SPARK_WORKER_DIR=/tmp/spark-worker

# Port configurations (optional, Spark picks defaults if not set)
SPARK_MASTER_PORT=7077
SPARK_MASTER_WEBUI_PORT=8080
SPARK_WORKER_WEBUI_PORT=8081
```

3. To start and stop the spark service
master 
```
$SPARK_HOME/sbin/start-master.sh
$SPARK_HOME/sbin/stop-master.sh
```
worker
```
$SPARK_HOME/sbin/start-worker.sh spark://fa24-cs425-6605.cs.illinois.edu:7077
$SPARK_HOME/sbin/stop-worker.sh
```

## Usage of Spark
1. submit job in master node or worker node
```
# master node
spark-submit \
  --master spark://fa24-cs425-6605.cs.illinois.edu:7077 \
  --deploy-mode cluster \
  traffic_signs_test1.py

# worker node
spark-submit \
  --master spark://fa24-cs425-6605.cs.illinois.edu:7077 \
  --deploy-mode client \
  traffic_signs_test1.py


spark-submit \
  --master spark://fa24-cs425-6605.cs.illinois.edu:7077 \
    --conf spark.driver.port=7078 \
  --conf spark.blockManager.port=7079 \
  --deploy-mode client \
  traffic_signs_test1.py
```

simulate the failure
```
sleep 1.5
ssh user@<worker_ip1> "pkill -f spark"
ssh user@<worker_ip2> "pkill -f spark"
```

## Command
1. Before start the stream processing function, restart the mp2 & mp3 service for knowing the membership and HyDFS system.
```
cd ~/cs425g66/mp4
./restart.sh
```

2. Before running the stream porcessing function, ensure creating the target file on HyDFS system
```
go run mp3_client.go create --localfilename ~/TrafficSigns_1000.txt --HyDFSfilename TrafficSigns_1000.txt
```

3. Two type of application that can run
```
go run client.go app1_op1 app1_op2 TrafficSigns_1000.txt h.txt 3 "Parking" stateless && 
go run client.go app2_op1 app2_op2 TrafficSigns_1000.txt b.txt 3 "Unpunched Telespar" stateful 

# to test the result, get the file from HyDFS
go run mp3_client.go get --localfilename ~/t3.txt --HyDFSfilename b.txt
vim ~/t3.txt
```

4. If want new type of operation, update the ops/app*.go and rebuild it
```
go build app1_1.go
```

5. To test the fault tolerance, can run ```test2.sh``` script

## Mechansim
During the process, it will produce ```$STAGE_$TASK_PROC```, ```$STAGE_$TASK_ACKED``` and ```$STAGE_$TASK_STATE``` to record the status.

