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

		if strings.HasPrefix(message, "NEXT_STAGE") {
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

	nextStageAddrMutex.Lock()
	nextStageAddrMap[ID] = append(nextStageAddrMap[ID], nextStage)
	nextStageAddrMutex.Unlock()

    localFilename := fmt.Sprintf("%s/1_%s", os.Getenv("HOME"), taskNo)
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

    go handleStage1Acks(ID, ackMap)

	go handleStage1resend(ID, ackMap)

    scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading file %s: %v\n", localFilename, err)
        return
    }

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

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
		fmt.Printf("line: %s, address: %s\n", line, nextStageAddrMap)
		nextStageAddrMutex.Lock()
		fmt.Printf("achienve lock\n")
		conn, err := net.Dial("udp", nextStageAddrMap[ID][0])
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStage, err)
			nextStageAddrMutex.Unlock()
			i--
			continue
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("Error sending line to next stage %s: %v\n", nextStage, err)
			nextStageAddrMutex.Unlock()
			i--
			continue
		}
		ackMap[line]++
		conn.Close()
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

	// Send end of task message to next stage
    endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
	conn, err := net.Dial("udp", nextStageAddrMap[ID][0])
	if err != nil {
		fmt.Printf("Error connecting to next stage %s: %v\n", nextStageAddrMap[ID], err)
		return
	}
    _, err = conn.Write([]byte(endMessage))
    if err != nil {
        fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStageAddrMap[ID], err)
        return
    }

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)
    logConn, err := net.Dial("udp", leaderAddr)
    if err != nil {
        fmt.Printf("Error connecting to leader %s: %v\n", leaderAddr, err)
        return
    }
    defer logConn.Close()

	logMessage := fmt.Sprintf("[Log] 1 %s", taskNo)
    _, err = logConn.Write([]byte(logMessage))
    if err != nil {
        fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
    }
}

func handleStage1Acks(ID string, ackMap map[string]int) {
    ackBuffer := make([]byte, 1024)
    for {
		nextStageAddrMutex.Lock()
		conn, err := net.Dial("udp", nextStageAddrMap[ID][0])
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStageAddrMap[ID], err)
			nextStageAddrMutex.Unlock()
			continue
		}
		nextStageAddrMutex.Unlock()
        n, err := conn.Read(ackBuffer)
        if err != nil {
            fmt.Printf("Error receiving ACK: %v\n", err)
            continue
        }

        ack := string(ackBuffer[:n])
        parts := strings.SplitN(ack, "@", 2)
        if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
            fmt.Printf("Invalid ACK received: %s\n", ack)
            continue
        }

        line := parts[1]

		nextStageAddrMutex.Lock()

        if _, exists := ackMap[line]; exists {
            delete(ackMap, line)
        }
        nextStageAddrMutex.Unlock()
    }
}

