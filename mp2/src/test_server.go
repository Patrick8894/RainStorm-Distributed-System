package main

import (
    "fmt"
    "log"
    "net"
    pb "mp2/proto" // Update this to your correct package path
    "google.golang.org/protobuf/proto"
)

func handleConnection(conn net.Conn) {
    defer conn.Close()

    // Create a buffer to hold incoming data
    buf := make([]byte, 1024)
    
    // Read data from the connection
    n, err := conn.Read(buf)
    if err != nil {
        log.Println("Error reading from connection:", err)
        return
    }

    // Unmarshal the protobuf message
    var message pb.SWIMMessage
    err = proto.Unmarshal(buf[:n], &message)
    if err != nil {
        log.Println("Failed to unmarshal message:", err)
        return
    }

    // Process the received message
    fmt.Printf("Received message: %+v\n", message)
}

func main() {
    // Start listening on port 9000
    listener, err := net.Listen("tcp", ":9000")
    if err != nil {
        log.Fatalf("Failed to listen on port 9000: %v", err)
    }
    defer listener.Close()

    log.Println("Server is listening on port 9000")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println("Failed to accept connection:", err)
            continue
        }
        go handleConnection(conn) // Handle the connection in a new goroutine
    }
}
