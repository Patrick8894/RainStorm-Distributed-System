package main

import (
    "flag"
    "fmt"
    "os"
    "sync" 
    "net"
    "io"
    "strings"
    "mp3/src/global"
    "mp3/src/cache"
)

// IMPORTANT: For HyDFS commands, pass arguments with with format "command filename" in TCP

func main() {

    // // Ensure cache directory exists
    // os.MkdirAll("cache", os.ModePerm)

    // Define the command-line arguments
    createCmd := flag.NewFlagSet("create", flag.ExitOnError)
    getCmd := flag.NewFlagSet("get", flag.ExitOnError)
    appendCmd := flag.NewFlagSet("append", flag.ExitOnError)
    lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
    storeCmd := flag.NewFlagSet("store", flag.ExitOnError)
    getFromReplicaCmd := flag.NewFlagSet("getfromreplica", flag.ExitOnError)
    listMemIdsCmd := flag.NewFlagSet("list_mem_ids", flag.ExitOnError)

    // Define the arguments for each command
    createLocalFile := createCmd.String("localfilename", "", "Local file name")
    createHyDFSFile := createCmd.String("HyDFSfilename", "", "HyDFS file name")

    getHyDFSFile := getCmd.String("HyDFSfilename", "", "HyDFS file name")
    getLocalFile := getCmd.String("localfilename", "", "Local file name")

    appendLocalFile := appendCmd.String("localfilename", "", "Local file name")
    appendHyDFSFile := appendCmd.String("HyDFSfilename", "", "HyDFS file name")

    lsHyDFSFile := lsCmd.String("HyDFSfilename", "", "HyDFS file name")

    getFromReplicaVM := getFromReplicaCmd.String("VMaddress", "", "VM address")
    getFromReplicaHyDFSFile := getFromReplicaCmd.String("HyDFSfilename", "", "HyDFS file name")
    getFromReplicaLocalFile := getFromReplicaCmd.String("localfilename", "", "Local file name")

    // Parse the command-line arguments
    if len(os.Args) < 2 {
        fmt.Println("Expected 'create', 'get', 'append', 'ls', 'store', 'getfromreplica', or 'list_mem_ids' subcommands")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "create":
        createCmd.Parse(os.Args[2:])
        fmt.Printf("Creating file %s in HyDFS as %s\n", *createLocalFile, *createHyDFSFile)
        createFile(*createLocalFile, *createHyDFSFile)
        
    case "get":
        getCmd.Parse(os.Args[2:])
        fmt.Printf("Getting file %s from HyDFS and saving as %s\n", *getHyDFSFile, *getLocalFile)
        getFile(*getHyDFSFile, *getLocalFile)
    case "append":
        appendCmd.Parse(os.Args[2:])
        fmt.Printf("Appending file %s to HyDFS file %s\n", *appendLocalFile, *appendHyDFSFile)
        appendFile(*appendLocalFile, *appendHyDFSFile)
    case "ls":
        lsCmd.Parse(os.Args[2:])
        fmt.Printf("Listing all machine address that store the given HyDFS file %s\n", *lsHyDFSFile)
        listMachine(*lsHyDFSFile)
    case "store":
        storeCmd.Parse(os.Args[2:])
        fmt.Println("Listing all files stored on Local machine")
        listFiles()
    case "getfromreplica":
        getFromReplicaCmd.Parse(os.Args[2:])
        fmt.Printf("Getting file %s from replica at %s and saving as %s\n", *getFromReplicaHyDFSFile, *getFromReplicaVM, *getFromReplicaLocalFile)
        getFileFromReplica(*getFromReplicaVM, *getFromReplicaHyDFSFile, *getFromReplicaLocalFile)
    case "list_mem_ids":
        listMemIdsCmd.Parse(os.Args[2:])
        fmt.Println("Listing all member IDs")
        listMemberIds()
    default:
        fmt.Println("Expected 'create', 'get', 'append', 'ls', 'store', 'getfromreplica', or 'list_mem_ids' subcommands")
        os.Exit(1)
    }
}

