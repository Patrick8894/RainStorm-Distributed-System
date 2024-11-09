package main

import (
    "fmt"
    "net"
    "os"
    "mp3/src/global"
    "sync" 
    "time" 
    "path/filepath"
    "strings"
    "io"
)

var LocalDir = "../data/"
var SelfAddress string
// localfile maps a filename to a list of replicas' ip e.x. localfile["file1"] = ["ip1", "ip2", "ip3"]
var localFile = make(map[string][]string)
// 
var cachedFile = make(map[string][]byte)

var localFileMutex sync.Mutex
var cachedFileMutex sync.Mutex
var diskMutex sync.Mutex

func main() {
    // Before starting the server, delete all files in the local directory
    err := deleteAllFiles(LocalDir)
    if err != nil {
        fmt.Println("Error deleting files:", err)
        os.Exit(1)
    }

    hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return
    }

    SelfAddress = hostname + ":" + global.HDFSPort

    membershipTicker := time.NewTicker(10 * time.Second)
    defer membershipTicker.Stop()

    fileTicker := time.NewTicker(30 * time.Second)
    defer fileTicker.Stop()

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
        case "update":
            handleUpdate(conn, filename)
        case "sync":
            handleSync(conn, filename)
        default:
            fmt.Println("Unknown command")
        }
    }
}

func handleCreate(conn net.Conn, filename string) {
    /*
    Create a new file with the given filename and write the content to it from Client.
    */
    filePath := LocalDir + filename

    // Check if the file already exists
    localFileMutex.Lock()
    if _, exists := localFile[filename]; exists {
        // File exists
        _, err := conn.Write([]byte("Fail: File already exists\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        localFileMutex.Unlock()
        return
    }
    localFileMutex.Unlock()

    // File does not exist, create it
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Receive the file content and write it to the file
    diskMutex.Lock()
    defer diskMutex.Unlock()
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
            continue
        }
        _, err = file.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }
    }

    localFileMutex.Lock()
    localFile[filename] = global.FindFileReplicas(filename)
    localFileMutex.Unlock()

    fmt.Printf("File %s created successfully\n", filename)
}

func handleAppend(conn net.Conn, filename string) {

    localFileMutex.Lock()
    // Check if the file exists
    if _, exists := localFile[filename]; !exists {
        // File does not exist
        _, err := conn.Write([]byte("Fail: File does not exist\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        localFileMutex.Unlock()
        return
    }
    localFileMutex.Unlock()

    // File exists, ready to append
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Receive the file content and append it to cached file

    cachedFileMutex.Lock()
    defer cachedFileMutex.Unlock()
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
            continue
        }

        // Append received data to the cached file content
        cachedFile[filename] = append(cachedFile[filename], buffer[:n]...)
    }

    fmt.Printf("Received data for %s. Cached file size: %d bytes\n", filename, len(cachedFile[filename]))
}

func handleGet(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    localFileMutex.Lock()
    // Check if the file exists
    if _, exists := localFile[filename]; !exists {
        // File does not exist
        _, err := conn.Write([]byte("Fail: File does not exist\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        localFileMutex.Unlock()
        return
    }
    localFileMutex.Unlock()

    // File exists, ready to append
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Check if the file exists on disk
    diskMutex.Lock()
    if _, err := os.Stat(filePath); err == nil {
        // File exists, read and return the content
        file, err := os.Open(filePath)
        if err != nil {
            fmt.Println("Error opening file:", err)
            diskMutex.Unlock()
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
                diskMutex.Unlock()
                return
            }
            if n == 0 {
                break
            }
            _, err = conn.Write(buffer[:n])
            if err != nil {
                fmt.Println("Error writing to connection:", err)
                diskMutex.Unlock()
                return
            }
        }

        fmt.Printf("File %s sent successfully from disk\n", filename)
    } else {
        fmt.Printf("File %s does not exist on disk\n", filename)
    }
    diskMutex.Unlock()

    // Check if additional cached content exists for the file
    cachedFileMutex.Lock()
    defer cachedFileMutex.Unlock()
    if content, exists := cachedFile[filename]; exists {
        // Send cached content in chunks
        bufferSize := 1024
        for start := 0; start < len(content); start += bufferSize {
            end := start + bufferSize
            if end > len(content) {
                end = len(content)
            }

            // Send the slice of content in the range [start:end]
            _, err := conn.Write(content[start:end])
            if err != nil {
                fmt.Println("Error writing cached content to connection:", err)
                return
            }
        }
        fmt.Printf("Cached content for file %s sent successfully\n", filename)
    } else {
        fmt.Printf("No cached content to send for file %s\n", filename)
    }
}

func handleList(conn net.Conn) {
    localFileMutex.Lock()
    defer localFileMutex.Unlock()

    var files []string
    for file, _ := range localFile {
        files = append(files, file)
    }

    response := strings.Join(files, "\n")
    _, err := conn.Write([]byte(response))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }
}