func handleStage1resend(ID string, ackMap map[string]int) {
	for {
		nextStageAddrMutex.Lock()
		conn, err := net.Dial("udp", nextStageAddrMap[ID][0])
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStageAddrMap[ID], err)
			nextStageAddrMutex.Unlock()
			continue
		}
		for line := range ackMap {
			_, err := conn.Write([]byte(line))
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				nextStageAddrMutex.Unlock()
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

    opFile := strings.TrimSpace(params[0])
    taskNo := strings.TrimSpace(params[1])
    totalNum := strings.TrimSpace(params[2])
	X := strings.TrimSpace(params[3])
    nextStageListStr := strings.TrimSpace(params[4])
    recover := strings.TrimSpace(params[5])

	nextStageListStr = strings.Trim(nextStageListStr, "[]")
    nextStageList := strings.Fields(nextStageListStr)

	ID := fmt.Sprintf("2 %s", taskNo)

	nextStageAddrMutex.Lock()
	nextStageAddrMap[ID] = nextStageList
	nextStageAddrMutex.Unlock()

	processedFilename := fmt.Sprintf("%s/2_%s_PROC", os.Getenv("HOME"), taskNo)
	ackedFilename := fmt.Sprintf("%s/2_%s_ACKED", os.Getenv("HOME"), taskNo)

    // Log the received parameters
    fmt.Printf("Starting task server stage 2 on port %d with params: opFile=%s, taskNo=%s, totalNum=%s, nextStageList=%v, recover=%s\n",
        port, opFile, taskNo, totalNum, nextStageList, recover)

	processInput := make(map[string]int)
	ackMap := make(map[string]int)

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
			processInput[line]++
			ackMap[line]++
		}

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

    go handleStage2Acks(ID, ackMap, ackedFilename, taskNo)

	go handleStage2resend(ID, ackMap)

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
        n, _, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received request: %s\n", request)

		if request == "END_OF_TASK" {
			break
		}

		if _, exists := processInput[request]; exists {
			conn.Write([]byte("ACK@" + request))
			continue
		}

        // Run the external program with the request as input
        cmd := exec.Command("../ops/" + opFile, X)
        cmd.Stdin = strings.NewReader(request)
        output, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error running external program: %v\n", err)
            continue
        }

		if string(output) == "0" {
			continue
		}

		processInput[request] = 1

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		_, err = file.WriteString(fmt.Sprintf("%s\n", request))
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
        hash := sha256.Sum256([]byte(request))
        hashValue := hex.EncodeToString(hash[:])
        nextStageIndex := int(hashValue[0]) % len(nextStageList)

		nextStageAddrMutex.Lock()
        nextStage := nextStageAddrMap[ID][nextStageIndex]

		nextConns, err := net.Dial("udp", nextStage)
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStage, err)
			nextStageAddrMutex.Unlock()
			continue
		}
		defer conn.Close()

		_, err = nextConns.Write([]byte(request))
        if err != nil {
            fmt.Printf("Error sending response to next stage %s: %v\n", nextStage, err)
            nextStageAddrMutex.Unlock()
			continue
        }
		nextStageAddrMutex.Unlock()

		// Send ACK to previous stage
		conn.Write([]byte("ACK@" + request))

        fmt.Printf("Sent response to next stage: %s\n", nextStage)
	}

	for {
		nextStageAddrMutex.Lock()
		if len(ackMap) == 0 {
			nextStageAddrMutex.Unlock()
			break
		}
		nextStageAddrMutex.Unlock()
		time.Sleep(1 * time.Second)
	}

	// Send end of task message to all next stages
	endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
	nextStageAddrMutex.Lock()
	for _, nextStage := range nextStageAddrMap[ID] {
		nextConn, err := net.Dial("udp", nextStage)
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStage, err)
			continue
		}
		defer conn.Close()
		_, err = nextConn.Write([]byte(endMessage))
		if err != nil {
			fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStage, err)
			continue
		}
	}
	nextStageAddrMutex.Unlock()

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)
	logConn, err := net.Dial("udp", leaderAddr)
	if err != nil {
		fmt.Printf("Error connecting to leader %s: %v\n", leaderAddr, err)
		return
	}
	defer logConn.Close()

	logMessage := fmt.Sprintf("[Log] 2 %s", taskNo)
	_, err = logConn.Write([]byte(logMessage))
	if err != nil {
		fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
	}
}