func createFile(localfilename string, HyDFSfilename string) {
    /*
    Find the candidate server to create the HyDFS file and
    send the local file to the candidate server.
    */
    // TODO: Implement the create functionality here
    // check if localfilename exists
    _, err := os.Stat("data/"+localfilename)
    if err != nil {
        fmt.Println("Local file does not exist")
        return
    }

    // get the HyDFSfilename id
    candidates := global.FindFileReplicas(HyDFSfilename)
    if len(candidates) == 0 {
        fmt.Println("No candidates found")
        return
    }


    var wg sync.WaitGroup
    for _, candidate := range candidates {
        wg.Add(1)
        // use go routine to send the file to the candidate
        go func(candidate string){
            defer wg.Done()
            createFileToCandidate(candidate, localfilename, HyDFSfilename)
        }(candidate)
    }

    wg.Wait()

    // delete the cache entry
    cache.CacheMutex.Lock()
    cache.DeleteCacheEntry(HyDFSfilename)
    cache.CacheMutex.Unlock()

}

func getFile(HyDFSfilename string, localfilename string) {
    /*
    Find server in the candidates 0 index to get the HyDFS file and
    save the file to the local machine.
    */
    // TODO: Implement the get functionality here

    // First check if the file exists in the cache
    entry := cache.Cache[HyDFSfilename]

    if entry != nil {
        // Check if the cached file is up-to-date
        if cache.CheckCacheValidity(entry.Value.(*cache.CacheEntry).LastModified) { 
            // Cache is up-to-date, copy the file to the local machine
            err := os.WriteFile(localfilename, entry.Value.(*cache.CacheEntry).Data, 0644)
            if err != nil {
                fmt.Println("Error writing to local file:", err)
                return
            }
        } else {
            // Cache is not up-to-date, delete the cache entry
            cache.DeleteCacheEntry(HyDFSfilename)
        }
    }
        

    // get the candidate server to get the HyDFS file
    candidates := global.FindFileReplicas(HyDFSfilename)
    if len(candidates) == 0 {
        fmt.Println("No candidates found")
        return
    }

    // get the first candidate to get the file
    getFileFromReplica(candidates[0], HyDFSfilename, localfilename)

    // save the local file content to the cache
    cache.AddToCache(HyDFSfilename, localfilename)
}

func appendFile(localfilename string, HyDFSfilename string) {
    // TODO: Implement the append functionality here

    // check if localfilename exists
    _, err := os.Stat("data/"+localfilename)
    if err != nil {
        fmt.Println("Local file does not exist")
        return
    }

    // get the HyDFSfilename ip
    candidates := global.FindFileReplicas(HyDFSfilename)
    if len(candidates) == 0 {
        fmt.Println("No candidates found")
        return
    }

    var wg sync.WaitGroup
    responses := make(chan string, len(candidates))
    
    conns := make([]net.Conn, len(candidates))

    for i, candidate := range candidates {
        // use go routine to send the file to the candidate

        conn, err := net.Dial("tcp", candidate + ":" + global.HDFSPort)
        if err != nil {
            fmt.Println("Error connecting to server:", err)
            continue
        }
        defer conn.Close()
        conns[i] = conn

        wg.Add(1)
        go func(candidate string) {
            defer wg.Done()
            response := checkAppendFileToCandidate(conns[i], localfilename, HyDFSfilename)
            responses <- response
        }(candidate)
    }

    wg.Wait()
    close(responses)
    for response := range responses {
        if strings.HasPrefix(response, "Fail") {
            fmt.Println("server fail to creat files", response)
            return
        }
    }

    for i := 0; i < len(candidates); i++ {
        go appendFileToCandidate(conns[i], localfilename, HyDFSfilename)
    }

    cache.CacheMutex.Lock()
    cache.DeleteCacheEntry(HyDFSfilename)
    cache.CacheMutex.Unlock()
}


