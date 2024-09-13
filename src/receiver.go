package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "strings"
)

func main() {
    // Ensure the filename suffix is provided as a command-line argument
    if len(os.Args) < 2 {
        fmt.Println("Usage: receiver <filename_suffix>")
        os.Exit(1)
    }
    filenameSuffix := os.Args[1]

    // Start listening for connections on port 8080
    err := startServer("8080", filenameSuffix)
    if err != nil {
        fmt.Println("Error starting server:", err)
        os.Exit(1)
    }
}

// startServer starts a TCP server on the specified port
func startServer(port, filenameSuffix string) error {
    // Listen for incoming connections
    listener, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return fmt.Errorf("error starting TCP listener: %v", err)
    }
    defer listener.Close()

    fmt.Printf("Server is listening on port %s\n", port)

    // Accept connections in a loop
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting connection: %v\n", err)
            continue
        }

        // Handle the connection in a new goroutine
        go handleConnection(conn, filenameSuffix)
    }
}

// handleConnection handles an individual connection
func handleConnection(conn net.Conn, filenameSuffix string) {
    defer conn.Close()

    // Create a buffered reader to read data from the connection
    reader := bufio.NewReader(conn)

    // Read data until '\x00' is encountered
    var data strings.Builder
    for {
        b, err := reader.ReadByte()
        if err != nil {
            fmt.Printf("Error reading from connection: %v\n", err)
            return
        }
        if b == '\x00' {
            break
        }
        data.WriteByte(b)
    }

    // Print the received data
    fmt.Printf("Received data: %s\n", data.String())

    // Construct the file name using the provided filename suffix
    fileName := fmt.Sprintf("vm%s.log", filenameSuffix)

    // Split the received data into grep options
    grepOptions := strings.Split(data.String(), "\n")

    // Prepare the arguments for the grep command
    args := append(grepOptions, fileName)

    // Execute the grep command
    cmd := exec.Command("grep", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error executing grep: %v\n", err)
    }

    // Send the grep output back to the client
    response := string(output) + "\x00"
    _, err = conn.Write([]byte(response))
    if err != nil {
        fmt.Printf("Error sending response: %v\n", err)
    }
}