func handleUpdate(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    // Check if the file already exists
    localFileMutex.Lock()
    if _, exists := localFile[filename]; exists {
        // File exists
        _, err := conn.Write([]byte("Fail: File already exists\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
        }
        localFileMutex.Unlock()
        return
    }

    localFile[filename] = global.FindFileReplicas(filename)
    localFileMutex.Unlock()

    // File does not exist, create it
    _, err := conn.Write([]byte("Success: Ready to receive file content\n"))
    if err != nil {
        fmt.Println("Error writing to connection:", err)
        return
    }

    // Receive the file content and write it to the file
    diskMutex.Lock()
    defer diskMutex.Unlock()

    file, err := os.Create(filePath)
    if err != nil {
        fmt.Println("Error creating file:", err)
        diskMutex.Unlock()
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
            continue
        }
        _, err = file.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }
    }

    fmt.Printf("File %s updated successfully\n", filename)
}

func handleSync(conn net.Conn, filename string) {
    filePath := LocalDir + filename

    diskMutex.Lock()
    fileExists := true
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        fileExists = false
    }
    diskMutex.Unlock()

    // Send back response to inform sender whether the file exists
    if fileExists {
        _, err := conn.Write([]byte("File exists\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    } else {       
        localFileMutex.Lock()
        localFile[filename] = global.FindFileReplicas(filename)
        localFileMutex.Unlock()

        _, err := conn.Write([]byte("File does not exist\n"))
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    }

    // Open the file in append mode, create it if it does not exist
    diskMutex.Lock()
    defer diskMutex.Unlock()
    
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // Receive the file content and append it to the file
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
            continue
        }
        _, err = file.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }
    }

    // Remove the entry from cachedFile
    cachedFileMutex.Lock()
    delete(cachedFile, filename)
    cachedFileMutex.Unlock()

    fmt.Printf("File %s synchronized and cached entry removed\n", filename)
}

func updateMembershipList() {
    localFileMutex.Lock()
    response := global.GetMembership()

    if mapsEqual(response, global.Cluster) {
        return
    }
    global.Cluster = response

    for file, _ := range localFile {
        replicas := global.FindFileReplicas(file)
        localFile[file] = replicas
    }
    localFileMutex.Unlock()
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

func syncFiles() {
    localFileMutex.Lock()
    defer localFileMutex.Unlock()

    for filename, replicas := range localFile {
        if len(replicas) == 0 {
            continue
        }

        // Check if this node is the primary replica
        if replicas[0] == SelfAddress {

            // if _, exists := localFile[filename]; !exists {
            //     fmt.Printf("File %s does not exist locally\n", filename)
            //     continue
            // }

            for _, replica := range replicas[1:] {
                syncReplicaFile(filename, replica)
            }
        
            // Append cached content to disk and delete the entry
            filePath := LocalDir + filename
        
            // Check if additional cached content exists for the file
            cachedFileMutex.Lock()
            defer cachedFileMutex.Unlock()

            if content, exists := cachedFile[filename]; exists {

                diskMutex.Lock()
                defer diskMutex.Unlock()
                // Open the file in append mode, create it if it does not exist
                file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                if err != nil {
                    fmt.Println("Error opening file:", err)
                    return
                }
                defer file.Close()

                // Append cached content to the file
                bufferSize := 1024
                for start := 0; start < len(content); start += bufferSize {
                    end := start + bufferSize
                    if end > len(content) {
                        end = len(content)
                    }
        
                    // Write the slice of content in the range [start:end]
                    _, err := file.Write(content[start:end])
                    if err != nil {
                        fmt.Println("Error writing cached content to file:", err)
                        return
                    }
                }
                fmt.Printf("Cached content for file %s appended to disk successfully\n", filename)
        
                // Remove the entry from cachedFile
                delete(cachedFile, filename)
            } else {
                fmt.Printf("No cached content to append for file %s\n", filename)
            }
        } else {
            // Check if this node is a replica
            isReplica := false
            for _, replica := range replicas {
                if replica == SelfAddress {
                    isReplica = true
                    break
                }
            }

            // If this node is not a replica, handle accordingly
            if !isReplica {
                sendPrimaryReplica(filename, replicas[0])
            }
        }
    }
}

func syncReplicaFile(filename string, replica string) {

    filePath := LocalDir + filename

    conn, err := net.Dial("tcp", replica)
    if err != nil {
        fmt.Println("Error connecting to replica:", err)
        return
    }
    defer conn.Close()

    command := fmt.Sprintf("sync %s", filename)
    _, err = conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return
    }

    // Read the response from the replica
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading response from replica:", err)
        return
    }
    response := string(buffer[:n])

    if response == "File does not exist\n" {
        // Send file content on disk if it exists
        diskMutex.Lock()
        if _, err := os.Stat(filePath); err == nil {
            // File exists, read and return the content
            file, err := os.Open(filePath)
            if err != nil {
                fmt.Println("Error opening file:", err)
                diskMutex.Unlock()
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
                    diskMutex.Unlock()
                    return
                }
                if n == 0 {
                    break
                }
                _, err = conn.Write(buffer[:n])
                if err != nil {
                    fmt.Println("Error writing to connection:", err)
                    diskMutex.Unlock()
                    return
                }
            }

            fmt.Printf("File %s sent successfully from disk\n", filename)
        } else {
            fmt.Printf("File %s does not exist on disk\n", filename)
        }
        diskMutex.Unlock()
    }

    // Check if additional cached content exists for the file
    cachedFileMutex.Lock()
    defer cachedFileMutex.Unlock()
    if content, exists := cachedFile[filename]; exists {
        // Send cached content in chunks
        bufferSize := 1024
        for start := 0; start < len(content); start += bufferSize {
            end := start + bufferSize
            if end > len(content) {
                end = len(content)
            }

            // Send the slice of content in the range [start:end]
            _, err := conn.Write(content[start:end])
            if err != nil {
                fmt.Println("Error writing cached content to connection:", err)
                return
            }
        }
        fmt.Printf("Cached content for file %s sent successfully\n", filename)
    } else {
        fmt.Printf("No cached content to send for file %s\n", filename)
    }
}

