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
)

var workerPort = 8091
var portMutex sync.Mutex
var nextPort = 9000
var leader = "fa24-cs425-6605.cs.illinois.edu"
var leaderPort = "8090"


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

        // Schedule the task
        port := processTaskRequest(message)
        var response string
        if port != "" {
            response = fmt.Sprintf("Success, port=%s", port)
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
    parts := strings.Split(message, ",")

    taskType := strings.TrimSpace(parts[0])

    portMutex.Lock()
    port := nextPort
    nextPort++
    portMutex.Unlock()

	logFile = "../../mp1/data/worker.log"
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFile, err)
		return ""
	}
	
	log.WriteString(fmt.Sprintf("Task %s scheduled on port %d\n", taskType, port))
	log.Close()

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
    totalNum := strings.TrimSpace(params[2])
    nextStage := strings.TrimSpace(params[3])
    recover := strings.TrimSpace(params[4])

    localFilename := fmt.Sprintf("%s/1_%s", os.Getenv("HOME"), taskNo)
    cmd := exec.Command("go", "run", "../../mp3/src/client.go", "get", "--localfilename", localFilename, "--HyDFSfilename", hydfsSrcFile)

    err := cmd.Run()
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

    conn, err := net.Dial("udp", nextStage)
    if err != nil {
        fmt.Printf("Error connecting to next stage %s: %v\n", nextStage, err)
        return
    }
    defer conn.Close()

	ackMap := make(map[string]int)
    var ackMapMutex sync.Mutex

    go handleStage1Acks(conn, ackMap, &ackMapMutex)

	go handleStage1resend(conn, ackMap, &ackMapMutex)

    scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading file %s: %v\n", localFilename, err)
        return
    }

    for scanner.Scan() {
        line := scanner.Text()
        _, err = conn.Write([]byte(line))
        if err != nil {
            fmt.Printf("Error sending line to next stage %s: %v\n", nextStage, err)
            return
        }
    }

	for {
        ackMapMutex.Lock()
        if len(ackMap) == 0 {
            ackMapMutex.Unlock()
            break
        }
        ackMapMutex.Unlock()
        time.Sleep(1 * time.Second)
    }

	// Send end of task message to next stage
    endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
    _, err = conn.Write([]byte(endMessage))
    if err != nil {
        fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStageAddr, err)
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

func handleStage1Acks(conn net.Conn, ackMap map[string]int, ackMapMutex *sync.Mutex) {
    ackBuffer := make([]byte, 1024)
    for {
        n, err := conn.Read(ackBuffer)
        if err != nil {
            fmt.Printf("Error receiving ACK: %v\n", err)
            return
        }

        ack := string(ackBuffer[:n])
        parts := strings.SplitN(ack, "$", 2)
        if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
            fmt.Printf("Invalid ACK received: %s\n", ack)
            continue
        }

        line := parts[1]

        ackMapMutex.Lock()
        if count, exists := ackMap[line]; exists {
            delete(ackMap, line)
        }
        ackMapMutex.Unlock()
    }
}

