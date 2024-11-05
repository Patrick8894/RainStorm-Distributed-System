package main

import (
    "flag"
    "fmt"
    "os"
    "mp3/src/global"
)

var cluster map[string]global.NodeInfo
// IMPORTANT: For HyDFS commands, pass arguments with with format "command filename" in TCP

func main() {
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
    // TODO: Implement the create functionality here
}

func getFile(HyDFSfilename string, localfilename string) {
    // TODO: Implement the get functionality here
}

func appendFile(localfilename string, HyDFSfilename string) {
    // TODO: Implement the append functionality here
}

func listMachine(HyDFSfilename string) {
    cluster = global.GetMembership()
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

func getFileFromReplica(VMaddress string, HyDFSfilename string, localfilename string) {
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
            break
        }
        _, err = localFile.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error writing to local file:", err)
            return
        }
    }

    fmt.Printf("File %s retrieved and saved as %s\n", HyDFSfilename, localfilename)
}

func listMemberIds() {
    cluster = global.GetMembership()
    for _, node := range cluster {
        fmt.Println(node.Address, global.HashFunc(node.Address))
    }
}