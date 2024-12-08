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

1. ```
vim $SPARK_HOME/conf/spark-env.sh
```

2. ```
SPARK_MASTER_HOST=172.22.156.220

SPARK_WORKER_CORES=4       # Number of cores per worker
SPARK_WORKER_MEMORY=4g     # Memory per worker
SPARK_MASTER=spark://172.22.156.220:7077

 ---

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

3. 
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
submit job in master
```
spark-submit \
  --master spark://<master_ip>:7077 \
  --deploy-mode cluster \
  --total-executor-cores 40 \  # Adjust based on your setup
  --executor-memory 2g \
  traffic_signs_test1.py
```
simulate the failure
```
sleep 1.5
ssh user@<worker_ip1> "pkill -f spark"
ssh user@<worker_ip2> "pkill -f spark"
```
## Command