func handleStage1resend(conn net.Conn, ackMap map[string]int, ackMapMutex *sync.Mutex) {
	for {
		ackMapMutex.Lock()
		for line := range ackMap {
			_, err := conn.Write([]byte(line))
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				return
			}
		}
		ackMapMutex.Unlock()
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

    // Log the received parameters
    fmt.Printf("Starting task server stage 2 on port %d with params: opFile=%s, taskNo=%s, totalNum=%s, nextStageList=%v, recover=%s\n",
        port, opFile, taskNo, totalNum, nextStageList, recover)

	processInput := make(map[string]int)
	// Create the file
    processedFilename := fmt.Sprintf("%s/2_%s_PROC", os.Getenv("HOME"), taskNo)
    file, err := os.Create(processedFilename)
    if err != nil {
        fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
        return
    }
    file.Close()

	// Run the command to create the file in HyDFS
    cmd := exec.Command("go", "run", "../../mp3/src/client.go", "create", "--localfilename", localFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
        return
    }

	// Create the file
    ackedFilename := fmt.Sprintf("%s/2_%s_ACKED", os.Getenv("HOME"), taskNo)
    file, err := os.Create(localFilename)
    if err != nil {
        fmt.Printf("Error creating file %s: %v\n", ackedFilename, err)
        return
    }
    file.Close()

	// Run the command to create the file in HyDFS
    cmd := exec.Command("go", "run", "../../mp3/src/client.go", "create", "--localfilename", localFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
        return
    }
	
	ackMap := make(map[string]int)
    var ackMapMutex sync.Mutex

	if recover == "true" {
		// Get the processed data from HyDFS
		cmd := exec.Command("go", "run", "../../mp3/src/client.go", "get", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
		err = cmd.Run()
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
		isInput := true
		for scanner.Scan() {
			line := scanner.Text()
			if isInput {
				processInput[line]++
			} else {
				ackMap[line]++
			}
			isInput = !isInput
		}

		file.Close()
		
		// Get the acked data from HyDFS
		cmd := exec.Command("go", "run", "../../mp3/src/client.go", "get", "--localfilename", ackedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_ACKED", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to get file from HyDFS: %v\n", err)
			return
		}

		// load the acked data into the ackMap
		file, err := os.Open(ackedFilename)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", ackedFilename, err)
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			delete(ackMap, line)
		}

		file.Close()
	}

	// Create a map to store UDP connections for each next stage address
	nextStageConns := make(map[string]*net.UDPConn)

	for _, nextStage := range nextStageList {
		nextStageAddr := fmt.Sprintf("%s:%d", nextStage, port)
		udpAddr, err := net.ResolveUDPAddr("udp", nextStageAddr)
		if err != nil {
			fmt.Printf("Error resolving address %s: %v\n", nextStageAddr, err)
			continue
		}

		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			fmt.Printf("Error connecting to next stage %s: %v\n", nextStageAddr, err)
			continue
		}

		nextStageConns[nextStage] = conn

		defer conn.Close()
	}

    go handleStage2Acks(nextStageConns, ackMap, &ackMapMutex, ackedFilename)

	go handleStage2resend(nextStageConns, ackMap, &ackMapMutex)

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
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received request: %s\n", request)

		if request == "END_OF_TASK" {
			break
		}

		if exists := processInput[request]; exists {
			conn.Write("ACK$%s\n", request)
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

		if string(output) == "0\n" {
			continue
		}

		++processInput[request]

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		_, err = file.WriteString(fmt.Sprintf("%s%s", request, response))
        if err != nil {
            fmt.Printf("Error writing to file %s: %v\n", processedFilename, err)
            continue
        }
		file.Close()

		// Send the processed data to HyDFS
		cmd = exec.Command("go", "run", "../../mp3/src/client.go", "put", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_PROC", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
			continue
		}

		// Hash the response to determine the next stage
        hash := sha256.Sum256([]byte(response))
        hashValue := hex.EncodeToString(hash[:])
        nextStageIndex := int(hashValue[0]) % len(nextStageList)
        nextStage := nextStageList[nextStageIndex]

		_, err = nextStageConns[nextStage].Write([]byte(fmt.Sprintf("%s\n", response)))
        if err != nil {
            fmt.Printf("Error sending response to next stage %s: %v\n", nextStageAddr, err)
            continue
        }

		// Send ACK to previous stage
		conn.Write("ACK$%s\n", request)

        fmt.Printf("Sent response to next stage %s\n", nextStageAddr)
	}

	for {
		ackMapMutex.Lock()
		if len(ackMap) == 0 {
			ackMapMutex.Unlock()
			break
		}
		ackMapMutex.Unlock()
		time.Sleep(1 * time.Second)
	}

	// Send end of task message to all next stage
	endMessage := fmt.Sprintf("END_OF_TASK %s", taskNo)
	for _, nextStage := range nextStageList {
		_, err = nextStageConns[nextStage].Write([]byte(endMessage))
		if err != nil {
			fmt.Printf("Error sending end of task message to next stage %s: %v\n", nextStage, err)
			return
		}
	}

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

