package main

import (
	"strings"
	"os/exec"
	"sync"
	"testing"
	"fmt"
	"reflect"
)


func TestOneMahine( t *testing.T){
	// Test one machine and compare the grep's result

	// Test Frequent pattern
	test_args := []string {"-c", "GET", "../../data/test_vm1.log"}
	args := strings.Join(test_args, "\n")
	// get distributed result
	var wg sync.WaitGroup
	responses := make([]string, 10)
	wg.Add(1)
	ip := "fa24-cs425-6602.cs.illinois.edu"
	// ip := "localhost"
	args += "\x00"
	go connectAndSend(ip, 1, args, &wg, responses)
	wg.Wait()
	// get local result
    // Execute the grep command
    cmd := exec.Command("grep", test_args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error executing grep: %v\n", err)
    }

    
	local_responses := make([]string, 10)
    local_responses[1] = string(output)

	// Compare frequent pattern grep the result
	if reflect.DeepEqual(responses , local_responses) == false {
		t.Errorf("Remote Output %+v is not equal to Local Expected Output %+v", responses, local_responses)
	}

}


func TestOneMachine_infrequent( t *testing.T){
	// Test one machine and compare the grep's result

	// Test inFrequent pattern
	test_args := []string {"-H", "DELETE", "../../data/test_vm1.log"}
	args := strings.Join(test_args, "\n")
	// get distributed result
	var wg sync.WaitGroup
	responses := make([]string, 10)
	wg.Add(1)
	ip := "fa24-cs425-6602.cs.illinois.edu"
	// ip := "localhost"
	args += "\x00"
	go connectAndSend(ip, 1, args, &wg, responses)
	wg.Wait()
	// get local result
    // Execute the grep command
    cmd := exec.Command("grep", test_args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error executing grep: %v\n", err)
    }

    
	local_responses := make([]string, 10)
    local_responses[1] = string(output)

	// Compare frequent pattern grep the result
	if reflect.DeepEqual(responses , local_responses) == false {
		t.Errorf("Remote Output %+v is not equal to Local Expected Output %+v", responses, local_responses)
	}

}

func TestAllMachine( t *testing.T){
	/*
	Test all machines and compare the grep's result from machine 1
	*/
	
	responses := make([]string, 10)
	var wg sync.WaitGroup
	for i := 1; i < 10; i++ {
		// Use Sprintf to embed the integer into a string
		file_path := fmt.Sprintf("../../data/test_vm%d.log", i)
		test_args := []string {"-c", "GET", file_path}
		ip := fmt.Sprintf("fa24-cs425-66%02d.cs.illinois.edu", i + 1)

		args := strings.Join(test_args, "\n")
		// get distributed result
		wg.Add(1)
		// ip := "localhost"
		args += "\x00"
		go connectAndSend(ip, i, args, &wg, responses)
	}
	wg.Wait()

	local_responses := make([]string, 10)
	// get local result
	for i := 1; i < 10; i++ {
		file_path := fmt.Sprintf("../../data/test_vm%d.log", i)
		test_args := []string {"-c", "GET", file_path}
		cmd := exec.Command("grep", test_args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing grep: %v\n", err)
		}
	
		local_responses[i] = string(output)
	}

	// Compare the result
	if reflect.DeepEqual(responses , local_responses) == false {
		t.Errorf("Remote Output %+v is not equal to Local Expected Output %+v", responses, local_responses)
	}

}


func TestAllMachine_infrequent( t *testing.T){
	/*
	Test all machines and compare the grep's result from machine 1
	*/
	
	responses := make([]string, 10)
	var wg sync.WaitGroup
	for i := 1; i < 10; i++ {
		// Use Sprintf to embed the integer into a string
		file_path := fmt.Sprintf("../../data/test_vm%d.log", i)
		test_args := []string {"-H", "DELETE", file_path}
		ip := fmt.Sprintf("fa24-cs425-66%02d.cs.illinois.edu", i + 1)

		args := strings.Join(test_args, "\n")
		// get distributed result
		wg.Add(1)
		// ip := "localhost"
		args += "\x00"
		go connectAndSend(ip, i, args, &wg, responses)
	}
	wg.Wait()

	local_responses := make([]string, 10)
	// get local result
	for i := 1; i < 10; i++ {
		file_path := fmt.Sprintf("../../data/test_vm%d.log", i)
		test_args := []string {"-c", "GET", file_path}
		cmd := exec.Command("grep", test_args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing grep: %v\n", err)
		}
	
		local_responses[i] = string(output)
	}

	// Compare the result
	if reflect.DeepEqual(responses , local_responses) == false {
		t.Errorf("Remote Output %+v is not equal to Local Expected Output %+v", responses, local_responses)
	}

}
