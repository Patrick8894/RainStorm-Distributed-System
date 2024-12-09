package main

import (
	"sync"
	"bufio"
    "fmt"
    "net"
	"os"
	"os/exec"
    "strings"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

var workerPort = 8091
var portMutex sync.Mutex
var nextPort = 9000
var leader = "fa24-cs425-6605.cs.illinois.edu"
var leaderPort = "8090"

// a map to remember next stage addresses for each task
var nextStageAddrMap = make(map[string][]string)
var nextStageAddrMutex sync.Mutex

func main() {
    addr := net.UDPAddr{
        Port: workerPort,
        IP:   net.ParseIP("0.0.0.0"),
    }

    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Println("Error starting UDP server:", err)
        return
    }
    defer conn.Close()

    fmt.Printf("Worker listening on port %d\n", workerPort)

    buffer := make([]byte, 1024)
    for {
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        message := string(buffer[:n])
        fmt.Println("Received message:", message)

		if strings.HasPrefix(message, "Next") {
			nextStageAddrMutex.Lock()
			parts := strings.Split(message, ",")
			ID := parts[1] + " " + parts[2]
			nextStageAddrMap[ID] = strings.Fields(strings.Trim(parts[3], "[]"))
			nextStageAddrMutex.Unlock()
			continue
		}

        // Schedule the task
        port := processTaskRequest(message)
		
        var response string
        if port != "" {
            response = fmt.Sprintf("Success: port=%s", port)
        } else {
            response = "Failure: Unable to process request"
        }

        _, err = conn.WriteToUDP([]byte(response), clientAddr)
        if err != nil {
            fmt.Println("Error sending response:", err)
            continue
        }
    }
}

func processTaskRequest(message string) string {

    parts := strings.Split(message, "^")

    taskType := strings.TrimSpace(parts[0])

    portMutex.Lock()
    port := nextPort
    nextPort++
    portMutex.Unlock()

	logFile := "../../mp1/data/worker.log"
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Task %s scheduled on port %d\n", taskType, port))
	log.Close()

	fmt.Printf("Task %s scheduled on port %d\n", taskType, port)
    switch taskType {
    case "1":
        go startTaskServerStage1(port, parts[1:])
    case "2":
        go startTaskServerStage2(port, parts[1:])
    case "3":
        go startTaskServerStage3(port, parts[1:])
    default:
        return ""
    }

    return fmt.Sprintf("%d", port)
}

