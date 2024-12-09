package main

import (
    "fmt"
    "net"
    "strings"
	"sync"
	"time"
	"encoding/json"
	"os"
	"strconv"
	"sort"
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
	Address string
}

var SWIMPort = "8082"
var port = 8090
var workerPort = "8091"
var addressTaskMap = make(map[string][]Task)
var cluster map[string]string
var clusterLock sync.Mutex

var ClientAddr *net.UDPAddr
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

		clusterLock.Lock()

        if strings.HasPrefix(message, "[Log]") {
            // worker send ACK message
            fmt.Println("Handling worker message", message)
            handleLogMessage(message, clientAddr, conn)
        } else {
            if len(addressTaskMap) == 0 {
                fmt.Println("Processing client request", clientAddr)
				ClientAddr = clientAddr
                processClientRequest(message)
            } else {
                fmt.Println("Cannot process client request, tasks are currently running")
            }
        }
		clusterLock.Unlock()
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
        time.Sleep(2 * time.Second) // Update membership every 10 seconds
    }
}

func updateMembership() {
    response := getMembership()

	clusterLock.Lock()
	defer clusterLock.Unlock()

	fmt.Printf("Updating membership\n")

    if mapsEqual(response, cluster) {
        return
    }

	// Find addresses missing in the new membership
    missingAddresses := findMissingAddresses(cluster, response)

	fmt.Printf("Missing addresses: %v\n", missingAddresses)

	cluster = response

	// find all stage 3 task
	stage3Tasks := make([]Task, 0)
	for _, tasks := range addressTaskMap {
		for _, t := range tasks {
			if t.Stage == 3 {
				stage3Tasks = append(stage3Tasks, t)
			}
		}
	}

	// sort stage 3 tasks by index
	sort.Slice(stage3Tasks, func(i, j int) bool {
		return stage3Tasks[i].Index < stage3Tasks[j].Index
	})

	fmt.Printf("Stage 3 tasks: %v\n", stage3Tasks)
	stage3Addresses := make([]string, len(stage3Tasks))
	for i, task := range stage3Tasks {
		stage3Addresses[i] = task.Address + ":" + strconv.Itoa(task.Port)
	}
	
	for _, task := range stage3Tasks {
		for j := 0; j < len(missingAddresses); j++ {
			if task.Address == missingAddresses[j] {
				address := scheduleTask(task.Message, task.Stage, task.Index, true)
				fmt.Printf("Rescheduling task: %s\n to address: %s\n", task.Message, address)
				stage3Addresses[task.Index] = address
				break
			}
		}
	}

	// find all stage 2 task
	stage2Tasks := make([]Task, 0)
	for _, tasks := range addressTaskMap {
		for _, t := range tasks {
			if t.Stage == 2 {
				stage2Tasks = append(stage2Tasks, t)
			}
		}
	}


	// sort stage 2 tasks by index
	sort.Slice(stage2Tasks, func(i, j int) bool {
		return stage2Tasks[i].Index < stage2Tasks[j].Index
	})

	fmt.Printf("Stage 2 tasks: %v\n", stage2Tasks)
	stage2Addresses := make([]string, len(stage2Tasks))
	for i, task := range stage2Tasks {
		stage2Addresses[i] = task.Address + ":" + strconv.Itoa(task.Port)
	}

	// find all stage 1 task
	stage1Tasks := make([]Task, 0)
	for _, tasks := range addressTaskMap {
		for _, t := range tasks {
			if t.Stage == 1 {
				stage1Tasks = append(stage1Tasks, t)
			}
		}
	}

	// sort stage 1 tasks by index
	sort.Slice(stage1Tasks, func(i, j int) bool {
		return stage1Tasks[i].Index < stage1Tasks[j].Index
	})

	for _, task := range stage2Tasks {
		missing := false
		for _, address := range missingAddresses {
			if task.Address == address {
				missing = true
				// update first task results with new address
				parts := strings.Split(task.Message, "^")
				task.Message = fmt.Sprintf("%s^%s^%s^%d^%s^%s^%v", parts[0], parts[1], parts[2], parts[3], parts[4], stage3Addresses)
				stage2Addresses[task.Index] = scheduleTask(task.Message, task.Stage, task.Index, true)
				fmt.Printf("Rescheduling task: %s\n to address: %s\n", task.Message, stage2Addresses[task.Index])
				// find previous stage address
				previousStageAddr := stage1Tasks[task.Index].Address + ":" + workerPort
				conn, err := net.Dial("udp", previousStageAddr)
				if err != nil {
					fmt.Println("Error connecting to worker:", err)
					continue
				}
				// Send new next stage message
				stage2AddressesArr := []string{stage2Addresses[task.Index]}
				_, err = conn.Write([]byte(fmt.Sprintf("Next,1,%d,%v", task.Index, stage2AddressesArr)))
				if err != nil {
					fmt.Println("Error sending message:", err)
					conn.Close()
					continue
				}
				conn.Close()
				break
			}
		}
		if !missing {
			// update next stage in the message of all stage 2 tasks
			workerAddr := fmt.Sprintf("%s:%s", task.Address, workerPort)
			conn, err := net.Dial("udp", workerAddr)
			if err != nil {
				fmt.Println("Error connecting to worker:", err)
				continue
			}
			// Send new next stage message
			_, err = conn.Write([]byte(fmt.Sprintf("Next,2,%d,%v", task.Index, stage3Addresses)))
			if err != nil {
				fmt.Println("Error sending message:", err)
				conn.Close()
				continue
			}
			conn.Close()
		}
	}

	for address := range addressTaskMap {
		if _, exists := response[address]; !exists {
			delete(addressTaskMap, address)
		}
	}
}

