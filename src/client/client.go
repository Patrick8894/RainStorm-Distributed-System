package main

import (
	"fmt"
	// "os"
	"net"
	"os/exec"
	"strings"
)

func main() {

	ln, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		fmt.Println("Error in connection")
		return
	}

	defer ln.Close()

	fmt.Println("Client is listening")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in connection")
			return
		}
		go handleConnection(conn)
		// grep_function("log", "../../data/vm1.log")
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// read the pattern from the server
	pattern := make([]byte, 1024)
	_, err := conn.Read(pattern)

	if err != nil {
		fmt.Println("Error in reading the pattern")
		return
	}

	fmt.Println("Pattern received from client:", string(pattern))

	// send the response to the client
	grep_result := grep_function(string(pattern), "../../data/vm1.log")

	conn.Write([]byte(grep_result))
}

func grep_function(pattern string, file_path string) string {

	cmd := exec.Command("grep", "-cH", pattern, file_path)

	fmt.Println("Command to execute:", cmd)

	var out strings.Builder

	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		fmt.Printf("Grep Command execution error %q", err)
	}

	fmt.Printf("%q\n", out.String())
	return out.String()
}