func startTaskServerStage1(port int, params []string) {
    if len(params) < 5 {
        fmt.Println("Invalid parameters for stage 1")
        return
    }

	hydfsSrcFile := strings.TrimSpace(params[0])
    taskNo := strings.TrimSpace(params[1])
    totalNumstr := strings.TrimSpace(params[2])
    nextStage := strings.TrimSpace(params[3])

	totalNum, err := strconv.Atoi(totalNumstr)

	ID := fmt.Sprintf("1 %s", taskNo)
	
	END := false
	endPtr := &END

	nextStageAddrMutex.Lock()
	// clean up the next stage address map
	nextStageAddrMap[ID] = []string{}
	nextStageAddrMap[ID] = append(nextStageAddrMap[ID], nextStage)
	fmt.Printf("Next stage for task %s: %v\n", taskNo, nextStageAddrMap[ID])
	nextStageAddrMutex.Unlock()

    localFilename := fmt.Sprintf("%s/tmp/1_%s", os.Getenv("HOME"), taskNo)
    cmd := exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", localFilename, "--HyDFSfilename", hydfsSrcFile)

    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error executing command for stage 1: %v\n", err)
        return
    }

	file, err := os.Open(localFilename)
    if err != nil {
        fmt.Printf("Error opening file %s: %v\n", localFilename, err)
        return
    }
    defer file.Close()

	ackMap := make(map[string]int)

	// setup a listen udp connection
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	defer conn.Close()

	fmt.Printf("Starting UDP server on port %d\n", port)

    go handleStage1Acks(ID, ackMap, conn, endPtr)

	go handleStage1resend(ID, ackMap, conn, endPtr)

	fmt.Printf("Start sending data\n")

    scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading file %s: %v\n", localFilename, err)
        return
    }

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	fmt.Printf("Total lines: %d\n", len(lines))

	totalLines := len(lines)
	partitionSize := totalLines / totalNum
	taskNoInt, err := strconv.Atoi(taskNo)
	start := partitionSize * taskNoInt
	end := start + partitionSize

	if taskNoInt == totalNum-1 {
		end = totalLines
	}

    for i := start; i < end; i++ {
		line := lines[i]
		nextStageAddrMutex.Lock()

		nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStageAddrMap[ID][0])
		if err != nil {
			fmt.Printf("Error resolving UDP address: %v\n", err)
			nextStageAddrMutex.Unlock()
			i--
			continue
		}

		_, err = conn.WriteToUDP([]byte("ACK@" + line), nextStageUdpAddr)
		if err != nil {
			fmt.Printf("Error sending line to next stage %s: %v\n", nextStage, err)
			nextStageAddrMutex.Unlock()
			i--
			continue
		}
		ackMap[line]++
		nextStageAddrMutex.Unlock()
	}

	fmt.Printf("All data sent \n")

	for {
        nextStageAddrMutex.Lock()
        if len(ackMap) == 0 {
            nextStageAddrMutex.Unlock()
            break
        }
        nextStageAddrMutex.Unlock()
        time.Sleep(1 * time.Second)
    }

	fmt.Printf("All ACKs received\n")

	// Send end of task message to next stage
    endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
	nextStageAddrMutex.Lock()
	nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStageAddrMap[ID][0])
	if err != nil {
		fmt.Printf("Error resolving UDP address: %v\n", err)
		nextStageAddrMutex.Unlock()
		return
	}
	ackMap[endMessage]++
	nextStageAddrMutex.Unlock()

	// find the current time
	timeoutDuration := 30 * time.Second
	timeoutTime := time.Now().Add(timeoutDuration)

	// wait for all ACKs
	for {
		if time.Now().After(timeoutTime) {
			fmt.Println("Timeout reached, exiting loop.")
			break
		}
		nextStageAddrMutex.Lock()
		if len(ackMap) == 0 {
			nextStageAddrMutex.Unlock()
			break
		}
		nextStageAddrMutex.Unlock()
		time.Sleep(1 * time.Second)
	}

	nextStageAddrMutex.Lock()

	fmt.Printf("Sending end of task message to next stage %s\n", nextStageAddrMap[ID][0])

    _, err = conn.WriteToUDP([]byte(endMessage), nextStageUdpAddr)
    if err != nil {
        fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStageAddrMap[ID][0], err)
        return
    }

	_, err = conn.WriteToUDP([]byte(endMessage), nextStageUdpAddr)
    if err != nil {
        fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStageAddrMap[ID][0], err)
        return
    }

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)

	leaderUdpAddr, err := net.ResolveUDPAddr("udp", leaderAddr)
	if err != nil {
		fmt.Printf("Error resolving UDP address: %v\n", err)
		return
	}

	logMessage := fmt.Sprintf("[Log] 1 %s", taskNo)
    _, err = conn.WriteToUDP([]byte(logMessage), leaderUdpAddr)
    if err != nil {
        fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
    }
	nextStageAddrMutex.Unlock()

	END = true
	fmt.Printf("End of task\n")
}

func handleStage1Acks(ID string, ackMap map[string]int, conn *net.UDPConn, endPtr *bool) {
    ackBuffer := make([]byte, 1024)
    for {
		if *endPtr {
			break
		}

        n, _, err := conn.ReadFromUDP(ackBuffer)
        if err != nil {
            continue
        }

        ack := string(ackBuffer[:n])
		// fmt.Printf("Received ACK: %s\n", ack)
        parts := strings.SplitN(ack, "@", 2)
        if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
            fmt.Printf("Invalid ACK received: %s\n", ack)
            continue
        }

        line := parts[1]

		fmt.Printf("Received ACK for line: %s\n", line)

		nextStageAddrMutex.Lock()

        if _, exists := ackMap[line]; exists {
            delete(ackMap, line)
        }
        nextStageAddrMutex.Unlock()
    }
}

func handleStage1resend(ID string, ackMap map[string]int, conn *net.UDPConn, endPtr *bool) {
	for {
		nextStageAddrMutex.Lock()
		if *endPtr {
			nextStageAddrMutex.Unlock()
			break
		}
		nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStageAddrMap[ID][0])
		if err != nil {
			fmt.Printf("Error resolving UDP address: %v\n", err)
			nextStageAddrMutex.Unlock()
			continue
		}
		for line := range ackMap {
			_, err := conn.WriteToUDP([]byte(line), nextStageUdpAddr)
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				continue
			}
		}
		nextStageAddrMutex.Unlock()
		time.Sleep(2 * time.Second)
	}
}