func handleStage2Acks(ID string, ackMap map[string]int, ackedFilename string, taskNo string) {
	ackBuffer := make([]byte, 1024)
	for {
		nextStageAddrMutex.Lock()
		nextStages := nextStageAddrMap[ID]
		nextStageAddrMutex.Unlock()

		for _, address := range nextStages {
			conn, err := net.Dial("udp", address)
			if err != nil {
				fmt.Printf("Error connecting to next stage %s: %v\n", address, err)
				continue
			}

			timeoutDuration := 200 * time.Millisecond // Set a very short timeout duration
			conn.SetReadDeadline(time.Now().Add(timeoutDuration))

			n, err := conn.Read(ackBuffer)
			if err != nil {
				conn.Close()
				continue
			}
			conn.Close()

			ack := string(ackBuffer[:n])
			parts := strings.SplitN(ack, "@", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
				fmt.Printf("Invalid ACK received: %s\n", ack)
				continue
			}

			line := parts[1]

			nextStageAddrMutex.Lock()
			if _, exists := ackMap[line]; exists {

				file, err := os.Create(ackedFilename)
				if err != nil {
					fmt.Printf("Error creating file %s: %v\n", ackedFilename, err)
					nextStageAddrMutex.Unlock()
					continue
				}

				_, err = file.WriteString(line)
				if err != nil {
					fmt.Printf("Error writing to file %s: %v\n", ackedFilename, err)
					nextStageAddrMutex.Unlock()
					continue
				}
				file.Close()

				cmd := exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", ackedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
				err = cmd.Run()
				if err != nil {
					fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
					nextStageAddrMutex.Unlock()
					continue
				}

				delete(ackMap, line)
			}
			nextStageAddrMutex.Unlock()
		}
	}
}

func handleStage2resend(ID string, ackMap map[string]int) {
	for {
		nextStageAddrMutex.Lock()
		for line := range ackMap {
			hash := sha256.Sum256([]byte(line))
			hashValue := hex.EncodeToString(hash[:])
			nextStageIndex := int(hashValue[0]) % len(nextStageAddrMap[ID])
			nextStage := nextStageAddrMap[ID][nextStageIndex]

			conn, err := net.Dial("udp", nextStage)
			if err != nil {
				fmt.Printf("Error connecting to next stage %s: %v\n", nextStage, err)
				continue
			}

			_, err = conn.Write([]byte(fmt.Sprintf("%s\n", line)))
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				conn.Close()
				continue
			}
			conn.Close()
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

	// Log the received parameters
    fmt.Printf("Starting task server stage 3 on port %d with params: opFile=%s, stateful=%s, taskNo=%s, totalNum=%s, hydfsDestFilename=%s, recover=%s\n",
        port, opFile, stateful, taskNo, totalNum, hydfsDestFilename, recover)

	processedInput := make(map[string]int)
	processedFilename := fmt.Sprintf("%s/3_%s_PROC", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path
	state := make(map[string]int)
	stateFilename := fmt.Sprintf("%s/3_%s_STATE", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path

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

			cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_STATE", taskNo))
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
        n, _, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received request: %s\n", request)

		if request == "END_OF_TASK" {
			totalNum -= 1
			if totalNum == 0 {
				break
			}
		}

		if _, exists := processedInput[request]; exists {
			conn.Write([]byte("ACK@" + request))
			continue
		}

		// Run the external program with the request as input
        cmd := exec.Command("../ops/" + opFile)
        cmd.Stdin = strings.NewReader(request)
        output, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error running external program: %v\n", err)
            continue
        }

		outputStr := string(output)

		processedInput[request] = 1
		if stateful == "stateful" {
			state[outputStr] += 1
		}

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		_, err = file.WriteString(request + "\n")
        if err != nil {
            fmt.Printf("Error writing to file %s: %v\n", processedFilename, err)
            continue
        }
		file.Close()

		// Send the processed data to HyDFS
		cmd = exec.Command("go", "run", "mp3_client.go", "append", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
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
			cmd = exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
				continue
			}
		}

		conn.Write([]byte("ACK@" + request))

        fmt.Printf("Sent processed data to HyDFS: %s\n", hydfsDestFilename)
    }

	leaderAddr := fmt.Sprintf("%s:%s", leader, leaderPort)
	logConn, err := net.Dial("udp", leaderAddr)
	if err != nil {
		fmt.Printf("Error connecting to leader %s: %v\n", leaderAddr, err)
		return
	}
	defer logConn.Close()

	logMessage := fmt.Sprintf("[Log] 3 %s", taskNo)
	_, err = logConn.Write([]byte(logMessage))
	if err != nil {
		fmt.Printf("Error sending log message to leader %s: %v\n", leaderAddr, err)
	}
}