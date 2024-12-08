package main

import (
    "fmt"
    "net"
    "os"
)

var leader string = "fa24-cs425-6605.cs.illinois.edu"
var leaderPort string = "8090"

func main() {
    if len(os.Args) != 8 {
        fmt.Println("Usage: <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>")
        return
    }

    op1Exe := os.Args[1]
    op2Exe := os.Args[2]
    hydfsSrcFile := os.Args[3]
    hydfsDestFilename := os.Args[4]
    numTasks := os.Args[5]
    X := os.Args[6]
    stateful := os.Args[7]

    if stateful != "stateful" && stateful != "stateless" {
        fmt.Println("Error: <stateful> must be either 'stateful' or 'stateless'")
        return
    }

    message := fmt.Sprintf("%s %s %s %s %s %s %s", op1Exe, op2Exe, hydfsSrcFile, hydfsDestFilename, numTasks, X, stateful)
    fmt.Println("Command Message:", message)

    conn, err := net.Dial("udp", leader+":"+leaderPort)
    if err != nil {
        fmt.Println("Error connecting to leader:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Println("Error sending message:", err)
        return
    }

    buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	response := string(buffer[:n])
	fmt.Println("Response:", response)
}