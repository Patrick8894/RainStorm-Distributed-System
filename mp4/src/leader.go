package main

import (
    "fmt"
    "net"
    "strings"
)

type State int

const (
    Suspected State = iota
    Alive
    Down
    Join
)

type NodeInfo struct {
	ID 	string
    Address string
    State   State
}

type Task struct {
    Message string
    Port    int
	Stage  int
	Index  int
}

var SWIMPort = "8082"
var port = 8090
var workerPort = 8091
// Map of address to list of tasks
var addressTaskMap = make(map[string][]Task)
var cluster map[string]string
var clusterLock sync.Mutex

var ClientAddr string
var logFile string = "../../mp1/data/leader.log"

func main() {
    addr := net.UDPAddr{
        Port: port,
        IP:   net.ParseIP("0.0.0.0"),
    }

    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Println("Error starting UDP server:", err)
        return
    }
    defer conn.Close()

	cluster = getMembership()
	
    // Periodically update membership list
	go updateMembershipPeriodically()

    buffer := make([]byte, 1024)
    for {
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        message := string(buffer[:n])
        fmt.Println("Received message:", message)

        if strings.HasPrefix(message, "[Log]") {
            // worker send ACK message
            handleLogMessage(message, clientAddr)
        } else {
            clusterLock.Lock()
            if len(addressTaskMap) == 0 {
				ClientAddr = clientAddr
                processClientRequest(message)
            } else {
                fmt.Println("Cannot process client request, tasks are currently running")
            }
            clusterLock.Unlock()
        }
    }
}

func getMembership() map[string]string {
    /*
    Get the membership list from local SWIM protocol.
    */
    hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return nil
    }

    conn, err := net.Dial("udp", hostname + ":" + SWIMPort)
    if err != nil {
        fmt.Println("Error dialing introducer:", err)
        return nil
    }
    defer conn.Close()

    data := []byte("ls")
    _, err = conn.Write(data)
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }

    buffer := make([]byte, 4096)

    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Failed to read message:", err)
        return nil
    }

    var response map[string]NodeInfo
    err = json.Unmarshal(buffer[:n], &response)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return nil
    }

	addressSet := make(map[string]string)
	for _, nodeInfo := range response {
		address := strings.Split(nodeInfo.Address, ":")[0] // Remove the port
		addressSet[address] = address
	}

    return addressSet
}

func updateMembershipPeriodically() {
    for {
        updateMembership()
        time.Sleep(10 * time.Second) // Update membership every 10 seconds
    }
}

func updateMembership() {
    response := getMembership()

	clusterLock.Lock()
	defer clusterLock.Unlock()

    if mapsEqual(response, cluster) {
        return
    }

	// Find addresses missing in the new membership
    missingAddresses := findMissingAddresses(cluster, response)

    // Reschedule tasks for missing addresses
    for _, address := range missingAddresses {
        if tasks, exists := addressTaskMap[address]; exists {
            for _, task := range tasks {
                scheduleTask(task.Message, task.Stage, task.Index true)
            }
            delete(addressTaskMap, address) // Remove tasks after rescheduling
        }
    }

    cluster = response
}

func findMissingAddresses(oldMap, newMap map[string]NodeInfo) []string {
    var missingAddresses []string
    for address := range oldMap {
        if _, exists := newMap[address]; !exists {
            missingAddresses = append(missingAddresses, address)
        }
    }
    return missingAddresses
}

func mapsEqual(a, b map[string]global.NodeInfo) bool {
    if len(a) != len(b) {
        return false
    }
    for k, v := range a {
        if bv, ok := b[k]; !ok || v != bv {
            return false
        }
    }
    return true
}