func startTaskServerStage2(port int, params []string) {
    if len(params) < 6 {
        fmt.Println("Invalid parameters for stage 2")
        return
    }

	opFile1 := strings.TrimSpace(params[0])
	opFile2 := strings.TrimSpace(params[1])
    taskNo := strings.TrimSpace(params[2])
    totalNum := strings.TrimSpace(params[3])
	X := strings.TrimSpace(params[4])
    nextStageListStr := strings.TrimSpace(params[5])
    recover := strings.TrimSpace(params[6])

	END := false
	endPtr := &END

	nextStageListStr = strings.Trim(nextStageListStr, "[]")
    nextStageList := strings.Fields(nextStageListStr)

	ID := fmt.Sprintf("2 %s", taskNo)

	nextStageAddrMutex.Lock()
	nextStageAddrMap[ID] = nextStageList
	nextStageAddrMutex.Unlock()

	seen := make(map[string]int)

	processedFilename := fmt.Sprintf("%s/tmp/2_%s_PROC", os.Getenv("HOME"), taskNo)
	ackedFilename := fmt.Sprintf("%s/tmp/2_%s_ACKED", os.Getenv("HOME"), taskNo)

    // Log the received parameters
    fmt.Printf("Starting task server stage 2 on port %d with params: opFile1=%s, opFile2=%s, taskNo=%s, totalNum=%s, nextStageList=%v, recover=%s\n",
        port, opFile1, opFile2, taskNo, totalNum, nextStageList, recover)

	processInput := make(map[string]int)
	ackMap := make(map[string]int)

	fmt.Printf("Recover: %s\n", recover)

	if recover == "true" {
		// Get the processed data from HyDFS
		cmd := exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to get file from HyDFS: %v\n", err)
			return
		}
		// load the processed data into the processedInput map
		file, err := os.Open(processedFilename)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", processedFilename, err)
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "@", 2)
			processInput[parts[0]]++
			ackMap[parts[1]]++
		}

		fmt.Printf("Processed input: %d\n", len(processInput))

		file.Close()
		
		// Get the acked data from HyDFS
		cmd = exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", ackedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to get file from HyDFS: %v\n", err)
			return
		}

		// load the acked data into the ackMap
		file, err = os.Open(ackedFilename)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", ackedFilename, err)
			return
		}

		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			delete(ackMap, line)
		}

		file.Close()
	} else {
		// Create the file
		
		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			return
		}
		file.Close()

		// Run the command to create the file in HyDFS
		cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
			return
		}

		// Create the file
		file, err = os.Create(ackedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", ackedFilename, err)
			return
		}
		file.Close()

		// Run the command to create the file in HyDFS
		cmd = exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", ackedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
			return
		}
	}

	fmt.Printf("Starting UDP server on port %d\n", port)

	addr := net.UDPAddr{
        Port: port,
        IP:   net.ParseIP("0.0.0.0"),
    }

	conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Printf("Error starting UDP server on port %d: %v\n", port, err)
        return
    }
    defer conn.Close()

	go handleStage2resend(ID, ackMap, conn, endPtr)

	endOfTask := false

	fmt.Printf("Start receiving data\n")

	buffer := make([]byte, 1024)
    for {
		// set a timeout for the read operation
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout reached, exiting loop.")
				break
			}
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received message: %s\n", request)

		if strings.HasPrefix(request, "END_OF_TASK") {
			fmt.Printf("Received end of task message\n")
			endOfTask = true
			nextStageAddrMutex.Lock()
			_, err = conn.WriteToUDP([]byte("ACK@" + request), clientAddr)
			if err != nil {
				fmt.Printf("Error sending ACK to previous stage: %s\n", clientAddr)
				continue
			}
			if len(ackMap) == 0 {
				nextStageAddrMutex.Unlock()
				break
			}
			nextStageAddrMutex.Unlock()
			continue
		}

		if strings.HasPrefix(request, "ACK") {
			handleStage2Acks(ID, ackMap, ackedFilename, taskNo, request)
			if endOfTask {
				nextStageAddrMutex.Lock()
				if len(ackMap) == 0 {
					nextStageAddrMutex.Unlock()
					break
				}
				nextStageAddrMutex.Unlock()
			}
			continue
		}

		// fmt.Printf("Processing request: %s\n", request)

		if _, exists := processInput[request]; exists {
			conn.WriteToUDP([]byte("ACK@" + request), clientAddr)
			continue
		}

        // Run the external program with the request as input
        cmd := exec.Command("../ops/" + opFile1, X)
        cmd.Stdin = strings.NewReader(request)
        output, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error running external program: %v\n", err)
            continue
        }

		if string(output) == "0\n" {
			ackMessage := fmt.Sprintf("ACK@%s", request)
			_, err = conn.WriteToUDP([]byte(ackMessage), clientAddr)
			if err != nil {
				fmt.Printf("Error sending ACK to previous stage: %s\n", clientAddr)
				continue
			}
			continue
		}

		seen[request] = 1

		cmd = exec.Command("../ops/" + opFile2)
		cmd.Stdin = strings.NewReader(request)
		output, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error running external program: %v\n", err)
			continue
		}

		processInput[request] = 1
		fmt.Printf("Processed input: %d\n", len(processInput))

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		_, err = file.WriteString(fmt.Sprintf("%s@%s\n", request, request + "^" + string(output)))
        if err != nil {
            fmt.Printf("Error writing to file %s: %v\n", processedFilename, err)
            continue
        }
		file.Close()

		// Send the processed data to HyDFS
		cmd = exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			continue
		}

		// Hash the response to determine the next stage
        hash := sha256.Sum256([]byte(string(output)))
        hashValue := hex.EncodeToString(hash[:])
        nextStageIndex := int(hashValue[0]) % len(nextStageList)

		nextStageAddrMutex.Lock()
		ackMap[request + "^" + string(output)]++
        nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStageAddrMap[ID][nextStageIndex])
		if err != nil {
			fmt.Printf("Error resolving UDP address: %v\n", err)
			nextStageAddrMutex.Unlock()
			continue
		}

		_, err = conn.WriteToUDP([]byte(request + "^" + string(output)), nextStageUdpAddr)
        if err != nil {
            fmt.Printf("Error sending response to next stage %s: %v\n", nextStageAddrMap[ID][nextStageIndex], err)
            nextStageAddrMutex.Unlock()
			continue
        }
		nextStageAddrMutex.Unlock()

		// Send ACK to previous stage
		ackMessage := fmt.Sprintf("ACK@%s", request)
		_, err = conn.WriteToUDP([]byte(ackMessage), clientAddr)
		if err != nil {
			fmt.Printf("Error sending ACK to previous stage: %s\n", clientAddr)
			continue
		}

        // fmt.Printf("Sent ack to previus stage: %s\n", clientAddr)
	}

	// Send end of task message to all next stages
	nextStageAddrMutex.Lock()
	fmt.Printf("All data sent\n")
	endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
	for _, nextStage := range nextStageAddrMap[ID] {
		fmt.Printf("Sending end of task message to next stage %s\n", nextStage)
		nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStage)
		if err != nil {
			fmt.Printf("Error resolving UDP address: %v\n", err)
			continue
		}

		_, err = conn.WriteToUDP([]byte(endMessage + "^" + nextStage), nextStageUdpAddr)
		if err != nil {
			fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStage, err)
			continue
		}
		ackMap[endMessage + "^" + nextStage]++
	}
	nextStageAddrMutex.Unlock()

	timeoutDuration := 30 * time.Second
	timeoutTime := time.Now().Add(timeoutDuration)
	readTimeout := 5 * time.Second

	// Wait for all ACKs
	for {
		if time.Now().After(timeoutTime) {
			fmt.Println("Timeout reached, exiting loop.")
			break
		}
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
		ack := string(buffer[:n])
		fmt.Printf("Received ACK: %s\n", ack)
		if strings.HasPrefix(ack, "ACK") {
			fmt.Printf("ackMap: %v\n", ackMap)
			handleStage2Acks(ID, ackMap, ackedFilename, taskNo, ack)
			if len(ackMap) == 0 {
				break
			}
		}
	}

	fmt.Printf("All ACKs received\n")

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)

	leaderUdpAddr, err := net.ResolveUDPAddr("udp", leaderAddr)
	if err != nil {
		fmt.Printf("Error resolving UDP address: %v\n", err)
		return
	}

	logMessage := fmt.Sprintf("[Log] 2 %s", taskNo)
	_, err = conn.WriteToUDP([]byte(logMessage), leaderUdpAddr)
	if err != nil {
		fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
	}
	*endPtr = true
	fmt.Printf("End of task, process input number: %d\n", len(seen))
}