func findMissingAddresses(oldMap, newMap map[string]string) []string {
    var missingAddresses []string
    for address := range oldMap {
        if _, exists := newMap[address]; !exists {
            missingAddresses = append(missingAddresses, address)
        }
    }
    return missingAddresses
}

func mapsEqual(a, b map[string]string) bool {
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

func handleLogMessage(message string, workerAddr *net.UDPAddr, conn *net.UDPConn) {
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
		return
	}

    fmt.Printf("create log file\n")
	
	log.WriteString(fmt.Sprintf("Task completed: stage=%d, index=%d\n", stage, index))
	log.Close()

	for addr, tasks := range addressTaskMap {
		for i, task := range tasks {
			if task.Stage == stage && task.Index == index {
				fmt.Printf("Task completed: stage=%d, index=%d\n", stage, index)
				addressTaskMap[addr] = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
		if len(addressTaskMap[addr]) == 0 {
			delete(addressTaskMap, addr)
			fmt.Printf("All tasks completed for address: %s\n", addr)
			fmt.Printf("len(addressTaskMap): %d\n", len(addressTaskMap))
			if len(addressTaskMap) == 0 {
				sendCompletionMessage(conn)
			}
		}
	}

	// print addressTaskMap
	fmt.Printf("AddressTaskMap after:\n")
	for addr, tasks := range addressTaskMap {
		fmt.Printf("Address: %s\n", addr)
		for _, task := range tasks {
			fmt.Printf("Task: %s\n", task.Message)
		}
	}
}

func sendCompletionMessage(conn *net.UDPConn) {
	fmt.Printf("All tasks completed\n")
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return
	}
	
	log.WriteString("All tasks completed\n")
	log.Close()

    message := "All tasks completed"
    _, err = conn.WriteToUDP([]byte(message), ClientAddr)
    if err != nil {
        fmt.Println("Error sending completion message:", err)
    }
}

func processClientRequest(message string) {
    parts := strings.Split(message, "^")
    fmt.Println("Received message parts:", parts)
    if len(parts) != 7 {
        fmt.Println("Invalid message, message legnth is less then 7")
        return
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
        fmt.Println("Invalid number of tasks:", numTasks)
        return
    }

	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return
	}
	
    fmt.Printf("Creat Log file\n")

	log.WriteString(fmt.Sprintf("Received request: %s\n", message))
	log.Close()

	var firstTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("3^%s^%s^%d^%s^%s", op2Exe, stateful, i, numTasks, hydfsDestFilename)
        result := scheduleTask(taskMessage, 3, i, false)
		if result == "" {
			i--
			continue
		}
        firstTaskResults = append(firstTaskResults, result)
    }

	var secondTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("2^%s^%s^%d^%s^%s^%v", op1Exe, op2Exe, i, numTasks, X, firstTaskResults)
        result := scheduleTask(taskMessage, 2, i, false)
		if result == "" {
			i--
			continue
		}
        secondTaskResults = append(secondTaskResults, result)
    }

	var thirdTaskResults []string
    for i := 0; i < numTasksInt; i++ {
        taskMessage := fmt.Sprintf("1^%s^%d^%s^%s", hydfsSrcFile, i, numTasks, secondTaskResults[i])
        result := scheduleTask(taskMessage, 1, i, false)
		if result == "" {
			i--
			continue
		}
        thirdTaskResults = append(thirdTaskResults, result)
    }
}

func scheduleTask(message string, stage int, index int, recover bool) string {
    fmt.Printf("Scheduling task: stage=%d, index=%d\n", stage, index)

	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Scheduling task: stage=%d, index=%d\n", stage, index))
	log.Close()

	// Find the address with the least workload
	var leastLoadedAddress string
    minWorkload := int(^uint(0) >> 1) // Max int value

    addresses := make([]string, 0, len(cluster))
	for address := range cluster {
		addresses = append(addresses, address)
	}

	sort.Strings(addresses)

	fmt.Printf("Addresses: %v\n", addresses)

	fmt.Printf("AddressTaskMap: %v\n", addressTaskMap)

	for _, address := range addresses {
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
    workerAddr := fmt.Sprintf("%s:%s", leastLoadedAddress, workerPort)
    conn, err := net.Dial("udp", workerAddr)
    if err != nil {
        return ""
    }
    defer conn.Close()

    fullMessage := fmt.Sprintf("%s^%t", message, recover)
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
    taskPortPair := Task{
        Message: message,
        Port:    port,
		Stage:   stage,
		Index:   index,
		Address: leastLoadedAddress,
    }
    addressTaskMap[leastLoadedAddress] = append(addressTaskMap[leastLoadedAddress], taskPortPair)

    return leastLoadedAddress + ":" + strconv.Itoa(port)
}
