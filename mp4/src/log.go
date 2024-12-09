package main

import (
    "fmt"
    "os"
    "os/exec"
)

func main() {
    if len(os.Args) != 8 {
        fmt.Println("Usage: <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>")
        return
    }

    op1Exe := os.Args[1]
    op2Exe := os.Args[2]
    hydfsSrcFile := os.Args[3]
    hydfsDestFilename := os.Args[4]
    numTasks := os.Args[5]
    X := os.Args[6]
    stateful := os.Args[7]

    fmt.Printf("op1_exe: %s\n", op1Exe)
    fmt.Printf("op2_exe: %s\n", op2Exe)
    fmt.Printf("hydfs_src_file: %s\n", hydfsSrcFile)
    fmt.Printf("hydfs_dest_filename: %s\n", hydfsDestFilename)
    fmt.Printf("num_tasks: %s\n", numTasks)
    fmt.Printf("X: %s\n", X)
    fmt.Printf("stateful: %s\n", stateful)

	localfilename := "/localfile"
	outputfilename := "/outputfile"

	cmd := exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", localfilename, "--HyDFSfilename", hydfsSrcFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing get: %v\n", err)
		fmt.Printf("Command output: %s\n", string(output))
		return
	}

	// read the local file
	file, err := os.Open(localfilename)
	if err != nil {
		fmt.Printf("Error opening local file: %v\n", err)
		return
	}

	outputfile, err := os.Create(outputfilename)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}

	state := make(map[string]int)

	// read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		
		cmd = exec.Command("../ops/" + op1Exe, X)
		cmd.Stdin = strings.NewReader(request)
		output, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error running external program: %v\n", err)
			continue
		}

		if strings.TrimSpace(string(output)) != "1" {
			continue
		}

		cmd = exec.Command("../ops/" + op2Exe)
		cmd.Stdin = strings.NewReader(line)
		output, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error running external program: %v\n", err)
			continue
		}

		if stateful == "stateful" {
			state[output]++
		} else {
			outputfile.WriteString(output + "\n")
		}
	}

	if stateful == "stateful" {
		for key, value := range state {
			outputfile.WriteString(fmt.Sprintf("%s: %d\n", key, value))
		}
	}
}