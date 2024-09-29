// send the command to each node
package main
import (
	"flag"
	"fmt"
)

func main (){
	// receive the command from the command line
	command := flag.String("command", "", "Command to send to each node")
	flag.Parse()

	// Check if the command is provided
	if *command == "" {
		fmt.Println("Please provide a command using --command flag")
		return
	}

	fmt.Println("Command received:", *command)

	// send the command to each node
	// wait for the response from each node
	// print the response from each node

}