func handleStage2Acks(nextStageConns map[string]*net.UDPConn, ackMap map[string]int, ackMapMutex *sync.Mutex, ackedFilename string) {
	ackBuffer := make([]byte, 1024)
	for {
		for _, conn := range nextStageConns {
			n, err := conn.Read(ackBuffer)
			if err != nil {
				fmt.Printf("Error receiving ACK: %v\n", err)
				continue
			}

			ack := string(ackBuffer[:n])
			parts := strings.SplitN(ack, "$", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "ACK" {
				fmt.Printf("Invalid ACK received: %s\n", ack)
				continue
			}

			line := parts[1]

			ackMapMutex.Lock()
			if count, exists := ackMap[line]; exists {

				file, err := os.Create(ackedFilename)
				if err != nil {
					fmt.Printf("Error creating file %s: %v\n", ackedFilename, err)
					continue
				}

				_, err = file.WriteString(line)
				if err != nil {
					fmt.Printf("Error writing to file %s: %v\n", ackedFilename, err)
					continue
				}
				file.Close()

				delete(ackMap, line)
			}
			ackMapMutex.Unlock()
		}
	}
}

func handleStage2resend(nextStageConns map[string]*net.UDPConn, ackMap map[string]int, ackMapMutex *sync.Mutex) {
	for {
		ackMapMutex.Lock()
		for line := range ackMap {
			hash := sha256.Sum256([]byte(line))
			hashValue := hex.EncodeToString(hash[:])
			nextStageIndex := int(hashValue[0]) % len(nextStageList)
			nextStage := nextStageList[nextStageIndex]

			_, err := nextStageConns[nextStage].Write([]byte(fmt.Sprintf("%s\n", line)))
			if err != nil {
				fmt.Printf("Error resending line: %v\n", err)
				continue
			}
		}
		ackMapMutex.Unlock()
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
    totalNum := strings.TrimSpace(params[3])
    hydfsDestFilename := strings.TrimSpace(params[4])
    recover := strings.TrimSpace(params[5])

	// Log the received parameters
    fmt.Printf("Starting task server stage 3 on port %d with params: opFile=%s, stateful=%s, taskNo=%s, totalNum=%s, hydfsDestFilename=%s, recover=%s\n",
        port, opFile, stateful, taskNo, totalNum, hydfsDestFilename, recover)

	processedInput := make(map[string]int)
    processedFilename := fmt.Sprintf("%s/3_%s_PROC", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path
    file, err := os.Create(processedFilename)
    if err != nil {
        fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
        return
    }
    file.Close()

	cmd := exec.Command("go", "run", "../../mp3/src/client.go", "create", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
        return
    }

	state := make(map[string]string)
	stateFilename := fmt.Sprintf("%s/3_%s_STATE", os.Getenv("HOME"), taskNo) // Use ~/ as the start of the file path
    if state == "stateful" {		
		file, err := os.Create(stateFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", stateFilename, err)
			return
		}
		file.Close()

		cmd := exec.Command("go", "run", "../../mp3/src/client.go", "create", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("2_%s_STATE", taskNo))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command to create file in HyDFS: %v\n", err)
			return
		}
	}

	if recover == "true" {
		// Get the processed data from HyDFS
		cmd := exec.Command("go", "run", "../../mp3/src/client.go", "get", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
		err = cmd.Run()
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
			cmd := exec.Command("go", "run", "../../mp3/src/client.go", "get", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
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
			for scanner.Scan() {
				line := scanner.Text()
				if isKey {
					key := line
				} else {
					value := line
					state[key] = value
				}
				isKey = !isKey
			}

			file.Close()
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
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP:", err)
            continue
        }

        request := string(buffer[:n])
        fmt.Printf("Received request: %s\n", request)

		if request == "END_OF_TASK" {
			--totalNum
			if totalNum == 0 {
				break
			}
		}

		if exists := processedInput[request]; exists {
			conn.Write("ACK$%s\n", request)
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

		++processedInput[request]
		if stateful == "stateful" {
			++state[output]
		}

		file, err := os.Create(processedFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", processedFilename, err)
			continue
		}

		_, err = file.WriteString(request)
        if err != nil {
            fmt.Printf("Error writing to file %s: %v\n", processedFilename, err)
            continue
        }
		file.Close()

		// Send the processed data to HyDFS
		cmd = exec.Command("go", "run", "../../mp3/src/client.go", "put", "--localfilename", processedFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_PROC", taskNo))
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
			cmd = exec.Command("go", "run", "../../mp3/src/client.go", "put", "--localfilename", stateFilename, "--HyDFSfilename", fmt.Sprintf("3_%s_STATE", taskNo))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command to put file in HyDFS: %v\n", err)
				continue
			}
		}

		conn.Write("ACK$%s\n", request)

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