func createFileToCandidate(candidate string, localfilename string, HyDFSfilename string) {
    /*
    Connect and Send the local file to the candidate server to create the HyDFS file.
    */
    conn, err := net.Dial("tcp", candidate + ":" + global.HDFSPort)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return
    }
    defer conn.Close()

    // Send the "create" command with the HyDFS filename
    command := fmt.Sprintf("create %s", HyDFSfilename)
    _, err = conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return
    }


    // check the response from the server to check if server create the HyDFS file
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading from connection to check success or fail:", err)
        return
    }
    response := string(buffer[:n])
    if strings.HasPrefix(response, "Fail") {
        fmt.Println("Error creating file:", response)
        return
    }

    // Open the local file for reading
    localFile, err := os.Open(localfilename)
    if err != nil {
        fmt.Println("Error opening local file:", err)
        return
    }
    defer localFile.Close()

    // Read and send the file content
    for {
        n, err := localFile.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading from local file:", err)
            return
        }
        _, err = conn.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    }

    fmt.Printf("File %s sent to %s\n", localfilename, candidate)
}




func getFileFromReplica(VMaddress string, HyDFSfilename string, localfilename string) {
    /*
    Get the HyDFS file from the replica server and save it to the local machine.
    If the file exists, overwrite it.
    */
    // Connect to the server
    conn, err := net.Dial("tcp", VMaddress + ":" + global.HDFSPort)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return
    }
    defer conn.Close()

    // Send the "get" command with the HyDFS filename
    command := fmt.Sprintf("get %s", HyDFSfilename)
    _, err = conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return
    }


    // Check if the file exists
    if _, err := os.Stat(localfilename); err == nil {
        // File exists, overwrite it
        os.Remove(localfilename)
    }

    // Open the local file for writing
    localFile, err := os.Create(localfilename)
    if err != nil {
        fmt.Println("Error creating local file:", err)
        return
    }
    defer localFile.Close()

    // Retrieve and save the file content
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
        _, err = localFile.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to local file:", err)
            return
        }
    }

    fmt.Printf("File %s retrieved and saved as %s\n", HyDFSfilename, localfilename)
}


func checkAppendFileToCandidate(conn net.Conn, localfilename string, HyDFSfilename string) string{
    /*
    Check the server all reply "Success" to client
    If all sucess then start append
    */

    // Send the "append" command with the HyDFS filename
    command := fmt.Sprintf("append %s", HyDFSfilename)
    _, err := conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return "Fail"
    }

    // check the response from the server to check if server create the HyDFS file
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading from connection:", err)
        return "Fail"
    }

    response := string(buffer[:n])
    if strings.HasPrefix(response, "Fail") {
        fmt.Println("Error appending file:", response)
        return "Fail"
    }
    return "Success"
}


func appendFileToCandidate(conn net.Conn, localfilename string, HyDFSfilename string) {
    /*
    Connect and append the local file to the candidate server.
    */

    // Open the local file for reading
    localFile, err := os.Open(localfilename)
    if err != nil {
        fmt.Println("Error opening local file:", err)
        return
    }

    defer localFile.Close()

    // Read and send the file content
    buffer := make([]byte, 1024)
    for {
        n, err := localFile.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading from local file:", err)
            return
        }
        _, err = conn.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to connection:", err)
            return
        }
    }
}



func listMachine(HyDFSfilename string) {
    /*
    For Command "ls"
    list all the machine addresses that store the given HyDFS file.
    */
    global.Cluster = global.GetMembership()
    replicas := global.FindFileReplicas(HyDFSfilename)
    for _, replica := range replicas {
        fmt.Println(replica)
    }
}





func listFiles() {
    conn, err := net.Dial("tcp", "localhost:" + global.HDFSPort)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

    // Send the "ls" command
    _, err = conn.Write([]byte("ls"))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return
    }

    // Retrieve and print the response
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            fmt.Println("Error reading from connection:", err)
            return
        }
        if n == 0 {
            break
        }
        fmt.Print(string(buffer[:n]))
    }
}

func listMemberIds() {
    global.Cluster = global.GetMembership()
    for _, node := range global.Cluster {
        fmt.Println(node.Address, global.HashFunc(node.Address))
    }
}