func handleStage2Acks(ID string, ackMap map[string]int, ackedFilename string, taskNo string, ack string) {

	parts := strings.SplitN(ack, "@", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
		fmt.Printf("Invalid ACK received: %s\n", ack)
		return
	}

	fmt.Printf("Received ACK: %s\n", ack)

	line := parts[1]

	nextStageAddrMutex.Lock()
	if _, exists := ackMap[line]; exists {

		file, err := os.Create(ackedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", ackedFilename, err)
			nextStageAddrMutex.Unlock()
			return
		}

		fmt.Printf("Writing to file: %s, %s\n", ackedFilename, line)

		_, err = file.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("Error writing to file %s: %v\n", ackedFilename, err)
			nextStageAddrMutex.Unlock()
			return
		}
		file.Close()

		cmd := exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", ackedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			nextStageAddrMutex.Unlock()
			return
		}

		delete(ackMap, line)
	}
	nextStageAddrMutex.Unlock()
}

func handleStage2resend(ID string, ackMap map[string]int, conn *net.UDPConn, endPtr *bool) {
	for {
		nextStageAddrMutex.Lock()
		if *endPtr {
			nextStageAddrMutex.Unlock()
			break
		}
		for line := range ackMap {
			lineParts := strings.SplitN(line, "^", 2)
			hash := sha256.Sum256([]byte(lineParts[1]))
			hashValue := hex.EncodeToString(hash[:])
			nextStageIndex := int(hashValue[0]) % len(nextStageAddrMap[ID])
			nextStage := nextStageAddrMap[ID][nextStageIndex]

			fmt.Printf("Resending line %s to next stage %s\n", line, nextStage)

			nextStageUdpAddr, err := net.ResolveUDPAddr("udp", nextStage)
			if err != nil {
				fmt.Printf("Error resolving UDP address: %v\n", err)
				nextStageAddrMutex.Unlock()
				continue
			}

			_, err = conn.WriteToUDP([]byte(line), nextStageUdpAddr)
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				continue
			}
		}
		nextStageAddrMutex.Unlock()
		time.Sleep(2 * time.Second)
	}
}