func handleLogMessage(message string, workerAddr *net.UDPAddr) {
    parts := strings.Fields(message)
    if len(parts) != 3 {
        fmt.Println("Invalid log message format")
        return
    }

    stage, err := strconv.Atoi(parts[1])
    if err != nil {
        fmt.Println("Invalid stage in log message")
        return
    }

    index, err := strconv.Atoi(parts[2])
    if err != nil {
        fmt.Println("Invalid index in log message")
        return
    }

	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Task completed: stage=%d, index=%d\n", stage, index))
	log.Close()

    address := strings.Split(workerAddr.IP.String(), ":")[0]

    clusterLock.Lock()
    defer clusterLock.Unlock()

    if tasks, exists := addressTaskMap[address]; exists {
        for i, task := range tasks {
            if task.Stage == stage && task.Index == index {
                addressTaskMap[address] = append(tasks[:i], tasks[i+1:]...)
                break
            }
        }

        if len(addressTaskMap[address]) == 0 {
            delete(addressTaskMap, address)
            fmt.Printf("All tasks completed for address: %s\n", address)
        }
    }

	if len(addressTaskMap) == 0 {
        sendCompletionMessage()
    }
}

func sendCompletionMessage() {
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString("All tasks completed\n")
	log.Close()

    conn, err := net.DialUDP("udp", nil, ClientAddr)
    if err != nil {
        fmt.Println("Error connecting to client:", err)
        return
    }
    defer conn.Close()

    message := "All tasks completed"
    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Println("Error sending completion message:", err)
    }
}

func processClientRequest(message string) {
    parts := strings.Split(message, " ")
    if len(parts) != 7 {
        return "Error: Invalid message format"
    }

    op1Exe := parts[0]
    op2Exe := parts[1]
    hydfsSrcFile := parts[2]
    hydfsDestFilename := parts[3]
    numTasks := parts[4]
    X := parts[5]
    stateful := parts[6]

    numTasksInt, err := strconv.Atoi(numTasks)
    if err != nil {
        return "Error: Invalid numTasks value"
    }

	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Received request: %s\n", message))
	log.Close()

	var firstTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("3, %s, %s, %d, %s, %s", op2Exe, stateful, i, numTasks, hydfsDestFilename)
        result := scheduleTask(taskMessage, 3, i, false)
		if result == "" {
			i--
			continue
		}
        firstTaskResults = append(firstTaskResults, result)
    }

	var secondTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("2, %s, %d, %s, %s, %v", op1Exe, i, numTasks, X, firstTaskResults)
        result := scheduleTask(taskMessage, 2, i, false)
		if result == "" {
			i--
			continue
		}
        secondTaskResults = append(secondTaskResults, result)
    }

	var thirdTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("1, %s, %d, %s, %s", hydfsSrcFile, i, numTasks, secondTaskResults[i])
        result := scheduleTask(taskMessage, 1, i, false)
		if result == "" {
			i--
			continue
		}
        thirdTaskResults = append(thirdTaskResults, result)
    }
}

func scheduleTask(message string, stage int, index int, recover bool) string {
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Scheduling task: %s\n", message))
	log.Close()

	// Find the address with the least workload
	var leastLoadedAddress string
    minWorkload := int(^uint(0) >> 1) // Max int value

    for address := range cluster {
        workload := len(addressTaskMap[address])
        if workload < minWorkload {
            minWorkload = workload
            leastLoadedAddress = address
        }
    }

    if leastLoadedAddress == "" {
        return ""
    }

	// Send the message to the worker
    workerAddr := fmt.Sprintf("%s:%d", leastLoadedAddress, workerPort)
    conn, err := net.Dial("udp", workerAddr)
    if err != nil {
        return ""
    }
    defer conn.Close()

    fullMessage := fmt.Sprintf("%s, %t", message, recover)
    _, err = conn.Write([]byte(fullMessage))
    if err != nil {
        return ""
    }

	// Wait for success message
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        return ""
    }

	successMessage := string(buffer[:n])
    fmt.Println("Received success message:", successMessage)

	if !strings.HasPrefix(successMessage, "Success") {
		return ""
	}

    // Assume the success message contains the port number
    var port int
    _, err = fmt.Sscanf(successMessage, "Success: port=%d", &port)
    if err != nil {
        return ""
    }

	// Add the task to addressTaskMap
    taskPortPair := TaskPortPair{
        Message: message,
        Port:    port,
		Stage:   stage,
		Index:   index,
    }
    addressTaskMap[leastLoadedAddress] = append(addressTaskMap[leastLoadedAddress], taskPortPair)

    return leastLoadedAddress + ":" + port
}
