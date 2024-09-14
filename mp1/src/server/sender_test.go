package main

import (
	//"bytes"
	//"net"
	"strings"
	"os/exec"
	"sync"
	"testing"
	"fmt"
	"reflect"
)


func TestOneMahine( t *testing.T){
	// Test one machine and compare the grep's result

	test_args := []string {"-c", "GET", "\n", "../../data/test_vm2.log"}
	args := strings.Join(test_args, "\n")
	// get distributed result
	var wg sync.WaitGroup
	responses := make([]string, 3)
	wg.Add(1)
	// ip := "fa24-cs425-6602.cs.illinois.edu"
	ip := "localhost"
	args += "\x00"
	go connectAndSend(ip, 2, args, &wg, responses)
	wg.Wait()

	fmt.Println(responses)
	
	// get local result
    // Execute the grep command
    cmd := exec.Command("grep", test_args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error executing grep: %v\n", err)
    }

    // Send the grep output back to the client
    local_responses := string(output) + "\x00"

	if reflect.DeepEqual(responses[1], local_responses) == false {
		t.Errorf("Remote Output %+v is not equal to Local Expected Output %+v", responses, local_responses)
	}

}