func startTaskServerStage3(port int, params []string) {
	if len(params) < 6 {
        fmt.Println("Invalid parameters for stage 3")
        return
    }

    opFile := strings.TrimSpace(params[0])
    stateful := strings.TrimSpace(params[1])
    taskNo := strings.TrimSpace(params[2])
    totalNumstr := strings.TrimSpace(params[3])
    hydfsDestFilename := strings.TrimSpace(params[4])
    recover := strings.TrimSpace(params[5])

	totalNum, err := strconv.Atoi(totalNumstr)

	receivedEndOfTask := make(map[string]int)

	// Log the received parameters
    fmt.Printf("Starting task server stage 3 on port %d with params: opFile=%s, stateful=%s, taskNo=%s, totalNum=%s, hydfsDestFilename=%s, recover=%s\n",
        port, opFile, stateful, taskNo, totalNum, hydfsDestFilename, recover)

	processedInput := make(map[string]int)
	processedFilename := fmt.Sprintf("%s/tmp/3_%s_PROC", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path
	state := make(map[string]int)
	stateFilename := fmt.Sprintf("%s/tmp/3_%s_STATE", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path
	outputFilename := fmt.Sprintf("%s/tmp/3_%s_OUTPUT", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path

	if recover == "true" {
		// Get the processed data from HyDFS
		cmd := exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to get file from HyDFS: %v\n", err)
			return
		}
		// load the processed data into the processedInput map
		file, err := os.Open(processedFilename)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", processedFilename, err)
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			processedInput[line] = 1
		}

		fmt.Printf("[Recovery]: Processed input: %d\n", len(processedInput))

		file.Close()

		if stateful == "stateful" {
			// Get the state data from HyDFS
			cmd := exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command to get file from HyDFS: %v\n", err)
				return
			}

			// load the state data into the state map
			file, err := os.Open(stateFilename)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", stateFilename, err)
				return
			}
			
			isKey := true
			scanner := bufio.NewScanner(file)
			var key string
			for scanner.Scan() {
				line := scanner.Text()
				if isKey {
					key = line
				} else {
					value, _ := strconv.Atoi(line)
					state[key] = value
				}
				isKey = !isKey
			}

			file.Close()
		}
	} else {
		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			return
		}
		file.Close()

		cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
			return
		}

		if stateful == "stateful" {		
			file, err := os.Create(stateFilename)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", stateFilename, err)
				return
			}
			file.Close()

			cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
				return
			}
		}
	}

	// Start the UDP server
    addr := net.UDPAddr{
        Port: port,
        IP:   net.ParseIP("0.0.0.0"),
    }

    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Printf("Error starting UDP server on port %d: %v\n", port, err)
        return
    }
    defer conn.Close()

	buffer := make([]byte, 1024)
    for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout reached, exiting loop.")
				break
			}
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received request: %s\n", request)

		if strings.HasPrefix(request, "END_OF_TASK") {
			_, err = conn.WriteToUDP([]byte("ACK@" + request), clientAddr)
			if err != nil {
				fmt.Printf("Error sending ACK to previous stage: %s\n", clientAddr)
				continue
			}

			receivedEndOfTask[request]++
			if len(receivedEndOfTask) == totalNum {
				break
			}
			continue
		}

		if _, exists := processedInput[request]; exists {
			conn.WriteToUDP([]byte("ACK@" + request), clientAddr)
			continue
		}

		parts := strings.Split(request, "^")

		processedInput[request] = 1
		fmt.Printf("Processed input: %d\n", len(processedInput))
		if stateful == "stateful" {
			state[parts[1]] += 1
		}

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		// fmt.Printf("Writing to file: %s, %s\n", processedFilename, request)

		_, err = file.WriteString(request + "\n")
        if err != nil {
            fmt.Printf("Error writing to file %s: %v\n", processedFilename, err)
            continue
        }
		file.Close()

		// Send the processed data to HyDFS
		cmd := exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			continue
		}

		if stateful == "stateful" {
			file, err := os.Create(stateFilename)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", stateFilename, err)
				continue
			}

			for key, value := range state {
				_, err = file.WriteString(fmt.Sprintf("%s\n%d\n", key, value))
				if err != nil {
					fmt.Printf("Error writing to file %s: %v\n", stateFilename, err)
					continue
				}
			}

			// Send the processed data to HyDFS
			cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
				continue
			}
		}

		conn.WriteToUDP([]byte("ACK@" + request), clientAddr)

        // fmt.Printf("Sent processed data to HyDFS: %s\n", hydfsDestFilename)
    }

	if stateful == "stateful" {
		file, err := os.Create(outputFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outputFilename, err)
			return
		}

		for key, value := range state {
			_, err = file.WriteString(fmt.Sprintf("%s %d\n", key, value))
			if err != nil {
				fmt.Printf("Error writing to file %s: %v\n", outputFilename, err)
				return
			}
		}

		cmd := exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", outputFilename, "--HyDFSfilename", hydfsDestFilename)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			return
		}
	} else {
		file, err := os.Create(outputFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outputFilename, err)
			return
		}

		for key := range processedInput {
			lineParts := strings.Split(key, "^")
			fmt.Printf("Writing to file: %s, %s\n", outputFilename, lineParts[1])
			_, err = file.WriteString(lineParts[1] + "\n")
			if err != nil {
				fmt.Printf("Error writing to file %s: %v\n", outputFilename, err)
				return
			}
		}

		fmt.Printf("Output file: %s, Hydfs file: %s\n", outputFilename, hydfsDestFilename)
		cmd := exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", outputFilename, "--HyDFSfilename", hydfsDestFilename)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			return
		}
	}

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)
	leaderUdpAddr, err := net.ResolveUDPAddr("udp", leaderAddr)
	if err != nil {
		fmt.Printf("Error resolving UDP address: %v\n", err)
		return
	}

	logMessage := fmt.Sprintf("[Log] 3 %s", taskNo)
	_, err = conn.WriteToUDP([]byte(logMessage), leaderUdpAddr)
	if err != nil {
		fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
	}
	fmt.Printf("End of task, process input number: %d\n", len(processedInput))
}