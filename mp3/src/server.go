package main

import (
    "fmt"
    "hash/crc32"
    "net"
    "os"
    "mp2/src/global"
)

var LocalDir = "../data/"
var RingMod = 256
var ReplicationFactor = 3

var localFile = make(map[string][]string)
var cluster map[string]global.NodeInfo

var fileMutex sync.Mutex
var membershipMutex sync.Mutex
// TODO: add mutex in code



func main() {
    err := deleteAllFiles(LocalDir)
    if err != nil {
        fmt.Println("Error deleting files:", err)
        os.Exit(1)
    }

    membershipTicker := time.NewTicker(10 * time.Second)
    defer membershipTicker.Stop()

    fileTicker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    go func() {
        for range membershipTicker.C {
            updateMembershipList()
        }
    }()

    go func() {
        for range fileTicker.C {
            syncFiles()
        }
    }()

    go startServer()

    select {}
}

func deleteAllFiles(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()

    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }

    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
    return nil
}

func startServer() {
    listener, err := net.Listen("tcp", ":" + global.HDFSPort)
    if err != nil {
        fmt.Println("Error starting server:", err)
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Server started on port:", global.HDFSPort)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    fmt.Println("Client connected:", conn.RemoteAddr().String())

    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Println("Error reading from connection:", err)
            return
        }
        message := string(buffer[:n])
        fmt.Println("Received message:", message)

        // Handle the received message
        parts := strings.Fields(message)
        if len(parts) < 1 {
            fmt.Println("Invalid command")
            return
        }
    
        command := parts[0]
        var filename string
        if len(parts) > 1 {
            filename = parts[1]
        }

        switch command {
        case "create":
            handleCreate(conn, filename)
        case "append":
            handleAppend(conn, filename)
        case "get":
            handleGet(conn, filename)
        case "ls":
            handleList(conn)
        // TODO: might need to handle update for sync here
        default:
            fmt.Println("Unknown command")
        }
    }
}

func handleCreate(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    // Check if the file already exists
    if _, err := os.Stat(filePath); err == nil {
        // File exists
        _, err := conn.Write([]byte("Fail: File already exists\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        return
    }

    // File does not exist, create it
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Receive the file content and write it to the file
    file, err := os.Create(filePath)
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer file.Close()

    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading from connection:", err)
            return
        }
        if n == 0 {
            break
        }
        _, err = file.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }
    }

    fmt.Printf("File %s created successfully\n", filename)
}

func handleAppend(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    // Check if the file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        // File does not exist
        _, err := conn.Write([]byte("Fail: File does not exist\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        return
    }

    // File exists, ready to append
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Receive the file content and append it to the file
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading from connection:", err)
            return
        }
        if n == 0 {
            break
        }
        _, err = file.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }
    }

    fmt.Printf("File %s appended successfully\n", filename)
}

// handleGet handles the get command
func handleGet(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    // Check if the file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        // File does not exist
        _, err := conn.Write([]byte("Fail: File does not exist\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        return
    }

    // File exists, read and return the content
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    buffer := make([]byte, 1024)
    for {
        n, err := file.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading from file:", err)
            return
        }
        if n == 0 {
            break
        }
        _, err = conn.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    }

    fmt.Printf("File %s sent successfully\n", filename)
}

func handleList(conn net.Conn) {
    for filename, _ := range localFile {
        _, err := conn.Write([]byte(filename + "\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    }
}

func updateMembershipList() {
    response := global.GetMembership()

    if response == cluster {
        return
    }
    cluster = response

    for file, _ := range localFile {
        replicas := global.FindFileReplicas(file)
        localFile[file] = replicas
    }
}

func syncFiles() {
    // TODO: periodically sync files accross all replicas, sync if this node is primary replica of the file
}