// SendPrimaryReplica sends the file content to the primary replica before deleting the file
func sendPrimaryReplica(filename string, primaryReplica string) {

    conn, err := net.Dial("tcp", primaryReplica)
    if err != nil {
        fmt.Println("Error connecting to primary replica:", err)
        return
    }
    defer conn.Close()

    command := fmt.Sprintf("update %s", filename)
    _, err = conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return
    }

    // Read the response from the primary replica
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading response from primary replica:", err)
        return
    }
    response := string(buffer[:n])
    if strings.HasPrefix(response, "Fail") {
        fmt.Println("Primary replica returned an error:", response)
        return
    }

    // Open the file to read its content
    filePath := LocalDir + filename
    diskMutex.Lock()
    if _, err := os.Stat(filePath); err == nil {
        // File exists, read and return the content
        file, err := os.Open(filePath)
        if err != nil {
            fmt.Println("Error opening file:", err)
            diskMutex.Unlock()
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
                diskMutex.Unlock()
                return
            }
            if n == 0 {
                break
            }
            _, err = conn.Write(buffer[:n])
            if err != nil {
                fmt.Println("Error writing to connection:", err)
                diskMutex.Unlock()
                return
            }
        }

        fmt.Printf("File %s sent successfully from disk\n", filename)

        // Delete the file before releasing the lock
        err = os.Remove(filePath)
        if err != nil {
            fmt.Println("Error deleting file:", err)
        } else {
            fmt.Printf("File %s deleted successfully\n", filename)
        }
    } else {
        fmt.Printf("File %s does not exist on disk\n", filename)
    }

    diskMutex.Unlock()


    // Check if additional cached content exists for the file
    cachedFileMutex.Lock()
    defer cachedFileMutex.Unlock()
    if content, exists := cachedFile[filename]; exists {
        // Send cached content in chunks
        bufferSize := 1024
        for start := 0; start < len(content); start += bufferSize {
            end := start + bufferSize
            if end > len(content) {
                end = len(content)
            }

            // Send the slice of content in the range [start:end]
            _, err := conn.Write(content[start:end])
            if err != nil {
                fmt.Println("Error writing cached content to connection:", err)
                return
            }
        }
        fmt.Printf("Cached content for file %s sent successfully\n", filename)
    } else {
        fmt.Printf("No cached content to send for file %s\n", filename)
    }
    delete(localFile, filename)
}