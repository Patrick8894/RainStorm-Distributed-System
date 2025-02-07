# Rainstorm Framework

Rainstorm is a distributed, scalable stream processing framework built with Go and gRPC. It enables efficient, real-time data processing across multiple nodes while integrating essential distributed system components such as logging, membership detection, and distributed file storage. Designed for high performance, Rainstorm achieves stream processing speeds comparable to Apache Spark Streaming.

## Features

### 1. Distributed Stream Processing
- Efficiently processes real-time data across multiple worker nodes.
- Utilizes a master node to distribute tasks and aggregate results.
- Designed to handle high-throughput data streams with minimal latency.

### 2. Distributed Logging System
- Aggregates system logs from remote machines into a centralized storage.
- Enables real-time log monitoring and analysis.
- Ensures fault tolerance and reliability.

### 3. Robust Membership Detection
- Implements the **SWIM protocol** with a **suspicion mechanism** for resilient node membership detection.
- Uses **gossip-style messages** packaged into **Gossip-Based Protocol (GBP)** to propagate membership changes.
- Handles node failures efficiently to maintain system stability.

### 4. Distributed File Storage System
- Utilizes the **Chord algorithm** for efficient file location and retrieval.
- Separates in-memory and disk-based storage for optimized synchronization performance.
- Implements **LRU caching** to improve GET request performance and reduce disk I/O operations.

### 5. Seamless System Integration
- Integrates **logging, membership detection, and file storage** into the stream processing framework.
- Designed for **scalability**, allowing arbitrary worker nodes to join or leave the system dynamically.
- Ensures **high availability** and fault tolerance through distributed mechanisms.

## Architecture Overview
The Rainstorm Framework follows a **master-worker architecture**, where:
- The **master node** coordinates task distribution, membership monitoring, and log aggregation.
- **Worker nodes** process incoming data streams, store distributed files, and participate in membership updates.
- The **Chord-based file system** ensures efficient data partitioning and retrieval across nodes.
- The **gossip protocol** keeps nodes updated on system changes with minimal overhead.